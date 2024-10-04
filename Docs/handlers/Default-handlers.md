# Podman Service Handlers

## Overview

This Go package provides various handler functions to manage services using the Podman API. It allows pulling images, creating systemd unit files, checking existing services, and enabling or starting services based on a specified configuration.

## Table of Contents

- [Requirements](#requirements)
- [Configuration Loading](#configuration-loading)
- [Handlers Overview](#handlers-overview)
- [Functions](#functions)
  - [PullImage](#pullimage)
  - [CreateUnitFile](#createunitfile)
  - [CheckAndDisableExistingService](#checkanddisableexistingservice)
  - [EnableAndStartService](#enableandstartservice)
- [Error Handling](#error-handling)
- [Usage](#usage)

## Requirements

To use this package, ensure you have the following:

- Go (version 1.16 or later)
- Podman installed and configured on your system
- Systemd for service management

## Configuration Loading

The handlers package relies on a configuration structure defined in the `go-podman-api/config` package. The configuration should specify the services to be managed, including their names, whether they are enabled, and their respective unit file templates.

## Handlers Overview

The following functions are included in the `handlers` package:

### PullImage

This function pulls an image from a registry based on the configuration file.

#### Signature

```go
func PullImage(imageName string) utils.CommandResponse
```

#### Parameters

- `imageName`: The name of the service to pull.

#### Returns

- `utils.CommandResponse`: Contains the output or error message from the command execution.

### CreateUnitFile

This function creates a systemd unit file for the specified service.

#### Signature

```go
func CreateUnitFile(serviceName string) utils.CommandResponse
```

#### Parameters

- `serviceName`: The name of the service for which to create the unit file.

#### Returns

- `utils.CommandResponse`: Contains the output or error message from the command execution.

### CheckAndDisableExistingService

This function checks if a service is currently active and disables it if necessary.

#### Signature

```go
func CheckAndDisableExistingService(imageName string) bool
```

#### Parameters

- `imageName`: The name of the service to check.

#### Returns

- `bool`: Indicates whether the service was successfully checked and disabled.

### EnableAndStartService

This function enables and starts the specified systemd service.

#### Signature

```go
func EnableAndStartService(imageName string) utils.CommandResponse
```

#### Parameters

- `imageName`: The name of the service to enable and start.

#### Returns

- `utils.CommandResponse`: Contains the output or error message from the command execution.

## Error Handling

The handlers provide error messages to facilitate debugging and ensure clear communication of issues encountered during execution. The use of `utils.CommandResponse` allows for structured error handling and output management.

## Usage

To use the handlers in your application, ensure you import the `handlers` package and call the functions as needed. For example, to pull an image for a service, you might do the following:

```go
response := handlers.PullImage("my-service")
if response.Error != "" {
    log.Println("Error pulling image:", response.Error)
} else {
    log.Println("Image pulled successfully:", response.Output)
}
```

### Initialization

The configuration is automatically loaded when the `handlers` package is initialized, making it easy to use the functions without additional setup.

---

Feel free to modify any sections as necessary to better fit your project or to add specific examples and use cases!
