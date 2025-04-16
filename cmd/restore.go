package cmd

import (
	"log"
	"os"
	"path/filepath"

	"github.com/The-EpaG/trash-cli/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// restoreCmd represents the restore command
var RestoreCmd = &cobra.Command{
	Use:   "restore [file]",
	Short: "Restore a file from the trash",
	Long:  `Restore a file from the trash.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		trashHomeDir := viper.GetString("trash_home_dir")
		fileName := args[0]

		if _, err := os.Stat(filepath.Join(trashHomeDir, fileName)); err != nil {
			if os.IsNotExist(err) {
				log.Printf("Error: file '%s' does not exist in trash", fileName)
			} else {
				log.Printf("Error: cannot check if file '%s' exists in trash", fileName)
			}
			os.Exit(1)
		}

		if err := internal.RestoreFromTrash(trashHomeDir, trashHomeDir, fileName); err != nil {
			log.Fatalf("Error restoring '%s': %v", fileName, err)
		}
	},
}

func init() {
}
