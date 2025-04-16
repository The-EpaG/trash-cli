package internal

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Constants for better readability and maintainability, as per the Trash Specification.
const (
	TrashDirName        = "Trash"
	TrashInfoExt        = ".trashinfo"
	DefaultPermissions  = 0700 // Owner-only rwx
	InfoFilePermissions = 0600 // Owner-only rw
	LocalShareDirName   = ".local/share"
	TrashInfoSection    = "[Trash Info]"
)

// getHomeTrashPaths determines the base trash directory and its subdirectories for the current user.
func GetHomeTrashPaths() (filesDir string, infoDir string, err error) {
	usr, err := user.Current()
	if err != nil {
		return "", "", fmt.Errorf("cannot get current user info: %w", err)
	}

	homeTrashDir := filepath.Join(usr.HomeDir, LocalShareDirName, TrashDirName)
	filesDir = filepath.Join(homeTrashDir, viper.GetString("trash_files_dir"))
	infoDir = filepath.Join(homeTrashDir, viper.GetString("trash_info_dir"))
	return
}

// getTopTrashPaths determines the base trash directory and its subdirectories for a given top directory.
func GetTopTrashPaths(topDir string) (filesDir string, infoDir string, err error) {
	trashDir := filepath.Join(topDir, TrashDirName)
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
			_, err = os.Stat(filepath.Join(dir, TrashDirName))
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
	info.TrashFile = strings.TrimSuffix(filepath.Base(infoFilePath), TrashInfoExt)

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