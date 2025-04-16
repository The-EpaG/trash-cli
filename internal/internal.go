package internal

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/viper"
)

// Constants for better readability and maintainability, as per the Trash Specification.
const (
	trashDirName        = "Trash"
	trashInfoExt        = ".trashinfo"
	defaultPermissions  = 0700 // Owner-only rwx
	infoFilePermissions = 0600 // Owner-only rw
	localShareDirName   = ".local/share"
	trashInfoSection    = "[Trash Info]"
)

// getHomeTrashPaths determines the base trash directory and its subdirectories for the current user.
func GetHomeTrashPaths() (filesDir string, infoDir string, err error) {
	usr, err := user.Current()
	if err != nil {
		return "", "", fmt.Errorf("cannot get current user info: %w", err)
	}

	homeTrashDir := filepath.Join(usr.HomeDir, localShareDirName, trashDirName)
	filesDir = filepath.Join(homeTrashDir, viper.GetString("trash_files_dir"))
	infoDir = filepath.Join(homeTrashDir, viper.GetString("trash_info_dir"))
	return
}

// getTopTrashPaths determines the base trash directory and its subdirectories for a given top directory.
func GetTopTrashPaths(topDir string) (filesDir string, infoDir string, err error) {
	trashDir := filepath.Join(topDir, trashDirName)
	filesDir = filepath.Join(trashDir, viper.GetString("trash_files_dir"))
	infoDir = filepath.Join(trashDir, viper.GetString("trash_info_dir"))
	return
}

// GetTrashPaths determines the trash directory based on the XDG Trash specification and configuration
func GetTrashPaths(fileToTrashPath string) (filesDir string, infoDir string, err error) {
	// Check for top-level Trash directories
	var topDirs []string
	if fileToTrashPath != "" {
		// Find the top directory containing the file to trash
		absPath, err := filepath.Abs(fileToTrashPath)
		if err != nil {
			return "", "", fmt.Errorf("cannot get absolute path for '%s': %w", fileToTrashPath, err)
		}

		var dir string
		fileInfo, err := os.Stat(absPath)
		if err == nil && fileInfo.IsDir() {
			dir = absPath
		} else {
			dir = filepath.Dir(absPath)
		}

		for {
			// Check if the dir is top-level
			if dir == "/" || dir == "" {
				break
			}

			//check if .Trash exist in dir
			_, err = os.Stat(filepath.Join(dir, trashDirName))
			if err == nil {
				topDirs = append(topDirs, dir)
			}

			dir = filepath.Dir(dir)
		}
	}

	// use the best top-level trash if present
	if len(topDirs) > 0 {
		bestTopDir := topDirs[0]
		filesDir, infoDir, err = GetTopTrashPaths(bestTopDir)
		if err != nil {
			return "", "", fmt.Errorf("cannot get top trash paths: %w", err)
		}

		return filesDir, infoDir, nil
	}

	// Fallback to home trash directory
	filesDir, infoDir, err = GetHomeTrashPaths()
	if err != nil {
		return "", "", fmt.Errorf("cannot get home trash paths: %w", err)
	}

	return filesDir, infoDir, nil
}

// moveToTrash moves a file or directory to the trash directory and creates the corresponding .trashinfo file.
func MoveToTrash(filePath string, trashFilesDir, trashInfoDir string) (err error) {
	// Get the original absolute path before moving the file.
	originalAbsPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("cannot get absolute path: %w", err)
	}

	// Check if the source file exists and is accessible.
	_, err = os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file '%s' does not exist", filePath)
		}

		return fmt.Errorf("error checking file '%s': %w", filePath, err)
	}

	// Generate a unique ID for the file in the trash
	uniqueID := uuid.New().String()

	// Construct the final destination path using the unique ID
	finalFileNameInFiles := uniqueID
	destinationPath := filepath.Join(trashFilesDir, finalFileNameInFiles)

	// Create the .trashinfo file before moving the file
	if err = createTrashInfo(originalAbsPath, trashInfoDir, finalFileNameInFiles); err != nil {
		return fmt.Errorf("cannot create .trashinfo file: %w", err)
	}

	// Move the file/directory to the trash.
	if err = os.Rename(filePath, destinationPath); err != nil {
		return fmt.Errorf("cannot move '%s' to trash: %w", filePath, err)
	}

	return
}

