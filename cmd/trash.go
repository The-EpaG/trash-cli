package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/The-EpaG/trash-cli/internal"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// trashCmd represents the trash command
var TrashCmd = &cobra.Command{
	Use:   "trash [file...]",
	Aliases: []string{"rm", "remove", "delete", "del"},
	Short: "Trash a file or directory",
	Long:  `Move specified files or directories to the trash.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		operationSuccessful := true
		for _, file := range args {
			var tempTrashFilesDir, tempTrashInfoDir string
			tempTrashFilesDir, tempTrashInfoDir, err := internal.GetTrashPaths(file)
			if err != nil {
				log.Printf("Error: cannot retrieve trash path for '%s': %v", file, err)
			}
			if err = moveToTrash(file, tempTrashFilesDir, tempTrashInfoDir); err != nil {
				log.Printf("Error moving '%s': %v", file, err)
				operationSuccessful = false
			}
		}
		if !operationSuccessful {
			os.Exit(1)
		}
	},
}

// moveToTrash moves a file or directory to the trash directory and creates the corresponding .trashinfo file.
func moveToTrash(srcPath, trashFilesDir, trashInfoDir string) error {
	// Get the original absolute path before moving the file.
	originalAbsPath, err := filepath.Abs(srcPath)
	if err != nil {
		return fmt.Errorf("cannot get absolute path for %s: %w", srcPath, err)
	}

	// Check if the source file exists and is accessible.
	_, err = os.Stat(srcPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file %s does not exist", srcPath)
		}

		return fmt.Errorf("error checking file %s: %w", srcPath, err)
	}

	// Generate a unique ID for the file in the trash
	uniqueID := uuid.New().String()

	// Construct the final destination path using the unique ID
	destPath := filepath.Join(trashFilesDir, uniqueID)

	// Create the .trashinfo file before moving the file
	if err = createTrashInfo(originalAbsPath, trashInfoDir, uniqueID); err != nil {
		return fmt.Errorf("cannot create .trashinfo file for %s: %w", srcPath, err)
	}

	// Move the file/directory to the trash.
	if err = os.Rename(srcPath, destPath); err != nil {
		return fmt.Errorf("cannot move %s to trash: %w", srcPath, err)
	}

	return nil
}

// createTrashInfo creates a .trashinfo file with metadata about a trashed file, as per the Trash Specification.
func createTrashInfo(originalPath, infoDir, fileName string) error {
	infoFileTempPath := filepath.Join(infoDir, fileName+internal.TrashInfoExt+".tmp")
	infoFilePath := filepath.Join(infoDir, fileName+internal.TrashInfoExt)

	defer os.Remove(infoFileTempPath) // Ensure temporary file is cleaned up

	infoFile, err := os.OpenFile(infoFileTempPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, internal.InfoFilePermissions)
	if err != nil {
		return fmt.Errorf("cannot create info file '%s': %w", infoFileTempPath, err)
	}
	defer infoFile.Close()

	// Get the current date and time in ISO 8601 format, as required by the specification.
	deletionDate := time.Now().Format(time.RFC3339)

	// Prepare the content of the .trashinfo file.
	infoContent := fmt.Sprintf("%s\nPath=%s\nDeletionDate=%s\n", internal.TrashInfoSection, originalPath, deletionDate)

	_, err = infoFile.WriteString(infoContent)
	if err != nil {
		return fmt.Errorf("cannot write to info file '%s': %w", infoFileTempPath, err)
	}

	if err = os.Rename(infoFileTempPath, infoFilePath); err != nil {
		return fmt.Errorf("cannot rename temp info file '%s' to '%s': %w", infoFileTempPath, infoFilePath, err)
	}

	return nil
}
