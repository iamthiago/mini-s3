package cmd

import (
	"os"
	"path/filepath"

	"github.com/iamthiago/mini-s3/internal/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var (
	storageInstance storage.Storage
	cfgFile         string
	dataDir         string
)

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mini-s3.yaml)")
	rootCmd.PersistentFlags().StringVar(&dataDir, "data-dir", "", "path to data directory")

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	cobra.OnInitialize(initConfig, initStorage)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err == nil {
			configPath := filepath.Join(home, ".mini-s3.yaml")
			viper.AddConfigPath(home)
			viper.SetConfigType("yaml")
			viper.SetConfigName(".mini-s3")

			// Create the default config if it doesn't exist
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				createDefaultConfig(configPath)
			}
		}
	}

	viper.AutomaticEnv()
	_ = viper.ReadInConfig()
}

func createDefaultConfig(path string) {
	defaultConfig := `# mini-s3 configuration
data-dir: ./data
`
	err := os.WriteFile(path, []byte(defaultConfig), 0644)
	if err != nil {
		return
	}
}

func initStorage() {
	// Priority: CLI flag > config file > default
	rootDir := dataDir
	if rootDir == "" {
		rootDir = viper.GetString("data-dir")
	}
	if rootDir == "" {
		rootDir = "./data" // default
	}

	storageInstance = storage.NewLocalStorage(rootDir, storage.NewValueChecksum())
}
