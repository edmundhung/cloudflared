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
manifest cannot silently omit CLI behavior.

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

`argsUsage` preserves the free-form text supplied to `urfave/cli`. The
`arguments` array separately records each positional argument's name, help
text, requiredness, and whether it is variadic. `skipFlagParsing` marks
passthrough commands, such as `cloudflared access curl`, whose remaining
arguments must not be interpreted as cloudflared flags.

Option defaults are intentionally excluded because some cloudflared defaults
contain paths derived from the machine generating the manifest. The manifest
contains environment-variable names but never reads or serializes their
values.
