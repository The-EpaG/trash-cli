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

// GetHomeTrashPaths determines the base trash directory and its subdirectories for the current user.
func GetHomeTrashPaths() (trashFilesDir string, trashInfoDir string, err error) {
	usr, err := user.Current()
	if err != nil {
		return "", "", fmt.Errorf("cannot get current user info: %w", err)
	}

	trashHomeDir := filepath.Join(usr.HomeDir, LocalShareDirName, TrashDirName)
	trashFilesDir = filepath.Join(trashHomeDir, viper.GetString("trash.filesDir"))
	trashInfoDir = filepath.Join(trashHomeDir, viper.GetString("trash.infoDir"))
	return trashFilesDir, trashInfoDir, nil
}

// GetTopTrashPaths determines the base trash directory and its subdirectories for a given top directory.
func GetTopTrashPaths(topDirectory string) (filesDirectory string, infoDirectory string, err error) {
	trashDirectory := filepath.Join(topDirectory, TrashDirName)
	filesDirectory = filepath.Join(trashDirectory, viper.GetString("trash.filesDir"))
	infoDirectory = filepath.Join(trashDirectory, viper.GetString("trash.infoDir"))
	return
}

// GetTrashPaths determines the trash directory based on the XDG Trash specification and configuration
func GetTrashPaths(fileToTrashPath string) (filesDir string, infoDir string, err error) {
	topLevelDirs := []string{}
	if fileToTrashPath != "" {
		absPath, err := filepath.Abs(fileToTrashPath)
		if err != nil {
			return "", "", fmt.Errorf("cannot get absolute path for %q: %w", fileToTrashPath, err)
		}

		// Find the top directory containing the file to trash
		var dir string
		fileInfo, err := os.Stat(absPath)
		if err == nil && fileInfo.IsDir() {
			dir = absPath
		} else {
			dir = filepath.Dir(absPath)
		}

		// Check for top-level Trash directories
		for {
			if dir == "/" || dir == "" {
				break
			}

			trashDir := filepath.Join(dir, TrashDirName)
			if _, err = os.Stat(trashDir); err == nil {
				topLevelDirs = append(topLevelDirs, dir)
			}

			dir = filepath.Dir(dir)
		}
	}

	// Use the best top-level trash if present
	if len(topLevelDirs) > 0 {
		bestTopDir := topLevelDirs[0]
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

// ParseTrashInfo reads and parses a .trashinfo file.
func ParseTrashInfo(infoFilePath string) (*TrashInfo, error) {
	data, err := os.ReadFile(infoFilePath)
	if err != nil {
		return nil, fmt.Errorf("cannot read info file '%s': %w", infoFilePath, err)
	}

	info := &TrashInfo{
		TrashFile: strings.TrimSuffix(filepath.Base(infoFilePath), TrashInfoExt),
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(line, "Path="):
			info.Path = strings.TrimPrefix(line, "Path=")

		case strings.HasPrefix(line, "DeletionDate="):
			info.DeletionDate = strings.TrimPrefix(line, "DeletionDate=")
		}
	}

	return info, nil
}
