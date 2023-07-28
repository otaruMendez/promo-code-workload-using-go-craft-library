package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"example.com/main/internal/database"
	"example.com/main/internal/migration"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewCmdAPI() *cobra.Command {
	return &cobra.Command{
		Use:   "promo-api",
		Short: "Promo API",
		RunE: func(_ *cobra.Command, _ []string) error {
			r := gin.Default()

			err := godotenv.Load()
			if err != nil {
				log.Fatal("Error loading .env file")
			}

			err = envconfig.Process("", &cfg)
			if err != nil {
				log.Fatal(err.Error())
			}

			sqlxConn, err := database.Connect(cfg.Database)
			if err != nil {
				return fmt.Errorf("could not connect to database: %w", err)
			}

			promoRepo = *NewRepo(sqlxConn)

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

			r.GET("/promotions/:promotionsId", func(c *gin.Context) {
				promotionsId, err := strconv.Atoi(c.Param("promotionsId"))
				if err != nil {
					fmt.Printf("could not extract promoId: %s", err.Error())
					c.String(http.StatusBadRequest, "could not extract promotions Id")
					return
				}

				promo, err := promoRepo.GetPromo(context.TODO(), promotionsId)
				if err != nil {
					if err == database.ErrRecordNotFound {
						c.String(http.StatusNotFound, "promo not found")
						return
					}

					fmt.Printf("could not get promo: %s", err.Error())
					c.String(http.StatusInternalServerError, "could not get promo")
					return
				}

				c.JSON(http.StatusOK, promo)
			})

			r.Run(":8080")

			return nil
		},
	}
}
