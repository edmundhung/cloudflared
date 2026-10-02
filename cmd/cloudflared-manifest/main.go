// cloudflared-manifest emits a versioned JSON description of cloudflared's
// command-line interface. It is a release-time build tool, not part of the
// cloudflared binary.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"

	"github.com/cloudflare/cloudflared/cmd/cloudflared/cliapp"
	"github.com/cloudflare/cloudflared/cmd/cloudflared/clispec"
)

func main() {
	version := flag.String("version", "", "cloudflared release version represented by the manifest")
	output := flag.String("output", "-", "output path, or - for stdout")
	flag.Parse()

	if err := generate(*version, *output, os.Stdout); err != nil {
		log.Fatal().Err(err).Msg("Failed to generate cloudflared CLI manifest")
	}
}

func generate(version string, output string, stdout io.Writer) error {
	manifest, err := clispec.Build(
		version,
		cliapp.Flags(),
		cliapp.Commands(func(*cli.Context) {}),
		cliapp.CommandArguments(),
	)
	if err != nil {
		return err
	}
	if output == "-" {
		return clispec.WriteJSON(stdout, manifest)
	}

	var contents bytes.Buffer
	if err := clispec.WriteJSON(&contents, manifest); err != nil {
		return err
	}
	if err := os.WriteFile(output, contents.Bytes(), 0o600); err != nil {
		return fmt.Errorf("write manifest output %q: %w", output, err)
	}
	return nil
}
