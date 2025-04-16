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
	Use:   "list",
	Aliases: []string{"ls"},
	Short: "List files in the trash",
	Long:  `List all files currently in the trash.`,
	Run: func(cmd *cobra.Command, args []string) {
		trashFilesDir, trashInfoDir, err := internal.GetHomeTrashPaths()
		if err != nil {
			log.Fatalf("Critical error: %v", err)
		} else if trashFilesDir == "" || trashInfoDir == "" {
			log.Fatalf("Critical error: %v", err)
		}
		showDetails, err := cmd.Flags().GetBool("details")
		if err != nil {
			log.Fatalf("cannot get details flag: %v", err)
		}

		err = listTrash(trashFilesDir, trashInfoDir, showDetails)
		if err != nil {
			log.Fatalf("Error listing trash: %v", err)
		}
	},
}

func init() {
	ListCmd.Flags().BoolP("details", "l", false, "Show details of the trashed files")
}

// listTrash lists the files in the trash.
func listTrash(trashFilesDir, trashInfoDir string, showDetails bool) error {
	files, err := os.ReadDir(trashInfoDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("The trash is empty.")
			return nil
		}
		return fmt.Errorf("cannot read trash info directory '%s': %w", trashInfoDir, err)
	}

	if len(files) == 0 {
		fmt.Println("The trash is empty.")
		return nil
	}

	for _, file := range files {
		infoFilePath := filepath.Join(trashInfoDir, file.Name())

		trashInfo, err := internal.ParseTrashInfo(infoFilePath)
		if err != nil {
			log.Printf("Error parsing .trashinfo file for '%s': %v", file.Name(), err)
			continue // Skip to the next file
		}

		if showDetails {
			fmt.Printf("%s -> %s\n", filepath.Base(trashInfo.Path), trashInfo.TrashFile)
			fmt.Printf("    Path: %s\n", trashInfo.Path)
			fmt.Printf("    Deletion Date: %s\n", trashInfo.DeletionDate)
		} else {
			fmt.Printf("%s -> %s\n", filepath.Base(trashInfo.Path), trashInfo.TrashFile)
		}

	}
	return nil
}