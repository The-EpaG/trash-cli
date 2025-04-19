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

var RemoveCmd = &cobra.Command{
	Use:     "remove [file...]",
	Aliases: []string{"rm", "delete", "del"},
	Short:   "Move a file or directory to the trash",
	Long:    `Move specified files or directories to the trash.`,
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		operationSuccessful := true
		for _, file := range args {
			var tempTrashFilesDir, tempTrashInfoDir string
			tempTrashFilesDir, tempTrashInfoDir, err := internal.GetTrashPaths(file)

			// Skip this file if paths aren't found
			if err != nil {
				log.Printf("Error: cannot retrieve trash path for '%s': %v", file, err)
				operationSuccessful = false
				continue
			}
			if err = moveToTrash(file, tempTrashFilesDir, tempTrashInfoDir); err != nil {
				log.Printf("Error moving '%s' to trash: %v", file, err)
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

	// Verify that the source  exists and is accessible.
	if _, err = os.Stat(sourcePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s does not exist", sourcePath)
		}
		return fmt.Errorf("error accessing file %s: %w", sourcePath, err)
	}

	// Generate a unique identifier for the trashed file.
	uniqueIdentifier := uuid.New().String()

	// Construct the destination path in the trash using the unique identifier.
	destinationPath := filepath.Join(trashFilesDir, uniqueIdentifier)

	// Create the .trashinfo file for the source file.
	if err = createTrashInfo(absolutePath, trashInfoDir, uniqueIdentifier); err != nil {
		// Attempt to clean up if info file creation fails before move
		// This part is tricky, maybe log and continue is better?
		// For now, just return the error.
		return fmt.Errorf("cannot create .trashinfo file for %s: %w", sourcePath, err)
	}

	// Move the source file or directory to the trash.
	if err = os.Rename(sourcePath, destinationPath); err != nil {
		// If the move fails, we should ideally remove the .trashinfo file we just created.
		infoFilePath := filepath.Join(trashInfoDir, uniqueIdentifier+internal.TrashInfoExt)
		_ = os.Remove(infoFilePath) // Attempt cleanup, ignore error as the primary error is the Rename failure.
		return fmt.Errorf("cannot move %s to trash: %w", sourcePath, err)
	}

	return nil
}

// createTrashInfo creates a .trashinfo file with metadata about a trashed file, as per the Trash Specification.
func createTrashInfo(originalPath, infoDir, fileName string) error {
	// Using internal constant now
	infoFileTempPath := filepath.Join(infoDir, fileName+internal.TrashInfoExt+".tmp")
	infoFilePath := filepath.Join(infoDir, fileName+internal.TrashInfoExt)

	// Ensure infoDir exists
	if err := os.MkdirAll(infoDir, internal.DefaultPermissions); err != nil {
		return fmt.Errorf("cannot create info directory '%s': %w", infoDir, err)
	}

	defer os.Remove(infoFileTempPath) // Ensure temporary file is cleaned up

	infoFile, err := os.OpenFile(infoFileTempPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, internal.InfoFilePermissions)
	if err != nil {
		return fmt.Errorf("cannot create temp info file '%s': %w", infoFileTempPath, err)
	}
	// Use defer for closing the file to ensure it happens even on errors during write/rename
	defer infoFile.Close()

	deletionDate := time.Now().Format(time.RFC3339)

	// Using internal constant now
	infoContent := fmt.Sprintf("%s\nPath=%s\nDeletionDate=%s\n", internal.TrashInfoSection, originalPath, deletionDate)

	_, err = infoFile.WriteString(infoContent)
	if err != nil {
		return fmt.Errorf("cannot write to temp info file '%s': %w", infoFileTempPath, err)
	}

	// Ensure data is written to disk before renaming
	if err = infoFile.Sync(); err != nil {
		return fmt.Errorf("cannot sync temp info file '%s': %w", infoFileTempPath, err)
	}

	// Close the file explicitly before renaming, especially important on some OSes
	if err = infoFile.Close(); err != nil {
		// Log or return error? Renaming might still work, but data might be incomplete.
		// Let's return the error for safety.
		return fmt.Errorf("cannot close temp info file '%s' before rename: %w", infoFileTempPath, err)
	}

	if err = os.Rename(infoFileTempPath, infoFilePath); err != nil {
		return fmt.Errorf("cannot rename temp info file '%s' to '%s': %w", infoFileTempPath, infoFilePath, err)
	}

	return nil
}
