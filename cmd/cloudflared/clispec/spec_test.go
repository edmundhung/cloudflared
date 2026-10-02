package clispec_test

import (
	"bytes"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"

	"github.com/cloudflare/cloudflared/cmd/cloudflared/cliapp"
	"github.com/cloudflare/cloudflared/cmd/cloudflared/clispec"
)

func TestBuildManifest(t *testing.T) {
	t.Parallel()

	manifest, err := clispec.Build("2026.9.3", []cli.Flag{
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:    "config",
			Aliases: []string{"c"},
			Usage:   "Configuration file",
			EnvVars: []string{"TUNNEL_CONFIG"},
		}),
	}, []*cli.Command{
		{
			Name:    "access",
			Aliases: []string{"forward"},
			Usage:   "Access commands",
			Subcommands: []*cli.Command{
				{
					Name:            "curl",
					ArgsUsage:       "<url> [<curl args>...]",
					SkipFlagParsing: true,
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:     "allow-request",
							Aliases:  []string{"ar"},
							Usage:    "Continue without a token",
							Required: true,
						},
						altsrc.NewStringSliceFlag(&cli.StringSliceFlag{
							Name:    "header",
							Usage:   "Additional request header",
							EnvVars: []string{"TUNNEL_HEADERS"},
						}),
					},
				},
			},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, clispec.SchemaVersion, manifest.SchemaVersion)
	assert.Equal(t, "2026.9.3", manifest.CloudflaredVersion)
	require.Len(t, manifest.GlobalOptions, 1)
	assert.Equal(t, clispec.Option{
		Name:    "config",
		Aliases: []string{"c"},
		Type:    "string",
		Usage:   "Configuration file",
		EnvVars: []string{"TUNNEL_CONFIG"},
	}, manifest.GlobalOptions[0])

	require.Len(t, manifest.Commands, 2)
	assert.Equal(t, []string{"access"}, manifest.Commands[0].Path)
	curl := manifest.Commands[1]
	assert.Equal(t, []string{"access", "curl"}, curl.Path)
	assert.Equal(t, "<url> [<curl args>...]", curl.ArgsUsage)
	assert.True(t, curl.SkipFlagParsing)
	assert.Equal(t, []clispec.Option{
		{
			Name:     "allow-request",
			Aliases:  []string{"ar"},
			Type:     "boolean",
			Required: true,
			Usage:    "Continue without a token",
		},
		{
			Name:       "header",
			Type:       "string",
			Repeatable: true,
			Usage:      "Additional request header",
			EnvVars:    []string{"TUNNEL_HEADERS"},
		},
	}, curl.Options)
}

func TestCloudflaredCommandTreeCanBeSerialized(t *testing.T) {
	t.Parallel()

	manifest, err := clispec.Build(
		"2026.9.3",
		cliapp.Flags(),
		cliapp.Commands(func(*cli.Context) {}),
	)
	require.NoError(t, err)
	assert.Greater(t, len(manifest.Commands), 20)

	tunnelDiag := findCommand(t, manifest, "tunnel", "diag")
	assert.NotEmpty(t, tunnelDiag.Options)
	accessCurl := findCommand(t, manifest, "access", "curl")
	assert.True(t, accessCurl.SkipFlagParsing)
	assert.Equal(t, "forward", findCommand(t, manifest, "access").Aliases[0])
}

func TestWriteJSONIsDeterministic(t *testing.T) {
	t.Parallel()

	manifest, err := clispec.Build(
		"2026.9.3",
		cliapp.Flags(),
		cliapp.Commands(func(*cli.Context) {}),
	)
	require.NoError(t, err)

	var first bytes.Buffer
	require.NoError(t, clispec.WriteJSON(&first, manifest))
	var second bytes.Buffer
	require.NoError(t, clispec.WriteJSON(&second, manifest))
	assert.Equal(t, first.String(), second.String())
	assert.Contains(t, first.String(), `"schemaVersion": 1`)
	assert.NotContains(t, first.String(), "Destination")
}

func TestBuildRejectsUnsupportedFlags(t *testing.T) {
	t.Parallel()

	_, err := clispec.Build("2026.9.3", []cli.Flag{
		&cli.GenericFlag{Name: "unsupported"},
	}, nil)
	require.ErrorContains(t, err, "unsupported option type")
}

func findCommand(t *testing.T, manifest clispec.Manifest, path ...string) clispec.Command {
	t.Helper()
	for _, command := range manifest.Commands {
		if slices.Equal(path, command.Path) {
			return command
		}
	}
	require.FailNow(t, "command not found", "path: %v", path)
	return clispec.Command{}
}
