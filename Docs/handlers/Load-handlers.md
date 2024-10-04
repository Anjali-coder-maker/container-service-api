# Podman Service Handlers

## Overview

This Go package provides handler functions for managing Podman services, including pulling images, enabling or disabling services, and creating systemd unit files. The handlers interact with the systemd service manager and use the Podman container engine to ensure services are properly configured and running.

## Table of Contents

- [Requirements](#requirements)
- [Handlers Overview](#handlers-overview)
- [Functions](#functions)
  - [GetServiceState](#getservicestate)
  - [DisableService](#disableservice)
  - [EnableService](#enableservice)
  - [PullImageChroot](#pullimagechroot)
  - [CreateAndPlaceUnitFile](#createandplaceunitfile)
  - [MoveOverlayUpperToRoot](#moveoverlayuppertoroot)
  - [ReadConfigurations](#readconfigurations)
- [Error Handling](#error-handling)
- [Usage](#usage)

## Requirements

To use this package, ensure you have the following:

- Go (version 1.16 or later)
- Podman installed and configured on your system
- Systemd for service management

## Handlers Overview

The following functions are included in the `handlers` package:

### GetServiceState

This function checks the state of a service, determining if it is "enabled", "disabled", or "not found".

#### Signature

```go
func GetServiceState(service string) (string, error)
```

#### Parameters

- `service`: The name of the service to check.

#### Returns

- `(string, error)`: The state of the service (enabled, disabled, or not-found) and any error encountered.

### DisableService

This function disables a specified service and stops it if it is currently running.

#### Signature

```go
func DisableService(service string) error
```

#### Parameters

- `service`: The name of the service to disable.

#### Returns

- `error`: An error if disabling the service fails.

### EnableService

This function enables and starts a specified service.

#### Signature

```go
func EnableService(service string) error
```

#### Parameters

- `service`: The name of the service to enable and start.

#### Returns

- `error`: An error if enabling or starting the service fails.

### PullImageChroot

This function pulls a container image inside a chroot environment.

#### Signature

```go
func PullImageChroot(serviceName string, chrootpath string) (string, error)
```

#### Parameters

- `serviceName`: The name of the service whose image should be pulled.
- `chrootpath`: The path to the chroot environment.

#### Returns

- `(string, error)`: The pulled image name and any error encountered.

### CreateAndPlaceUnitFile

This function creates a systemd unit file for a service and places it in the specified chroot directory.

#### Signature

```go
func CreateAndPlaceUnitFile(serviceName, chrootDir string, serviceConfig config.ServiceConfig) error
```

#### Parameters

- `serviceName`: The name of the service.
- `chrootDir`: The directory where the unit file should be placed.
- `serviceConfig`: The configuration for the service.

#### Returns

- `error`: An error if creating or placing the unit file fails.

### MoveOverlayUpperToRoot

This function moves the overlay upper directory to the root directory.

#### Signature

```go
func MoveOverlayUpperToRoot() error
```

#### Returns

- `error`: An error if moving the overlay upper directory fails.

### ReadConfigurations

This function reads configurations from a specified file.

#### Signature

```go
func ReadConfigurations(filePath string) (map[string]bool, error)
```

#### Parameters

- `filePath`: The path to the configuration file.

#### Returns

- `(map[string]bool, error)`: A map of service names and their enabled states, and any error encountered.

## Error Handling

The handlers provide structured error messages for debugging and issue resolution. Each function returns an error if an operation fails, allowing for better error handling in the application.

## Usage

To use the handlers in your application, ensure you import the `handlers` package and call the functions as needed. For example, to enable a service, you might do the following:

```go
err := handlers.EnableService("my-service")
if err != nil {
    log.Println("Error enabling service:", err)
} else {
    log.Println("Service enabled successfully.")
}
```

### Initialization

The handlers rely on the configuration defined in the `go-podman-api/config` package to manage services. Ensure that the configuration is correctly set up before using the handlers.
