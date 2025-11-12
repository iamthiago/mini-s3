package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get an object from a bucket and save it to a local file",
	Long: `Get an object from a bucket and save it to a local file.

Example usage:
  mini-s3 get <bucket-name> <object-name> <output-dir>`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 3 {
			fmt.Println("Usage: mini-s3 get <bucket-name> <object-name> <output-dir>")
			return
		}

		bucket := args[0]
		object := args[1]
		outDir := args[2]

		fromBucket, objInfo, err := storageInstance.Get(bucket, object)
		if err != nil {
			fmt.Printf("Error getting object: %v\n", err)
			return
		}
		defer fromBucket.Close()

		path := filepath.Join(outDir, objInfo.Object)
		outFile, err := os.Create(path)
		if err != nil {
			fmt.Printf("Error creating fromBucket: %v\n", err)
		}
		defer outFile.Close()

		_, err = io.Copy(outFile, fromBucket)
		if err != nil {
			fmt.Printf("Error writing to file: %v\n", err)
		}
		fmt.Printf("Successfully saved %s to %s\n", objInfo.Object, path)
	},
}

func init() {
	rootCmd.AddCommand(getCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
