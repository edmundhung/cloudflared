# cloudflared CLI manifest

This package describes cloudflared's `urfave/cli` command tree as deterministic
JSON. The manifest is intended for tools that need command discovery without
downloading or executing a cloudflared binary.

Generate a manifest with:

```console
go run ./cmd/cloudflared-manifest \
  -version "$(git describe --tags --always --match '[0-9][0-9][0-9][0-9].*.*')" \
  -output cloudflared-cli-manifest-v1.json
```

The generator reads the same command constructors used by cloudflared. It
fails when it encounters an unsupported flag type so that the published
manifest cannot silently omit CLI behavior. File output also writes a
companion `.sha256` file.

Each cloudflared release publishes both files as GitHub release assets:

```text
https://github.com/cloudflare/cloudflared/releases/download/<version>/cloudflared-cli-manifest-v1.json
https://github.com/cloudflare/cloudflared/releases/download/<version>/cloudflared-cli-manifest-v1.json.sha256
```

## Version 1

The top-level document contains:

- `schemaVersion`: version of the JSON contract.
- `cloudflaredVersion`: release represented by the manifest.
- `globalOptions`: flags registered at the application level.
- `commands`: a flattened, sorted list of commands.

Each command records its complete path, aliases, help text, positional
arguments, flags, visibility, and parser behavior. Command options are local
to that command; consumers can inherit options from parent paths and
`globalOptions`.

`argsUsage` is the source of truth for positional arguments. It uses the small
grammar `<required>`, `[optional]`, and `[variadic...]`; the generator validates
that grammar and emits the parsed shape in `arguments`. `skipFlagParsing`
marks passthrough commands, such as `cloudflared access curl`, whose remaining
arguments must not be interpreted as cloudflared flags.

Option defaults are intentionally excluded because some cloudflared defaults
contain paths derived from the machine generating the manifest. The manifest
contains environment-variable names but never reads or serializes their
values.
