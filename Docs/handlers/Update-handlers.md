# Podman Service Update Handlers

## Overview

This Go package provides handler functions for managing updates to Podman services. It includes functionalities for pulling the latest container images, checking and updating services based on their current state, and restarting services as necessary. The update system ensures that services are running the latest versions of their respective images.

## Table of Contents

- [Requirements](#requirements)
- [Handlers Overview](#handlers-overview)
- [Functions](#functions)
  - [getCurrentContainerImageDigest](#getcurrentcontainerimagedigest)
  - [getRemoteContainerImageDigest](#getremotecontainerimagedigest)
  - [pullLatestImage](#pulllatestimage)
  - [checkAndUpdateService](#checkandupdateservice)
  - [restartService](#restartservice)
  - [UpdateServices](#updateservices)
- [Error Handling](#error-handling)
- [Usage](#usage)

## Requirements

To use this package, ensure you have the following:

- Go (version 1.16 or later)
- Podman installed and configured on your system
- Skopeo for inspecting remote images

## Handlers Overview

The following functions are included in the `handlers` package for managing service updates:

### getCurrentContainerImageDigest

This function retrieves the digest of the currently installed image for a specified service.

#### Signature

```go
func getCurrentContainerImageDigest(service string) (string, error)
```

#### Parameters

- `service`: The name of the service to check.

#### Returns

- `(string, error)`: The image digest and any error encountered.

### getRemoteContainerImageDigest

This function retrieves the digest of the remote image for a specified service.

#### Signature

```go
func getRemoteContainerImageDigest(service string) (string, error)
```

#### Parameters

- `service`: The name of the service to check.

#### Returns

- `(string, error)`: The remote image digest and any error encountered.

### pullLatestImage

This function pulls the latest container image for a specified service.

#### Signature

```go
func pullLatestImage(service string) (string, error)
```

#### Parameters

- `service`: The name of the service for which to pull the image.

#### Returns

- `(string, error)`: A success message or an error if the pull fails.

### checkAndUpdateService

This function checks if the local image for a service is up to date with the remote version and updates it if necessary.

#### Signature

```go
func checkAndUpdateService(service string, enabled bool) (bool, error)
```

#### Parameters

- `service`: The name of the service to check.
- `enabled`: A boolean indicating whether the service is enabled.

#### Returns

- `(bool, error)`: A boolean indicating if an update occurred and any error encountered.

### restartService

This function restarts a specified service after pulling the latest image.

#### Signature

```go
func restartService(service string) error
```

#### Parameters

- `service`: The name of the service to restart.

#### Returns

- `error`: An error if the restart process fails.

### UpdateServices

This function reads the service configuration from a file and checks each service for updates.

#### Signature

```go
func UpdateServices(filePath string) (bool, error)
```

#### Parameters

- `filePath`: The path to the configuration file.

#### Returns

- `(bool, error)`: A boolean indicating if any updates were made and any error encountered.

## Error Handling

The handlers provide structured error messages for debugging and issue resolution. Each function returns an error if an operation fails, allowing for better error handling in the application.

## Usage

To use the update handlers in your application, ensure you import the `handlers` package and call the functions as needed. For example, to update services based on the configuration file, you might do the following:

```go
updated, err := handlers.UpdateServices("/path/to/configuration.conf")
if err != nil {
    log.Println("Error updating services:", err)
} else if updated {
    log.Println("Services updated successfully.");
} else {
    log.Println("No updates were made.");
}
```

### Initialization

Ensure that the necessary configurations and Podman images are in place before using the update handlers.
