package main

import (
	"context"
	"log"

	"github.com/spf13/cobra"
)

func main() {
	ctx := context.Background()

	rootCmd := &cobra.Command{
		Use: "vervegroup [command]",
	}

	rootCmd.AddCommand(NewCmdScheduler())
	rootCmd.AddCommand(NewCmdWorkers())
	rootCmd.AddCommand(NewCmdAPI())

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		log.Fatal(err)
	}
}
