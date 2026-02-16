package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "lifesupport-backend",
	Short: "Life Support Backend Service",
	Long:  `Life Support Backend Service provides HTTP API and Temporal worker functionality.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	ctx := context.Background()	
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Global flags can be added here
}