// createTrashInfo creates a .trashinfo file with metadata about a trashed file, as per the Trash Specification.
func createTrashInfo(originalAbsPath, trashInfoDir, finalFileNameInFiles string) (err error) {
	infoFileName := finalFileNameInFiles + trashInfoExt
	// Use a temporary file to create the trashinfo atomically

	infoFileTmpPath := filepath.Join(trashInfoDir, infoFileName+".tmp")
	infoFilePath := filepath.Join(trashInfoDir, infoFileName)

	defer os.Remove(infoFileTmpPath) // Ensure temporary file is cleaned up

	infoFile, err := os.OpenFile(infoFileTmpPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, infoFilePermissions)
	if err != nil {
		return fmt.Errorf("cannot create info file '%s': %w", infoFileTmpPath, err)
	}
	defer infoFile.Close()

	// Get the current date and time in ISO 8601 format, as required by the specification.
	deletionDate := strings.Split(time.Now().Format(time.RFC3339), "+")[0]

	// Prepare the content of the .trashinfo file.
	infoContent := fmt.Sprintf("%s\nPath=%s\nDeletionDate=%s\n", trashInfoSection, originalAbsPath, deletionDate)

	_, err = infoFile.WriteString(infoContent)
	if err != nil {
		return fmt.Errorf("cannot write to info file '%s': %w", infoFileTmpPath, err)
	}
	if err = os.Rename(infoFileTmpPath, infoFilePath); err != nil {
		return fmt.Errorf("cannot rename temp info file '%s' to '%s': %w", infoFileTmpPath, infoFilePath, err)
	}

	return nil
}

// purgeTrash empties the trash directories.
func PurgeTrash(trashFilesDir, trashInfoDir string) error {
	if err := os.RemoveAll(trashFilesDir); err != nil {
		return fmt.Errorf("cannot empty trash files: %w", err)
	}
	return os.RemoveAll(trashInfoDir)
}

// TrashInfo struct to represent the data from .trashinfo files
type TrashInfo struct {
	Path         string
	DeletionDate string
	TrashFile    string
}

// parseTrashInfo reads and parses a .trashinfo file.
func ParseTrashInfo(infoFilePath string) (*TrashInfo, error) {
	content, err := os.ReadFile(infoFilePath)
	if err != nil {
		return nil, fmt.Errorf("cannot read info file '%s': %w", infoFilePath, err)
	}

	lines := strings.Split(string(content), "\n")
	info := &TrashInfo{}
	info.TrashFile = strings.TrimSuffix(filepath.Base(infoFilePath), trashInfoExt)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Path=") {
			info.Path = strings.TrimPrefix(line, "Path=")

		} else if strings.HasPrefix(line, "DeletionDate=") {
			info.DeletionDate = strings.TrimPrefix(line, "DeletionDate=")
		}

	}

	return info, nil
}

// listTrash lists the files in the trash.
func ListTrash(trashFilesDir, trashInfoDir string, showDetails bool) error {
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

		trashInfo, err := ParseTrashInfo(infoFilePath)
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

// restoreFromTrash restores a file from the trash.
func RestoreFromTrash(trashFilesDir, trashInfoDir string, fileNameInTrash string) error {
	infoFileName := fileNameInTrash + trashInfoExt
	infoFilePath := filepath.Join(trashInfoDir, infoFileName)
	trashFilePath := filepath.Join(trashFilesDir, fileNameInTrash)

	trashInfo, err := ParseTrashInfo(infoFilePath)
	if err != nil {
		return fmt.Errorf("cannot parse .trashinfo file: %w", err)
	}

	// Move the file back to its original location
	if err := os.Rename(trashFilePath, trashInfo.Path); err != nil {
		return fmt.Errorf("cannot move file back to original location: %w", err)
	}

	// Remove the .trashinfo file
	if err := os.Remove(infoFilePath); err != nil {
		return fmt.Errorf("cannot remove .trashinfo file: %w", err)
	}

	return nil
}