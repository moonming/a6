# Auto-Update

a6 can keep itself up to date in two ways:

- Explicit update command: `a6 update`
- Background update check on normal command startup

## Update a6 explicitly

Run:

```bash
a6 update
```

This command:

1. Prints the current version
2. Fetches the latest release from GitHub (`api7/a6`)
3. Compares versions
4. Downloads the release asset for your current OS/ARCH
5. Replaces the current `a6` binary

Use `--force` to skip confirmation prompts:

```bash
a6 update --force
```

If your current build is `dev`, a6 warns that version comparison may be inaccurate and asks whether to continue.

## Background update check

After command execution, a6 checks a cached update state and may show a notice on stderr:

```text
A new version of a6 is available: v1.2.3 → v1.3.0
Run 'a6 update' to update.
```

Behavior details:

- Never blocks command execution
- Uses cached state from `~/.config/a6/update-check.json` (or configured config dir)
- Refreshes cache at most once every 24 hours
- Skips checks for `update`, `version`, and `completion` commands
- Skips checks when stdout is not a TTY
- Skips checks in CI environments

## Disable update checks

Set `A6_NO_UPDATE_CHECK` to any non-empty value:

```bash
export A6_NO_UPDATE_CHECK=1
```

## Manual installation alternative

If automatic replacement fails due to permissions, install manually:

1. Download the correct asset from GitHub releases
2. Extract the binary
3. Replace your installed `a6` binary with appropriate permissions
