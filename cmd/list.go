package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/The-EpaG/trash-cli/internal"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var ListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List files in the trash",
	Long:    `List all files currently in the trash.`,
	Run: func(cmd *cobra.Command, args []string) {
		_, trashInfoDir, err := internal.GetHomeTrashPaths()
		if err != nil {
			log.Fatalf("Critical error: %v", err)
		} else if trashInfoDir == "" {
			log.Fatalf("Critical error: %v", err)
		}
		showDetails, err := cmd.Flags().GetBool("details")
		if err != nil {
			log.Fatalf("cannot get details flag: %v", err)
		}

		err = listTrash(trashInfoDir, showDetails)
		if err != nil {
			log.Fatalf("Error listing trash: %v", err)
		}
	},
}

func init() {
	ListCmd.Flags().BoolP("details", "l", false, "long listing format")
}

// listTrash lists the files in the trash.
func listTrash(trashInfoDir string, showDetails bool) error {
	entries, err := os.ReadDir(trashInfoDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return fmt.Errorf("cannot read trash info directory %q: %w", trashInfoDir, err)
	}

	if len(entries) == 0 {
		return nil
	}

	for _, entry := range entries {
		infoFilePath := filepath.Join(trashInfoDir, entry.Name())

		trashInfo, err := internal.ParseTrashInfo(infoFilePath)
		if err != nil {
			// Skip to the next file
			continue
		}

		if showDetails {
			log.Printf("%s -> %s\n", filepath.Base(trashInfo.Path), trashInfo.TrashFile)
			log.Printf("    Path: %s\n", trashInfo.Path)
			log.Printf("    Deletion Date: %s\n", trashInfo.DeletionDate)
		} else {
			log.Printf("%s -> %s\n", filepath.Base(trashInfo.Path), trashInfo.TrashFile)
		}
	}

	return nil
}
