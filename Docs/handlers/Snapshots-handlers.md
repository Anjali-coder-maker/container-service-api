# Podman Service Snapshot Handlers

## Overview

This Go package provides handler functions for managing Btrfs snapshots within the Podman service management system. It includes functionalities for checking configuration changes, creating new snapshots, and retrieving existing snapshots. The snapshot system allows for efficient state management and recovery in case of configuration issues.

## Table of Contents

- [Requirements](#requirements)
- [Handlers Overview](#handlers-overview)
- [Functions](#functions)
  - [IsConfigurationChanged](#isconfigurationchanged)
  - [CreateNewSnapshot](#createnewsnapshot)
  - [fetchSnapshots](#fetchsnapshots)
  - [findCurrentState](#findcurrentstate)
  - [findPreviousState](#findpreviousstate)
  - [getFileHash](#getfilehash)
- [Error Handling](#error-handling)
- [Usage](#usage)

## Requirements

To use this package, ensure you have the following:

- Go (version 1.16 or later)
- Podman installed and configured on your system
- Btrfs filesystem for snapshot management

## Handlers Overview

The following functions are included in the `handlers` package for managing snapshots:

### IsConfigurationChanged

This function checks if the current configuration differs from the configuration stored in the latest snapshot.

#### Signature

```go
func IsConfigurationChanged() bool
```

#### Returns

- `bool`: Returns `true` if the configuration has changed; `false` otherwise.

### CreateNewSnapshot

This function creates a new snapshot of the current state.

#### Signature

```go
func CreateNewSnapshot() error
```

#### Returns

- `error`: An error if creating the snapshot fails.

### fetchSnapshots

This function retrieves a list of existing snapshots from the specified snapshot directory.

#### Signature

```go
func fetchSnapshots(snapshotDir string) ([]string, error)
```

#### Parameters

- `snapshotDir`: The directory where snapshots are stored.

#### Returns

- `([]string, error)`: A list of snapshot names and any error encountered.

### findCurrentState

This function identifies the latest snapshot based on the timestamp in the snapshot names.

#### Signature

```go
func findCurrentState(snapshots []string) (string, error)
```

#### Parameters

- `snapshots`: A list of snapshot names.

#### Returns

- `(string, error)`: The name of the latest snapshot and any error encountered.

### findPreviousState

This function identifies the second latest snapshot, which represents the previous state.

#### Signature

```go
func findPreviousState(snapshots []string) (string, error)
```

#### Parameters

- `snapshots`: A list of snapshot names.

#### Returns

- `(string, error)`: The name of the previous snapshot and any error encountered.

### getFileHash

This function calculates the SHA-256 hash of a specified file.

#### Signature

```go
func getFileHash(filePath string) (string, error)
```

#### Parameters

- `filePath`: The path to the file for which the hash should be calculated.

#### Returns

- `(string, error)`: The hash of the file and any error encountered.

## Error Handling

The handlers provide structured error messages for debugging and issue resolution. Each function returns an error if an operation fails, allowing for better error handling in the application.

## Usage

To use the snapshot handlers in your application, ensure you import the `handlers` package and call the functions as needed. For example, to create a new snapshot, you might do the following:

```go
err := handlers.CreateNewSnapshot()
if err != nil {
    log.Println("Error creating snapshot:", err)
} else {
    log.Println("Snapshot created successfully.")
}
```

### Initialization

Ensure that the necessary configurations and Btrfs snapshots are in place before using the snapshot handlers.
