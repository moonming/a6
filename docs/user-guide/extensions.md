# CLI Extensions

a6 supports third-party CLI extensions. An extension is an external executable that is installed locally and exposed as a top-level `a6` command.

This is separate from `a6 plugin` commands:

- `a6 plugin` manages APISIX server-side plugins
- `a6 extension` manages local CLI extensions

## Install an extension

Install from a GitHub repository:

```bash
a6 extension install <owner/repo>
```

Example:

```bash
a6 extension install api7/a6-dashboard
```

a6 fetches the latest GitHub release, selects an asset for your OS/ARCH, downloads it, and stores a manifest for later upgrades.

## List installed extensions

```bash
a6 extension list
```

Default output is table format.

You can force a format:

```bash
a6 extension list --output json
a6 extension list --output yaml
```

## Upgrade extensions

Upgrade one extension:

```bash
a6 extension upgrade <name>
```

Upgrade all installed extensions:

```bash
a6 extension upgrade --all
```

## Remove an extension

```bash
a6 extension remove <name>
```

Skip confirmation:

```bash
a6 extension remove <name> --force
```

## Where extensions are stored

Extensions are installed under:

```text
~/.config/a6/extensions/
```

Path precedence follows a6 config rules:

1. `A6_CONFIG_DIR/extensions`
2. `XDG_CONFIG_HOME/a6/extensions`
3. `~/.config/a6/extensions`

Each extension has its own directory:

```text
~/.config/a6/extensions/
└── a6-dashboard/
    ├── manifest.yaml
    └── a6-dashboard
```

Manifest example:

```yaml
name: dashboard
owner: api7
repo: a6-dashboard
version: 1.0.0
description: APISIX dashboard management
path: a6-dashboard
```

## Build your own extension

Use a repository name following the `a6-<name>` convention and publish platform assets in GitHub Releases.

Extension assets should follow the goreleaser naming convention used by `a6 extension install`:

- Linux/macOS: `a6-<name>_<version>_<os>_<arch>.tar.gz`
- Windows: `a6-<name>_<version>_<os>_<arch>.zip`

`<version>` can be either plain semver (for example `1.2.3`) or prefixed (for example `v1.2.3`).

For goreleaser, make sure the release archive includes an executable named exactly `a6-<name>` (or `a6-<name>.exe` on Windows).

When installed, the extension command appears as:

```bash
a6 <name> ...
```

All flags and arguments after `<name>` are passed through to the extension executable.
