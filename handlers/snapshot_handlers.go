//go:build load

package handlers

import (
	"crypto/sha256"
	"fmt"
	"go-podman-api/utils"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func IsConfigurationChanged() bool {
	// Fetch snapshots
	snapshots, err := fetchSnapshots(snapshotDir)
	if err != nil {
		fmt.Println("Error fetching snapshots:", err)
		return true // Assume configuration has changed if we can't fetch the snapshots
	}

	// Find the current state (latest snapshot)
	currentSnapshot, err := findCurrentState(snapshots)
	if err != nil {
		fmt.Println("Error finding current snapshot:", err)
		return true // Assume configuration has changed if no current snapshot is found
	}

	// Construct the full path to the configuration file in the current snapshot
	currentConfigFilePath := filepath.Join(snapshotDir, currentSnapshot, "etc/service-manager/configuration.conf")

	// Check if the configuration file exists in the current snapshot
	if _, err := os.Stat(currentConfigFilePath); os.IsNotExist(err) {
		fmt.Println("No configuration file found in the current snapshot. Assuming configuration has changed.")
		return true
	}

	// Calculate the hash of the declared (current) configuration
	currentHash, err := getFileHash(defaultConfigFilePath)
	if err != nil {
		fmt.Println("Error calculating hash of declared configuration:", err)
		return true
	}

	// Calculate the hash of the snapshot configuration
	snapshotHash, err := getFileHash(currentConfigFilePath)
	if err != nil {
		fmt.Println("Error calculating hash of snapshot configuration:", err)
		return true
	}

	// Compare the two hashes
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

func CreateNewSnapshot() error {
	// Get the current timestamp in the format ddmmyyyyhhmmss
	currentTime := time.Now().Format("02012006150405")
	newSnapshotName := fmt.Sprintf("system_%s", currentTime)
	fmt.Printf("Creating new snapshot %s\n", newSnapshotName)

	// Execute the command to create the snapshot with the new naming convention
	resp := utils.ExecuteCommand("sudo", "btrfs", "subvolume", "snapshot", "/mnt/@", filepath.Join(snapshotDir, newSnapshotName))
	if resp.Error != "" {
		return fmt.Errorf("error creating snapshot %s: %v", newSnapshotName, resp.Error)
	}

	fmt.Println("Snapshot", newSnapshotName, "created successfully.")
	return nil
}

func fetchSnapshots(snapshotDir string) ([]string, error) {
	// Open the snapshot directory
	files, err := os.ReadDir(snapshotDir)
	if err != nil {
		return nil, fmt.Errorf("error reading snapshot directory: %v", err)
	}

	// Filter and collect snapshot names that match the pattern "system_[timestamp]"
	var snapshots []string
	for _, file := range files {
		if file.IsDir() && strings.HasPrefix(file.Name(), "system_") {
			snapshots = append(snapshots, file.Name())
		}
	}
	return snapshots, nil
}

func findCurrentState(snapshots []string) (string, error) {
	if len(snapshots) == 0 {
		return "", fmt.Errorf("no snapshots available")
	}

	// Sort the snapshots lexicographically (timestamp-based sorting)
	sort.Strings(snapshots)

	// Get the latest (largest) snapshot
	currentState := snapshots[len(snapshots)-1]
	return currentState, nil
}

func findPreviousState(snapshots []string) (string, error) {
	if len(snapshots) <= 1 {
		return "", fmt.Errorf("no previous snapshot available")
	}

	// Sort the snapshots lexicographically (timestamp-based sorting)
	sort.Strings(snapshots)

	// Get the second latest (second largest) snapshot
	previousState := snapshots[len(snapshots)-2]
	return previousState, nil
}
