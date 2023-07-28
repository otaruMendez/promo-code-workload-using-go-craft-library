package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"example.com/main/internal/database"
	"example.com/main/internal/migration"
	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type InsertPromoWorkerContext struct {
	context context.Context
}

type ScheduleJobsWorkerContext struct{}

var promoRepo Repo

var schedulerID int

var cfg Config

var schedulerRedisPool *redis.Pool

var workersRedisPool *redis.Pool

func getRedisPool(redisHost string, maxActive int, maxIdle int) *redis.Pool {
	return &redis.Pool{
		MaxActive: maxActive,
		MaxIdle:   maxIdle,
		Wait:      true,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", fmt.Sprintf("%s:6379", redisHost))
		},
	}
}

func NewCmdScheduler() *cobra.Command {
	return &cobra.Command{
		Use:   "promo-scheduler",
		Short: "Promo Scheduler",
		RunE: func(_ *cobra.Command, _ []string) error {
			err := godotenv.Load()
			if err != nil {
				log.Fatal("Error loading .env file")
			}

			err = envconfig.Process("", &cfg)
			if err != nil {
				return fmt.Errorf("could not load env config: %w", err)
			}

			logger, err := zap.NewProductionConfig().Build()
			if err != nil {
				return err
			}

			err = migration.PerformUp(
				logger.With(zap.String("module", "migration-up")),
				"migrations/",
				cfg.Database.DSN,
			)
			if err != nil {
				log.Fatal("error performing migrations",
					zap.Error(err),
				)
			}

			schedulerRedisPool = getRedisPool(cfg.RedisHost, 1, 1)

			pool := work.NewWorkerPool(ScheduleJobsWorkerContext{}, 1, "vervegroup", schedulerRedisPool)

			// schedule job every 30 mins
			pool.PeriodicallyEnqueue("0 0,30 * * * *", "schedule_jobs")
			pool.JobWithOptions("schedule_jobs", work.JobOptions{MaxFails: 1}, ScheduleJobs)

			pool.Start()

			// Wait for a signal to quit:
			signalChan := make(chan os.Signal, 1)
			signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
			<-signalChan

			// Stop the pool
			pool.Stop()

			return nil
		},
	}
}

func batchPromoRecordsFromCSV(file io.Reader, start int, end int) (promoList []Promo) {
	fileScanner := bufio.NewScanner(file)

	if start < 1 {
		start = 1
	}

	filePos := 1
	for fileScanner.Scan() {
		if filePos < start {
			filePos++
			continue
		}

		csvRecord := fileScanner.Text()
		recordItems := strings.Split(csvRecord, ",")

		promoList = append(promoList, Promo{
			ID:    recordItems[0],
			Price: recordItems[1],
			Date:  recordItems[2],
		})

		if filePos >= end {
			break
		}

		filePos++
	}

	return promoList
}

func NewCmdWorkers() *cobra.Command {
	return &cobra.Command{
		Use:   "promo-workers",
		Short: "Promo Workers",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			err := godotenv.Load()
			if err != nil {
				log.Fatal("Error loading .env file")
			}

			err = envconfig.Process("", &cfg)
			if err != nil {
				return fmt.Errorf("could not load env config: %w", err)
			}

			logger, err := zap.NewProductionConfig().Build()
			if err != nil {
				return err
			}

			err = migration.PerformUp(
				logger.With(zap.String("module", "migration-up")),
				"migrations/",
				cfg.Database.DSN,
			)
			if err != nil {
				log.Fatal("error performing migrations",
					zap.Error(err),
				)
			}

			sqlxConn, err := database.Connect(cfg.Database)
			if err != nil {
				return fmt.Errorf("could not connect to database: %w", err)
			}

			promoRepo = *NewRepo(sqlxConn)

			workersRedisPool = getRedisPool(cfg.RedisHost, 5, 5)

			pool := work.NewWorkerPool(InsertPromoWorkerContext{context: ctx}, 10, "vervegroup", workersRedisPool)

			pool.Job("insert_promo", (*InsertPromoWorkerContext).InsertPromo)

			pool.Start()

			// Wait for a signal to quit:
			signalChan := make(chan os.Signal, 1)
			signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
			<-signalChan

			// Stop the pool
			pool.Stop()

			return nil
		},
	}
}

func ScheduleJobs(_ *work.Job) error {
	schedulerID++

	log.Println("running scheduler with ID: ", schedulerID)

	file, err := os.Open(cfg.PromoCSVPath)
	if err != nil {
		return fmt.Errorf("can't open promo csv file: %w", err)
	}

	noOfRecords, err := lineCounter(file)
	if err != nil {
		return fmt.Errorf("could not read lines of file: %w", err)
	}

	enqueuer := work.NewEnqueuer("vervegroup", schedulerRedisPool)

	noOfJobs := int(math.Ceil(float64(noOfRecords) / float64(cfg.BatchSize)))
	for i := 0; i < noOfJobs; i++ {
		jobArgs := work.Q{"start": i + 1, "end": 1 + cfg.BatchSize, "schedulerId": schedulerID}
		if _, err := enqueuer.EnqueueUnique("insert_promo", jobArgs); err != nil {
			fmt.Printf("could not enqueue job: %s", err.Error())
		}
	}

	return nil
}

func (c *InsertPromoWorkerContext) InsertPromo(job *work.Job) error {
	log.Println("runing insert promo worker for job: ", job.Name)

	file, err := os.Open(cfg.PromoCSVPath)
	if err != nil {
		return fmt.Errorf("can't open promo csv file: %w", err)
	}

	promoRecords := batchPromoRecordsFromCSV(
		bufio.NewReader(file),
		int(job.ArgInt64("start")),
		int(job.ArgInt64("end")),
	)
	for i, promo := range promoRecords {
		filePos := int(job.ArgInt64("start")) + i
		version := int(job.ArgInt64("schedulerId"))
		if err := promoRepo.UpsertPromo(c.context, version, filePos, promo); err != nil {
			log.Println("could not insert promo", err, zap.Any("promo", promo))
		}
	}

	return nil
}

// https://stackoverflow.com/a/24563853/2893446
func lineCounter(r io.Reader) (int, error) {
	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, err
		}
	}
}
