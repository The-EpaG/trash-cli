package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/The-EpaG/trash-cli/internal"
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

		err = internal.ListTrash(trashFilesDir, trashInfoDir, showDetails)
		if err != nil {
			log.Fatalf("Error listing trash: %v", err)
		}
	},
}

func init() {
	ListCmd.Flags().BoolP("details", "l", false, "Show details of the trashed files")
}