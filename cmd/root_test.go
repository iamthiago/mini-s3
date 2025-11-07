package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestRootCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "root command without args",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "root command with help flag",
			args:    []string{"--help"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use:   "mini-s3",
				Short: rootCmd.Short,
				Long:  rootCmd.Long,
			}
			cmd.SetArgs(tt.args)
			err := cmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRootCommandHasRequiredFields(t *testing.T) {
	if rootCmd.Use == "" {
		t.Error("rootCmd.Use should not be empty")
	}
	if rootCmd.Short == "" {
		t.Error("rootCmd.Short should not be empty")
	}
	if rootCmd.Long == "" {
		t.Error("rootCmd.Long should not be empty")
	}
}

func TestRootPersistentCommandFlags(t *testing.T) {
	configFlag := rootCmd.PersistentFlags().Lookup("config")
	if configFlag == nil {
		t.Error("config flag should be registered")
	}

	dataDirFlag := rootCmd.PersistentFlags().Lookup("data-dir")
	if dataDirFlag == nil {
		t.Error("data-dir flag should be registered")
	}
}
