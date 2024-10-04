# Podman Service Manager

## Overview

The Podman Service Manager is a Go application designed to manage containerized services using the Podman API. It provides functionalities for initializing configurations, managing services, rolling back to previous states, handling snapshots, and updating container images. This project leverages systemd for service management and Btrfs for efficient state management.

## Table of Contents

- [Requirements](#requirements)
- [Installation](#installation)
- [Configuration](#configuration)
- [Functionality](#functionality)
  - [Initialization](#initialization)
  - [Service Management](#service-management)
  - [Rollback Functionality](#rollback-functionality)
  - [Snapshot Management](#snapshot-management)
  - [Service Updates](#service-updates)
- [Usage](#usage)
- [Error Handling](#error-handling)
- [Contributing](#contributing)
- [License](#license)

## Requirements

To run this application, ensure you have the following installed on your system:

- Go (version 1.16 or later)
- Podman installed and configured
- Skopeo for inspecting remote images
- Btrfs filesystem for snapshot management

## Installation

1. Clone the repository:

   ```bash
   git clone <repository-url>
   cd podman-service-manager
   ```

2. Build the application:

   ```bash
   go build -o service-manager main.go
   ```

3. Run the application:

   ```bash
   ./service-manager
   ```

## Configuration

The application requires a configuration file to manage services, specified in the format:

```plaintext
service.<service-name>.enable = true  # or false
```

### Example Configuration File

```plaintext
service.service1.enable = true
service.service2.enable = false
```

This configuration file should be placed at `/etc/service-manager/configuration.conf` by default. Ensure that the application has the necessary permissions to read from this path.

## Functionality

### Initialization

The application initializes services by loading the configuration, pulling images, creating systemd unit files, and starting the services. The initialization process ensures that all specified services are up and running.

### Service Management

The application provides handlers for managing services, including:

- **Pulling images** from a registry.
- **Creating and placing systemd unit files** for services.
- **Enabling and starting services** with systemd.
- **Checking and disabling existing services** if necessary.

### Rollback Functionality

The rollback feature allows the application to revert to a previous state using Btrfs snapshots. It can:

- Read the snapshot directory.
- Identify the current and previous snapshots.
- Move the filesystem state back to the previous snapshot, restoring the system to its earlier state.

### Snapshot Management

The application efficiently manages snapshots with the following functionalities:

- **Creating new snapshots** of the current state.
- **Checking if the configuration has changed** since the last snapshot.
- **Fetching existing snapshots** for rollback purposes.

### Service Updates

The application can check and update services to ensure they are using the latest container images. It performs the following steps:

- Compares the current image digest with the remote image digest.
- Pulls the latest image if the digests do not match.
- Restarts the service after updating to the latest image.

## Usage

To manage services, run the application with the following command:

```bash
./service-manager --load /etc/service-manager/configuration.conf
```

or

```bash
./service-manager --update
```

### Important Note on Flags

- You **cannot** use multiple command-line flags at the same time. Only one flag can be specified in each execution of the application. If you attempt to use more than one flag, the application will display an error message and exit.

## Error Handling

The application provides structured error messages for debugging and issue resolution. Each function returns an error if an operation fails, allowing for better error handling in the application.

---
