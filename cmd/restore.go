package cmd

import (
	"fmt"
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

		if err := restoreFromTrash(tempTrashFilesDir, tempTrashInfoDir, fileName); err != nil {
			log.Fatalf("Error restoring '%s': %v", fileName, err)
		}
	},
}

func init() {
}

// restoreFromTrash restores a file from the trash.
func restoreFromTrash(trashDir, infoDir, filename string) error {
	infoFilepath := filepath.Join(infoDir, filename+internal.TrashInfoExt)
	trashFilepath := filepath.Join(trashDir, filename)

	trashInfo, err := internal.ParseTrashInfo(infoFilepath)
	if err != nil {
		return fmt.Errorf("cannot parse .trashinfo file: %w", err)
	}

	// Move the file back to its original location
	if err := os.Rename(trashFilepath, trashInfo.Path); err != nil {
		return fmt.Errorf("cannot move file back to original location: %w", err)
	}

	// Remove the .trashinfo file
	if err := os.Remove(infoFilepath); err != nil {
		return fmt.Errorf("cannot remove .trashinfo file: %w", err)
	}

	return nil
}
