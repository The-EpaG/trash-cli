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
	Use:     "trash [file...]",
	Aliases: []string{"rm", "remove", "delete", "del"},
	Short:   "Trash a file or directory",
	Long:    `Move specified files or directories to the trash.`,
	Args:    cobra.MinimumNArgs(1),
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
func moveToTrash(sourcePath, trashFilesDir, trashInfoDir string) error {
	// Get the absolute path of the source file.
	absolutePath, err := filepath.Abs(sourcePath)
	if err != nil {
		return fmt.Errorf("cannot get absolute path for %s: %w", sourcePath, err)
	}

	// Verify that the source file exists and is accessible.
	if _, err = os.Stat(sourcePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file %s does not exist", sourcePath)
		}
		return fmt.Errorf("error accessing file %s: %w", sourcePath, err)
	}

	// Generate a unique identifier for the trashed file.
	uniqueIdentifier := uuid.New().String()

	// Construct the destination path in the trash using the unique identifier.
	destinationPath := filepath.Join(trashFilesDir, uniqueIdentifier)

	// Create the .trashinfo file for the source file.
	if err = createTrashInfo(absolutePath, trashInfoDir, uniqueIdentifier); err != nil {
		return fmt.Errorf("cannot create .trashinfo file for %s: %w", sourcePath, err)
	}

	// Move the source file or directory to the trash.
	if err = os.Rename(sourcePath, destinationPath); err != nil {
		return fmt.Errorf("cannot move %s to trash: %w", sourcePath, err)
	}

	return nil
}

// createTrashInfo creates a .trashinfo file with metadata about a trashed file, as per the Trash Specification.
func createTrashInfo(originalPath, infoDir, fileName string) error {
	const trashInfoExt = ".trashinfo"
	const trashInfoSection = "[Trash Info]"

	infoFileTempPath := filepath.Join(infoDir, fileName+trashInfoExt+".tmp")
	infoFilePath := filepath.Join(infoDir, fileName+trashInfoExt)

	defer os.Remove(infoFileTempPath) // Ensure temporary file is cleaned up

	infoFile, err := os.OpenFile(infoFileTempPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, internal.InfoFilePermissions)
	if err != nil {
		return fmt.Errorf("cannot create info file '%s': %w", infoFileTempPath, err)
	}
	defer infoFile.Close()

	deletionDate := time.Now().Format(time.RFC3339)

	infoContent := fmt.Sprintf("%s\nPath=%s\nDeletionDate=%s\n", trashInfoSection, originalPath, deletionDate)

	_, err = infoFile.WriteString(infoContent)
	if err != nil {
		return fmt.Errorf("cannot write to info file '%s': %w", infoFileTempPath, err)
	}

	if err = os.Rename(infoFileTempPath, infoFilePath); err != nil {
		return fmt.Errorf("cannot rename temp info file '%s' to '%s': %w", infoFileTempPath, infoFilePath, err)
	}

	return nil
}
