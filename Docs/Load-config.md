# Load Configuration Setup

## Overview

This Go application manages the loading and updating of service configurations for the Podman API. It allows you to load a configuration file, update services based on the configuration, and revert to a previous state if needed.

## Table of Contents

- [Requirements](#requirements)
- [Configuration File](#configuration-file)
- [Command-Line Flags](#command-line-flags)
- [Loading the Configuration](#loading-the-configuration)
- [Updating Services](#updating-services)
- [Reverting Changes](#reverting-changes)
- [Running the Application](#running-the-application)

## Requirements

To run this application, ensure you have the following:

- Go (version 1.16 or later)
- Podman installed and configured on your system
- Access to the necessary service images in your Podman registry

## Configuration File

The application loads its configurations from a specified configuration file. The default configuration file path is `/etc/service-manager/configuration.conf`. You can define services in the configuration file in the following format:

```ini
# Define services in the following format:
# service.<service_name>.enable = true|false
```

## Command-Line Flags

The application supports the following command-line flags:

- `--load <file-path>`: Load the configuration from the specified file path. If not specified, the default configuration file is used.

  **Example**:
  ```bash
  service-manager --load /path/to/custom/config/file
  ```

- `--update`: Update the services based on the configuration file.

  **Example**:
  ```bash
  service-manager --update
  ```

- `--rollback`: Revert to the previous state.

## Loading the Configuration

To load the configuration, the application performs the following steps:

1. **Check and Create Configuration File**: The application ensures that the configuration file exists at the specified location. If it doesn't, it creates one with default content.

2. **Mount Root Device**: The application identifies and mounts the root device required for managing services.

3. **Load Configuration**:
   - If the `--load` flag is specified, the application checks if the configuration has changed. If changes are detected, it applies the configuration from the specified file.
   - If the `--update` flag is specified, the application updates the services based on the default configuration file.

4. **Snapshot Management**: After loading or updating the configuration, a snapshot of the current state is created.

5. **Unmount Disk**: Finally, the application unmounts the disk after completing the operations.

## Updating Services

When the `--update` flag is used, the application reads the default configuration file and applies any necessary updates to the services defined within. If no updates are detected, it skips the snapshot management.

## Reverting Changes

Using the `--rollback` flag will trigger the application to revert to the previous state. This is useful for rolling back any changes made by the latest configuration load or update.

## Running the Application

To run the application, use the following command:

```bash
go run main.go
```

Ensure your configuration file is correctly set up, and Podman is installed and configured on your system. The application will output the status of each operation, helping you track any issues that may arise.
