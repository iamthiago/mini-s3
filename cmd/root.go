package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mini-s3",
	Short: "A minimal S3-compatible object storage server",
	Long: `mini-s3 is a lightweight, self-hosted object storage server compatible with the Amazon S3 API.

It allows you to store, retrieve, and manage objects using familiar S3 commands and tools.

Features:
- S3-compatible API for easy integration with existing tools
- Local file system backend for object storage
- Simple configuration and deployment

Example usage:
  mini-s3 --data-dir /path/to/data

You can use AWS CLI or SDKs to interact with mini-s3 as you would with Amazon S3.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mini-s3.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
