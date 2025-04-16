package cmd

import (
	"fmt"
	"log"

	"github.com/The-EpaG/trash-cli/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// purgeCmd represents the purge command
var PurgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Empty the trash",
	Long:  `Empty all files from the trash.`,
	Run: func(cmd *cobra.Command, args []string) {
		trashHomeDir := viper.GetString("trash_home_dir")
		if err := internal.PurgeTrash(trashHomeDir, trashHomeDir); err != nil {
			log.Fatalf("Error purging trash: %v", err)
		}
		fmt.Println("Trash purged successfully")
	},
}

func init() {
}