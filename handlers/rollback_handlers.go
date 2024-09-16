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

//previous numbering logic
// func RevertToPreviousState() error {
// 	fmt.Println("Reading snapshot directory contents for revert")

// 	// Verify that the snapshots directory exists and can be read
// 	snapshots, err := os.ReadDir(snapshotDir)
// 	if err != nil {
// 		return fmt.Errorf("error reading snapshot directory: %v", err)
// 	}

// 	var currentSnapshot string
// 	var previousSnapshot string
// 	var currentNumber int
// 	var previousNumber int

// 	fmt.Println("Found snapshots:", snapshots) // Debugging line

// 	// Identify the current and previous snapshots
// 	for _, snapshot := range snapshots {
// 		if snapshot.IsDir() && strings.Contains(snapshot.Name(), "previous") {
// 			fmt.Printf("Processing snapshot for previous: %s\n", snapshot.Name())
// 			parts := strings.Split(snapshot.Name(), "_")
// 			if len(parts) > 1 {
// 				var num int
// 				fmt.Sscanf(parts[0], "@system%d", &num)
// 				fmt.Printf("Parsed previous snapshot number: %d\n", num) // Debugging line

// 				previousNumber = num
// 				previousSnapshot = snapshot.Name()

// 			}
// 		} else if snapshot.IsDir() && strings.Contains(snapshot.Name(), "current") {
// 			fmt.Printf("Processing snapshot for current: %s\n", snapshot.Name())
// 			parts := strings.Split(snapshot.Name(), "_")
// 			if len(parts) > 1 {
// 				fmt.Sscanf(parts[0], "@system%d", &currentNumber)
// 				currentSnapshot = snapshot.Name()
// 			}
// 		}
// 	}

// 	fmt.Println("Current snapshot:", currentSnapshot)   // Debugging line
// 	fmt.Println("Previous snapshot:", previousSnapshot) // Debugging line

// 	if currentSnapshot == "@system0_current" {
// 		fmt.Println("There is only one default state of the system saved.")
// 		return nil
// 	}

// 	if currentSnapshot == "" || previousSnapshot == "" {
// 		return fmt.Errorf("no current or previous snapshot found")
// 	}

// 	// Step 1: Delete the latest current snapshot
// 	currentSnapshotPath := filepath.Join(snapshotDir, currentSnapshot)
// 	fmt.Printf("Deleting current snapshot %s\n", currentSnapshotPath)
// 	resp := utils.ExecuteCommand("sudo", "rm", "-r", currentSnapshotPath)
// 	if resp.Error != "" {
// 		return fmt.Errorf("error deleting current snapshot %s: %v", currentSnapshotPath, resp.Error)
// 	}

// 	// // Step 2: Copy contents of previous snapshot to a new current snapshot
// 	// newCurrentSnapshotPath := filepath.Join(snapshotDir, fmt.Sprintf("@system%d_current", previousNumber))
// 	// fmt.Printf("Copying contents of %s to %s\n", previousSnapshot, newCurrentSnapshotPath)
// 	// resp = utils.ExecuteCommand("sudo", "cp", "-r", filepath.Join(snapshotDir, previousSnapshot), newCurrentSnapshotPath)
// 	// if resp.Error != "" {
// 	// 	return fmt.Errorf("error copying previous snapshot %s to current: %v", previousSnapshot, resp.Error)
// 	// }

// 	// Step 2: Create a snapshot of the previous snapshot as the new current snapshot
// 	newCurrentSnapshotPath := filepath.Join(snapshotDir, fmt.Sprintf("@system%d_current", previousNumber))
// 	fmt.Printf("Creating a snapshot of %s as %s\n", previousSnapshot, newCurrentSnapshotPath)
// 	resp = utils.ExecuteCommand("sudo", "btrfs", "subvolume", "snapshot", filepath.Join(snapshotDir, previousSnapshot), newCurrentSnapshotPath)
// 	if resp.Error != "" {
// 		return fmt.Errorf("error creating snapshot from %s to current: %v", previousSnapshot, resp.Error)
// 	}

// 	// Step 3: Move the previous snapshot to root with safety checks
// 	fmt.Printf("Preparing to move %s to /mnt/@\n", previousSnapshot)

