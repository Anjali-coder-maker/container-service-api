# Default Configuration Setup

## Overview

This Go application is designed to manage the initialization and configuration of services using the Podman API. It pulls images, checks for existing services, creates unit files, and enables and starts the services based on the specified configuration.

## Table of Contents

- [Requirements](#requirements)
- [Configuration](#configuration)
- [Initialization Process](#initialization-process)
- [Handlers](#handlers)
- [Running the Application](#running-the-application)

## Requirements

To run this application, ensure you have the following:

- Go (version 1.16 or later)
- Podman installed and configured on your system
- Access to the necessary service images in your Podman registry

## Configuration

The application's configuration is managed through a separate configuration file. You can define your services and their respective images in this configuration file. Below is an example structure of the configuration file:

```yaml
service.<service-name>.enable = true | false
```

The configuration file should be loaded through the `config.GetConfig()` function.

## Initialization Process

The `Initialize` function carries out the following steps for each service defined in the configuration:

1. **Load Configuration**: Load the configuration and confirm successful loading.
2. **Process Each Service**:
   - **Pull Image**: Pull the specified image for each service using the `handlers.PullImage` function.
   - **Check and Disable Existing Service**: Check if an existing service is running and disable it if necessary using `handlers.CheckAndDisableExistingService`.
   - **Create Unit File**: Create a unit file for the service using `handlers.CreateUnitFile`.
   - **Enable and Start Service**: Enable and start the service using `handlers.EnableAndStartService`.

## Handlers

The [handlers](/Docs/handlers/Default-handlers.md) implement the core functionality required to manage the services. Below are the key functions:

- `PullImage(serviceName string)`: Pulls the image for the specified service.
- `CheckAndDisableExistingService(serviceName string)`: Checks for and disables any currently running instance of the service.
- `CreateUnitFile(serviceName string)`: Creates a unit file for the specified service.
- `EnableAndStartService(serviceName string)`: Enables and starts the specified service.

## Running the Application

To run the application, use the following command:

```bash
go run main.go
```

Make sure your configuration file is correctly set up, and Podman is installed and configured on your system. The application will output the status of each step in the initialization process, helping you track any issues that may arise.
