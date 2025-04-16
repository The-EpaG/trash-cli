package cmd

import (
	"log"
	"os"
	"path/filepath"

	"github.com/The-EpaG/trash-cli/internal"
	"github.com/spf13/cobra"
)

// restoreCmd represents the restore command
var RestoreCmd = &cobra.Command{
	Use:   "restore [file]",
	Short: "Restore a file from the trash",
	Long:  `Restore a file from the trash.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fileName := args[0]
		tempTrashFilesDir, tempTrashInfoDir, err := internal.GetTrashPaths(fileName)
		if err != nil {
			log.Fatalf("Error: cannot retrieve trash path for '%s': %v", fileName, err)
		}

		if _, err := os.Stat(filepath.Join(tempTrashFilesDir, fileName)); err != nil {
			if os.IsNotExist(err) {
				log.Printf("Error: file '%s' does not exist in trash", fileName)
			} else {
				log.Printf("Error: cannot check if file '%s' exists in trash", fileName)
			}
			os.Exit(1)
		}

		if err := internal.RestoreFromTrash(tempTrashFilesDir, tempTrashInfoDir, fileName); err != nil {
			log.Fatalf("Error restoring '%s': %v", fileName, err)
		}
	},
}

func init() {
}
