package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/The-EpaG/trash-cli/internal"
)

// trashCmd represents the trash command
var TrashCmd = &cobra.Command{
	Use:   "trash [file...]",
	Aliases: []string{"rm", "remove", "delete", "del"},
	Short: "Move specified files to trash",
	Long:  `Move specified files to the trash.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		operationSuccessful := true
		for _, file := range args {
			var tempTrashFilesDir, tempTrashInfoDir string
			tempTrashFilesDir, tempTrashInfoDir, err := internal.GetTrashPaths(file)
			if err != nil {
				log.Printf("Error: cannot retrieve trash path for '%s': %v", file, err)
			}
			if err = internal.MoveToTrash(file, tempTrashFilesDir, tempTrashInfoDir); err != nil {
				log.Printf("Error moving '%s': %v", file, err)
				operationSuccessful = false
			}
		}
		if !operationSuccessful {
			os.Exit(1)
		}
	},
}