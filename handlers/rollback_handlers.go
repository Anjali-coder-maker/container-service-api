//go:build load

package handlers

import (
	"fmt"
	"go-podman-api/utils"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const configDirPath = "/etc/service-manager"
const defaultConfigFilePath = "/etc/service-manager/configuration.conf"
const snapshotDir = "/mnt/snapshots"

func MountDisk(devicePath string) error {
	fmt.Printf("Mounting disk at %s\n", devicePath)
	resp := utils.ExecuteCommand("sudo", "mount", devicePath, "/mnt")
	if resp.Error != "" {
		return fmt.Errorf("error mounting disk: %v", resp.Error)
	}
	return nil
}

func UnmountDisk() error {
	fmt.Println("Unmounting disk from /mnt")
	resp := utils.ExecuteCommand("sudo", "umount", "/mnt")
	if resp.Error != "" {
		return fmt.Errorf("error unmounting disk: %v", resp.Error)
	}
	return nil
}

func GetRootDevicePath() string {
	fmt.Println("Getting root device path")
	resp := utils.ExecuteCommand("sh", "-c", "df -Th | grep btrfs | grep /$ | cut -d' ' -f 1")
	if resp.Error != "" {
		return ""
	}
	return strings.TrimSpace(resp.Output)
}

func RevertToPreviousState() error {
	fmt.Println("Reading snapshot directory contents for revert")

	// Fetch snapshots
	snapshots, err := fetchSnapshots(snapshotDir)
	if err != nil {
		return fmt.Errorf("error reading snapshot directory: %v", err)
	}

	// Check if there is only one snapshot available
	if len(snapshots) == 1 {
		fmt.Println("There is only one current state saved. No previous state to roll back to.")
		return nil
	}

	// Find the current (latest) and previous (second latest) snapshots
	currentSnapshot, err := findCurrentState(snapshots)
	if err != nil {
		return fmt.Errorf("error finding current snapshot: %v", err)
	}

	previousSnapshot, err := findPreviousState(snapshots)
	if err != nil {
		return fmt.Errorf("error finding previous snapshot: %v", err)
	}

	// Debugging output
	fmt.Println("Current snapshot:", currentSnapshot)
	fmt.Println("Previous snapshot:", previousSnapshot)

	if currentSnapshot == "" || previousSnapshot == "" {
		return fmt.Errorf("no current or previous snapshot found")
	}

	// Step 1: Delete the current snapshot
	currentSnapshotPath := filepath.Join(snapshotDir, currentSnapshot)
	fmt.Printf("Deleting current snapshot %s\n", currentSnapshotPath)
	resp := utils.ExecuteCommand("sudo", "rm", "-rf", currentSnapshotPath)
	if resp.Error != "" {
		return fmt.Errorf("error deleting current snapshot %s: %v", currentSnapshotPath, resp.Error)
	}

	// Step 2: Create a Btrfs snapshot of the previous snapshot named 'previous_state'
	previousSnapshotPath := filepath.Join(snapshotDir, previousSnapshot)
	previousStateSnapshotPath := filepath.Join(snapshotDir, "previous_state")
	fmt.Printf("Creating Btrfs snapshot of %s as %s\n", previousSnapshotPath, previousStateSnapshotPath)
	resp = utils.ExecuteCommand("sudo", "btrfs", "subvolume", "snapshot", previousSnapshotPath, previousStateSnapshotPath)
	if resp.Error != "" {
		return fmt.Errorf("error creating Btrfs snapshot: %v", resp.Error)
	}

	// Step 3: Prepare to move 'previous_state' to /mnt/@

	// Check if /mnt/tmp exists and delete it if necessary
	if _, err := os.Stat("/mnt/tmp"); err == nil {
		fmt.Println("/mnt/tmp exists, deleting it")
		resp := utils.ExecuteCommand("sudo", "rm", "-rf", "/mnt/tmp")
		if resp.Error != "" {
			return fmt.Errorf("error deleting /mnt/tmp: %v", resp.Error)
		}
	}

	// Check if /mnt/@ exists and move it to /mnt/tmp
	if _, err := os.Stat("/mnt/@"); err == nil {
		fmt.Println("/mnt/@ exists, moving to /mnt/tmp")
		resp := utils.ExecuteCommand("sudo", "mv", "/mnt/@", "/mnt/tmp")
		if resp.Error != "" {
			return fmt.Errorf("error moving /mnt/@ to /mnt/tmp: %v", resp.Error)
		}
	}

	// Move 'previous_state' to /mnt/@
	fmt.Printf("Moving %s to /mnt/@\n", previousStateSnapshotPath)
	resp = utils.ExecuteCommand("sudo", "mv", previousStateSnapshotPath, "/mnt/@")
	if resp.Error != "" {
		fmt.Println("Error moving previous_state snapshot to /mnt/@, restoring /mnt/tmp back to /mnt/@")
		restoreResp := utils.ExecuteCommand("sudo", "mv", "/mnt/tmp", "/mnt/@")
		if restoreResp.Error != "" {
			return fmt.Errorf("error restoring /mnt/tmp back to /mnt/@: %v", restoreResp.Error)
		}
		return fmt.Errorf("error moving previous_state snapshot to /mnt/@: %v", resp.Error)
	}

	// Add a pause before rebooting
	time.Sleep(1000 * time.Millisecond)

	// Step 4: Reboot the system
	fmt.Println("Rebooting the system")
	resp = utils.ExecuteCommand("sudo", "reboot")
	if resp.Error != "" {
		return fmt.Errorf("error executing reboot: %v", resp.Error)
	}

	return nil
}
