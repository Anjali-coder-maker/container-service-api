//go:build load

package handlers

import (
	"crypto/sha256"
	"fmt"
	"go-podman-api/utils"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Comparing current state of the system with the declared state in which user wants to load it's system
func IsConfigurationChanged() bool {
	snapshots, err := os.ReadDir(snapshotDir)
	if err != nil {
		fmt.Println("Error reading snapshot directory:", err)
		return true // Assume configuration has changed if we can't read the directory
	}

	var currentConfigFilePath string

	// Find the current snapshot configuration file
	for _, snapshot := range snapshots {
		if snapshot.IsDir() && strings.Contains(snapshot.Name(), "current") {
			currentConfigFilePath = filepath.Join(snapshotDir, snapshot.Name(), "etc/service-manager/configuration.conf")
			break
		}
	}

	if currentConfigFilePath == "" {
		fmt.Println("No current snapshot configuration found. Assuming configuration is new.")
		return true
	}

	currentHash, err := getFileHash(defaultConfigFilePath)
	if err != nil {
		fmt.Println("Error calculating hash of current configuration:", err)
		return true
	}

	snapshotHash, err := getFileHash(currentConfigFilePath)
	if err != nil {
		fmt.Println("Error calculating hash of snapshot configuration:", err)
		return true
	}
	return currentHash != snapshotHash
}

func getFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func ManageSnapshots() error {
	fmt.Println("Reading snapshot directory contents")
	snapshots, err := os.ReadDir(snapshotDir)
	if err != nil {
		return fmt.Errorf("error reading snapshot directory: %v", err)
	}

	var latestSnapshot string
	var previousSnapshot string
	var latestNumber int
	var previousNumber int

	// Identify the latest current and previous snapshots
	for _, snapshot := range snapshots {
		if snapshot.IsDir() {
			if strings.Contains(snapshot.Name(), "current") {
				latestSnapshot = snapshot.Name()
				parts := strings.Split(snapshot.Name(), "_")
				if len(parts) > 1 {
					fmt.Sscanf(parts[0], "@system%d", &latestNumber)
				}
			} else if strings.Contains(snapshot.Name(), "previous") {
				previousSnapshot = snapshot.Name()
				parts := strings.Split(snapshot.Name(), "_")
				if len(parts) > 1 {
					fmt.Sscanf(parts[0], "@system%d", &previousNumber)
				}
			}
		}
	}

	// Check the snapshot scenarios and manage accordingly
	if latestSnapshot == "" && previousSnapshot == "" {
		fmt.Println("No snapshots found, continuing with other work.")
		return nil
	} else if latestSnapshot != "" && previousSnapshot == "" {
		fmt.Printf("Renaming %s to @system%d_previous\n", latestSnapshot, latestNumber)
		return renameSnapshot(latestSnapshot, fmt.Sprintf("@system%d_previous", latestNumber))
	} else if latestSnapshot != "" && previousSnapshot != "" {
		fmt.Printf("Renaming %s to @system%d\n", previousSnapshot, previousNumber)
		err = renameSnapshot(previousSnapshot, fmt.Sprintf("@system%d", previousNumber))
		if err != nil {
			return err
		}
		fmt.Printf("Renaming %s to @system%d_previous\n", latestSnapshot, latestNumber)
		return renameSnapshot(latestSnapshot, fmt.Sprintf("@system%d_previous", latestNumber))
	}

	return nil
}

func renameSnapshot(oldName, newName string) error {
	fmt.Printf("Renaming snapshot %s to %s\n", oldName, newName)
	oldPath := filepath.Join(snapshotDir, oldName)
	newPath := filepath.Join(snapshotDir, newName)
	err := os.Rename(oldPath, newPath)
	if err != nil {
		return fmt.Errorf("error renaming snapshot %s to %s: %v", oldName, newName, err)
	}
	return nil
}

func CreateNewSnapshot() error {
	latestSnapshot := getLatestSnapshotNumber()
	newSnapshotName := fmt.Sprintf("@system%d_current", latestSnapshot+1)
	fmt.Printf("Creating new snapshot %s\n", newSnapshotName)
	resp := utils.ExecuteCommand("sudo", "btrfs", "subvolume", "snapshot", "/mnt/@", filepath.Join(snapshotDir, newSnapshotName))
	if resp.Error != "" {
		return fmt.Errorf("error creating snapshot %s: %v", newSnapshotName, resp.Error)
	}
	fmt.Println("Snapshot", newSnapshotName, "created successfully.")
	return nil
}

func getLatestSnapshotNumber() int {
	latestNumber := 0
	snapshots, err := os.ReadDir(snapshotDir)
	if err != nil {
		fmt.Println("Error reading snapshot directory:", err)
		return latestNumber
	}

	for _, snapshot := range snapshots {
		if snapshot.IsDir() {
			parts := strings.Split(snapshot.Name(), "_")
			if len(parts) > 1 {
				var num int
				fmt.Sscanf(parts[0], "@system%d", &num)
				if num > latestNumber {
					latestNumber = num
				}
			}
		}
	}

	return latestNumber
}