// 	// Check if /mnt/tmp exists and delete it if necessary
// 	if _, err := os.Stat("/mnt/tmp"); err == nil {
// 		fmt.Println("/mnt/tmp exists, deleting it")
// 		resp := utils.ExecuteCommand("sudo", "rm", "-rf", "/mnt/tmp")
// 		if resp.Error != "" {
// 			return fmt.Errorf("error deleting /mnt/tmp: %v", resp.Error)
// 		}
// 	}

// 	// Check if /mnt/@ exists and move it to /mnt/tmp
// 	if _, err := os.Stat("/mnt/@"); err == nil {
// 		fmt.Println("/mnt/@ exists, moving to /mnt/tmp")
// 		resp := utils.ExecuteCommand("sudo", "mv", "/mnt/@", "/mnt/tmp")
// 		if resp.Error != "" {
// 			return fmt.Errorf("error moving /mnt/@ to /mnt/tmp: %v", resp.Error)
// 		}
// 	}

// 	// Move the previous snapshot to /mnt/@
// 	resp = utils.ExecuteCommand("sudo", "mv", filepath.Join(snapshotDir, previousSnapshot), "/mnt/@")
// 	if resp.Error != "" {
// 		fmt.Println("Error moving previous snapshot to /mnt/@, restoring /mnt/tmp back to /mnt/@")
// 		restoreResp := utils.ExecuteCommand("sudo", "mv", "/mnt/tmp", "/mnt/@")
// 		if restoreResp.Error != "" {
// 			return fmt.Errorf("error restoring /mnt/tmp back to /mnt/@: %v", restoreResp.Error)
// 		}
// 		return fmt.Errorf("error moving previous snapshot %s to /mnt/@: %v", previousSnapshot, resp.Error)
// 	}

// 	// Step 4: Rename the next most recent snapshot to previous
// 	newPreviousNumber := previousNumber - 1
// 	if newPreviousNumber >= 0 {
// 		newPreviousSnapshotPath := filepath.Join(snapshotDir, fmt.Sprintf("@system%d_previous", newPreviousNumber))
// 		previousSnapshotToRename := fmt.Sprintf("@system%d", newPreviousNumber)
// 		fmt.Printf("Renaming %s to %s\n", filepath.Join(snapshotDir, previousSnapshotToRename), newPreviousSnapshotPath)
// 		err = os.Rename(filepath.Join(snapshotDir, previousSnapshotToRename), newPreviousSnapshotPath)
// 		if err != nil {
// 			return fmt.Errorf("error renaming snapshot %s to %s: %v", previousSnapshotToRename, newPreviousSnapshotPath, err)
// 		}
// 	}

// 	// Add a pause before rebooting
// 	time.Sleep(1000 * time.Millisecond)

// 	// Step 5: Reboot the system
// 	fmt.Println("Rebooting the system")
// 	resp = utils.ExecuteCommand("sudo", "reboot")
// 	if resp.Error != "" {
// 		return fmt.Errorf("error executing reboot: %v", resp.Error)
// 	}

// 	return nil
// }


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
	resp := utils.ExecuteCommand("sudo", "rm", "-r", currentSnapshotPath)
	if resp.Error != "" {
		return fmt.Errorf("error deleting current snapshot %s: %v", currentSnapshotPath, resp.Error)
	}

	// Step 2: Move the previous snapshot to root with safety checks
	fmt.Printf("Preparing to move %s to /mnt/@\n", previousSnapshot)

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

	// Move the previous snapshot to /mnt/@
	resp = utils.ExecuteCommand("sudo", "mv", filepath.Join(snapshotDir, previousSnapshot), "/mnt/@")
	if resp.Error != "" {
		fmt.Println("Error moving previous snapshot to /mnt/@, restoring /mnt/tmp back to /mnt/@")
		restoreResp := utils.ExecuteCommand("sudo", "mv", "/mnt/tmp", "/mnt/@")
		if restoreResp.Error != "" {
			return fmt.Errorf("error restoring /mnt/tmp back to /mnt/@: %v", restoreResp.Error)
		}
		return fmt.Errorf("error moving previous snapshot %s to /mnt/@: %v", previousSnapshot, resp.Error)
	}

	// Add a pause before rebooting
	time.Sleep(1000 * time.Millisecond)

	// Step 3: Reboot the system
	fmt.Println("Rebooting the system")
	resp = utils.ExecuteCommand("sudo", "reboot")
	if resp.Error != "" {
		return fmt.Errorf("error executing reboot: %v", resp.Error)
	}

	return nil
}

