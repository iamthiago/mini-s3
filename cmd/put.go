package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// putCmd represents the put command
var putCmd = &cobra.Command{
	Use:   "put",
	Short: "Add objects to a bucket",
	Long: `Add objects to the specified bucket.

Example usage:
  mini-s3 put <bucket-name> <object-name>`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			fmt.Println("Usage: mini-s3 put <bucket-name> <object-name>")
			return
		}

		bucket := args[0]
		object := args[1]

		file, err := os.Open(object)
		if err != nil {
			fmt.Printf("Failed to open file: %v\n", err)
			return
		}
		defer file.Close()

		objectName := filepath.Base(object)

		_, err = storageInstance.Save(bucket, objectName, file)
		if err != nil {
			fmt.Printf("Failed to save file: %v\n", err)
			return
		}
		fmt.Printf("Successfully added %s to bucket %s\n", objectName, bucket)
	},
}

func init() {
	rootCmd.AddCommand(putCmd)
}
