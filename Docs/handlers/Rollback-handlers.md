# Podman Service Rollback Handlers

## Overview

This Go package provides handler functions specifically for managing rollback operations in the Podman service management system. It includes functionalities for mounting and unmounting disks, retrieving device paths, and reverting to a previous system state using Btrfs snapshots.

## Table of Contents

- [Requirements](#requirements)
- [Handlers Overview](#handlers-overview)
- [Functions](#functions)
  - [MountDisk](#mountdisk)
  - [UnmountDisk](#unmountdisk)
  - [GetRootDevicePath](#getrootdevicepath)
  - [RevertToPreviousState](#reverttopreviousstate)
- [Error Handling](#error-handling)
- [Usage](#usage)

## Requirements

To use this package, ensure you have the following:

- Go (version 1.16 or later)
- Podman installed and configured on your system
- Systemd for service management
- Btrfs filesystem for snapshot management

## Handlers Overview

The following functions are included in the `handlers` package for managing rollback operations:

### MountDisk

This function mounts a specified disk at `/mnt`.

#### Signature

```go
func MountDisk(devicePath string) error
```

#### Parameters

- `devicePath`: The path of the device to mount.

#### Returns

- `error`: An error if mounting the disk fails.

### UnmountDisk

This function unmounts the disk from `/mnt`.

#### Signature

```go
func UnmountDisk() error
```

#### Returns

- `error`: An error if unmounting the disk fails.

### GetRootDevicePath

This function retrieves the root device path of the Btrfs filesystem.

#### Signature

```go
func GetRootDevicePath() string
```

#### Returns

- `string`: The root device path, or an empty string if an error occurs.

### RevertToPreviousState

This function reverts the system to its previous state using Btrfs snapshots.

#### Signature

```go
func RevertToPreviousState() error
```

#### Returns

- `error`: An error if reverting to the previous state fails.

#### Functionality

1. **Snapshot Management**: Reads the snapshot directory and checks for existing snapshots.
2. **Current and Previous Snapshot Identification**: Identifies the current and previous snapshots.
3. **Snapshot Deletion**: Deletes the current snapshot.
4. **Snapshot Creation**: Creates a new Btrfs snapshot based on the previous snapshot.
5. **Directory Management**: Moves directories as needed to restore the previous state.
6. **System Reboot**: Reboots the system to apply the changes.

## Error Handling

The handlers provide structured error messages for debugging and issue resolution. Each function returns an error if an operation fails, allowing for better error handling in the application.

## Usage

To use the rollback handlers in your application, ensure you import the `handlers` package and call the functions as needed. For example, to revert to the previous state, you might do the following:

```go
err := handlers.RevertToPreviousState()
if err != nil {
    log.Println("Error reverting to previous state:", err)
} else {
    log.Println("Successfully reverted to previous state.")
}
```

### Initialization

Ensure that the necessary configurations and snapshots are in place before using the rollback handlers.
