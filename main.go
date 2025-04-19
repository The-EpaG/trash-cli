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

var rootCmd = &cobra.Command{
	Use:   "trash",
	Short: "A command-line utility for managing the trash",
	Long:  `trash is a command-line utility that allows you to manage files and directories in the trash, based on the XDG Trash specification.`,
}

// init executes at application start
func init() {
	log.SetFlags(0)

	// Load configuration
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.config/trash-cli")
	viper.SetDefault("trash.filesDir", "files")
	viper.SetDefault("trash.infoDir", "info")

	// Set default trash home directory
	usr, err := user.Current()
	if err != nil {
		log.Fatalf("cannot get current user info: %v", err)
	}
	defaultTrashHomeDir := filepath.Join(usr.HomeDir, localShareDirName, trashDirName)
	viper.SetDefault("trash.homeDir", defaultTrashHomeDir)

	// Read configuration file
	if err := viper.ReadInConfig(); err != nil {
		if !errors.As(err, &viper.ConfigFileNotFoundError{}) {
			log.Fatalf("Error reading config file: %v", err)
		}
	}

	// Register Cobra commands
	rootCmd.AddCommand(cmd.ListCmd)
	rootCmd.AddCommand(cmd.RemoveCmd)
	rootCmd.AddCommand(cmd.RestoreCmd)
	rootCmd.AddCommand(cmd.PurgeCmd)
}

func main() {
	if err := ensureDirs(); err != nil {
		log.Fatalf("Critical error: %v", err)
	}

	if err := rootCmd.Execute(); err != nil {
		// log.Fatalf("Critical error: %v", err)
	}
}

func ensureDirs() error {
	homeDir := viper.GetString("trash.homeDir")
	filesDir := filepath.Join(homeDir, viper.GetString("trash.filesDir"))
	infoDir := filepath.Join(homeDir, viper.GetString("trash.infoDir"))

	if err := os.MkdirAll(filesDir, defaultPermissions); err != nil {
		return fmt.Errorf("error creating files directory: %v", err)
	}

	if err := os.MkdirAll(infoDir, defaultPermissions); err != nil {
		return fmt.Errorf("error creating info directory: %v", err)
	}

	return nil
}
