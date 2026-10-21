# container-service-api

Go tooling that runs AhaOS system services as Podman containers instead of
native daemons. Service definitions (image, systemd unit template, privileged
flag) are declared in JSON and applied by two binaries built from this single
codebase:

- **`default-config-init`** — installer-phase provisioning of the default
  service set.
- **`service-manager`** — runtime CLI to load a configuration, update images,
  and roll back to a previous snapshot.

The project replaces daemons such as CUPS, PackageKit, UPower, UDisks2, fwupd,
bluetooth, samba, colord, and wpa_supplicant with containerized equivalents
while keeping their systemd units and D-Bus interfaces intact.

## Purpose and functionality

### Installer phase — `default-config-init`

For each service in the embedded `default-config.json` (15 services):

1. Pull `docker.io/ahaosv1/<service>` with `podman`.
2. Disable and mask the distribution's native unit.
3. Write `<service>-backend.service` from the unit-file template, running
   `podman run --name <service>-service-backend …`.
4. Enable and start the service with `systemctl`.

### Runtime — `service-manager`

| Command | Behaviour |
|---|---|
| `--load <file>` | Applies `service.<name>.enable = true\|false` entries. Pulls missing images inside an overlay/chroot, writes/removes units, then creates a Btrfs snapshot. |
| `--update` | Compares running container image IDs with the registry (skopeo); pulls and restarts only services whose image changed, then snapshots. |
| `--rollback` | Reverts the system to the previous Btrfs snapshot. |

`--load` and `--update` are mutually exclusive. The CLI mounts the root device
on `/mnt` before modifying anything and unmounts it afterwards.

Configuration file format (`/etc/service-manager/configuration.conf`):

```ini
service.cups.enable = true
service.bluetooth.enable = false
```

## Codebase organization

```
.
├── main.go                 # entrypoint: main() calls run()
├── main_default.go         # //go:build default — installer-phase run()
├── main_load.go            # //go:build load    — runtime CLI run()
├── config/
│   ├── config.go           # loads embedded JSON (//go:embed), no external files
│   ├── default-config.json # default service set applied at install time
│   └── registry-services.json # unit-file templates for all known services
├── handlers/
│   ├── common.go           # shared disable/mask logic (used by both builds)
│   ├── default_handlers.go # [default] pull, create unit, enable + start
│   ├── load_handlers.go    # [load]    enable/disable, chroot pull, unit placement
│   ├── update_handlers.go  # [load]    image diff + restart
│   ├── snapshot_handlers.go# [load]    Btrfs snapshot creation / change detection
│   └── rollback_handlers.go# [load]    mount root device, revert snapshot
├── utils/                  # command execution and env loading helpers
├── Docs/                   # per-area notes: default config, load config, handlers
└── build/                  # local build output (gitignored)
```

Key modules:

- **`config`** — embeds `default-config.json` and `registry-services.json` into
  the binary and exposes `GetConfig()` / `GetRegistryTemplates()`.
- **`handlers`** — all side effects (Podman, systemd, skopeo, Btrfs); split by
  build tag so each binary only links the code it needs.
- **`utils`** — thin wrapper around `exec.Command` returning structured
  `CommandResponse{Output, Error}`.

## Dependencies and integrations

**Build dependencies:** Go 1.22.5 (`module go-podman-api`), one Go library:
`github.com/joho/godotenv`.

**Runtime dependencies on the target system:** Podman, systemd, skopeo
(image digest checks), a Btrfs root for snapshots, and root privileges
(systemd unit management and container lifecycle).

**External services:** Docker Hub organization `ahaosv1` (public image
registry for the service images) and `salsa.debian.org` (upstream package
sources used during image builds). No cloud API keys or project IDs are
involved.

## Live product integration

**Integrated** This runs inside **AhaOS**:

- `default-config-init` executes during the **installer phase**, so the
  containerized service set exists on a freshly installed system.
- `service-manager` is shipped for runtime administration and is driven by the
  AhaOS Settings app for enabling/disabling services, updates, and rollback.

It is not a network service and has no hosted deployment: everything runs
on-device with root privileges. Status is live in the OS build; the code in
this repository is the source of truth for the installed behaviour.

## Build and deployment

The repository contains **no build automation** — no Makefile, CI workflow, or
Debian packaging. Binaries are produced manually and `build/` is gitignored.
Both builds require an explicit tag, because `main.go` only calls `run()` and
each tagged file provides it:

```bash
# Runtime CLI (--load, --update, --rollback)
go build -tags load -o build/service-manager .

# Installer-phase provisioning
go build -tags default -o build/default-config-init .
```

Building without a tag fails with `undefined: run`.

**Deployment:** the AhaOS image build installs both binaries and invokes
`default-config-init` during installation; `service-manager` remains available
on the installed system. Unit files are not shipped in the repo — they are
generated at runtime into `/etc/systemd/system/` and enabled via `systemctl`.
Because the build recipe lives outside this repository, the deployed artifacts
cannot be reproduced from the code here alone.

**Maintenance:** `main` is the default branch, `development` carries integrated
work, and feature work happens on topic branches (for example
`service-config-integration`). Additional design notes live in `Docs/`.
