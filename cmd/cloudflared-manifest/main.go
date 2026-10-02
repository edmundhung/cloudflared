// cloudflared-manifest emits a versioned JSON description of cloudflared's
// command-line interface. It is a release-time build tool, not part of the
// cloudflared binary.
package main

import (
	"bytes"
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

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
	checksum := sha256.Sum256(contents.Bytes())
	checksumOutput := output + ".sha256"
	checksumContents := fmt.Appendf(nil, "%x  %s\n", checksum, filepath.Base(output))
	if err := os.WriteFile(checksumOutput, checksumContents, 0o600); err != nil {
		return fmt.Errorf("write manifest checksum %q: %w", checksumOutput, err)
	}
	return nil
}
