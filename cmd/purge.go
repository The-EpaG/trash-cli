package cmd

import (
	"fmt"
	"log"
	"os"

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
		if err := emptyTrash(trashHomeDir, trashHomeDir); err != nil {
			log.Fatalf("Error purging trash: %v", err)
		}
		fmt.Println("Trash purged successfully")
	},
}

func init() {
}

// emptyTrash empties the trash directories.
func emptyTrash(filesDir, infoDir string) error {
	if err := os.RemoveAll(filesDir); err != nil {
		return fmt.Errorf("cannot empty trash files: %w", err)
	}
	return os.RemoveAll(infoDir)
}
