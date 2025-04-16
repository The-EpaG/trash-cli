package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/The-EpaG/trash-cli/cmd"
)

// Constants for better readability and maintainability, as per the Trash Specification.
const (
	trashDirName       = "Trash"
	defaultPermissions = 0700 // Owner-only rwx
	localShareDirName  = ".local/share"
)

// ensureDir ensures a directory exists, creating it if necessary.
func ensureDir(dirPath string) (err error) {
	err = os.MkdirAll(dirPath, defaultPermissions)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("cannot create directory '%s': %w", dirPath, err)
	}
	return
}

var rootCmd = &cobra.Command{
	Use:   "trash-cli",
	Short: "A command-line utility for managing the trash",
	Long:  `trash-cli is a command-line utility that allows you to manage files and directories in the trash, based on the XDG Trash specification.`,
}

// init executes at application start
func init() {
	// Configuration setup
	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.config/github.com/The-EpaG/trash-cli")
	viper.SetDefault("trash_files_dir", "files")
	viper.SetDefault("trash_info_dir", "info")
	//trash home
	usr, err := user.Current()
	if err != nil {
		log.Fatalf("cannot get current user info: %v", err)
	}
	viper.SetDefault("trash_home_dir", filepath.Join(usr.HomeDir, localShareDirName, trashDirName))
	//
	if err := viper.ReadInConfig(); err != nil {
		if !errors.As(err, &viper.ConfigFileNotFoundError{}) {
			log.Fatalf("Error reading config file: %v", err)
		}
	}

	// cobra commands
	rootCmd.AddCommand(cmd.ListCmd)
	rootCmd.AddCommand(cmd.TrashCmd)
	rootCmd.AddCommand(cmd.RestoreCmd)
	rootCmd.AddCommand(cmd.PurgeCmd)
}

func main() {
	log.SetFlags(0) // Removes timestamp/prefix from logs.

	// Ensure that the 'files' and 'info' directories exist in user path.
	trashHomeDir := viper.GetString("trash_home_dir")
	trashFilesDir := filepath.Join(trashHomeDir, viper.GetString("trash_files_dir"))
	trashInfoDir := filepath.Join(trashHomeDir, viper.GetString("trash_info_dir"))
	if err := ensureDir(trashFilesDir); err != nil {
		log.Fatalf("Critical error: %v", err)
	} else if err := ensureDir(trashInfoDir); err != nil {
		log.Fatalf("Critical error: %v", err)
	}
	// Cobra
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Critical error: %v", err)
	}
}
