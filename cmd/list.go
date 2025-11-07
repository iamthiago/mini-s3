package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List objects in a bucket",
	Long: `List all objects in the specified bucket.

This command retrieves and displays a list of objects stored in a bucket.

Example usage:
  mini-s3 list <bucket-name>`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("Usage: mini-s3 list <bucket-name>")
			return
		}

		bucket := args[0]

		objects, err := storageInstance.ListObjects(bucket)
		if err != nil {
			fmt.Printf("Failed to list objects: %v\n", err)
			return
		}

		if len(objects) == 0 {
			fmt.Println("No objects found")
			return
		}

		fmt.Printf("%-25s %-10s %s\n", "CREATED", "SIZE", "NAME")
		fmt.Println("-----------------------------------------------------------")
		for _, obj := range objects {
			timestamp := obj.CreatedAt.Format("2006-01-02 15:04:05")
			size := formatSize(obj.Size)
			fmt.Printf("%-25s %-10s %s\n", timestamp, size, obj.Object)
		}
	},
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func init() {
	rootCmd.AddCommand(listCmd)
}
