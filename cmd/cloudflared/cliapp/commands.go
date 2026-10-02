package cliapp

import (
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/cloudflare/cloudflared/cmd/cloudflared/access"
	"github.com/cloudflare/cloudflared/cmd/cloudflared/cliutil"
	cfdflags "github.com/cloudflare/cloudflared/cmd/cloudflared/flags"
	"github.com/cloudflare/cloudflared/cmd/cloudflared/management"
	"github.com/cloudflare/cloudflared/cmd/cloudflared/proxydns"
	"github.com/cloudflare/cloudflared/cmd/cloudflared/tail"
	"github.com/cloudflare/cloudflared/cmd/cloudflared/tunnel"
	"github.com/cloudflare/cloudflared/cmd/cloudflared/updater"
)

// VersionText is the user-facing description of the version command and flag.
const VersionText = "Print the version"

// Commands returns cloudflared's complete top-level command tree.
func Commands(showVersion func(c *cli.Context)) []*cli.Command {
	tunnelCommands := tunnel.Commands()
	accessCommands := access.Commands()
	commands := make([]*cli.Command, 0, 2+len(tunnelCommands)+1+len(accessCommands)+1+1)
	commands = append(commands,
		&cli.Command{
			Name:   "update",
			Action: cliutil.ConfiguredAction(updater.Update),
			Usage:  "Update the agent if a new version exists",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "beta",
					Usage: "specify if you wish to update to the latest beta version",
				},
				&cli.BoolFlag{
					Name:   cfdflags.Force,
					Usage:  "specify if you wish to force an upgrade to the latest version regardless of the current version",
					Hidden: true,
				},
				&cli.BoolFlag{
					Name:   "staging",
					Usage:  "specify if you wish to use the staging url for updating",
					Hidden: true,
				},
				&cli.StringFlag{
					Name:   "version",
					Usage:  "specify a version you wish to upgrade or downgrade to",
					Hidden: false,
				},
			},
			Description: `Looks for a new version on the official download server.
If a new version exists, updates the agent binary and quits.
Otherwise, does nothing.

To determine if an update happened in a script, check for error code 11.`,
		},
		&cli.Command{
			Name: "version",
			Action: func(c *cli.Context) error {
				if c.Bool("short") {
					fmt.Println(strings.Split(c.App.Version, " ")[0])
					return nil
				}
				showVersion(c)
				return nil
			},
			Usage:       VersionText,
			Description: VersionText,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "short",
					Aliases: []string{"s"},
					Usage:   "print just the version number",
				},
			},
		},
	)
	commands = append(commands, tunnelCommands...)
	commands = append(commands, proxydns.Command()) // removed feature, only here for error message
	commands = append(commands, accessCommands...)
	commands = append(commands, tail.Command())
	commands = append(commands, management.Command())
	return commands
}

// Flags returns cloudflared's global flags.
func Flags() []cli.Flag {
	flags := tunnel.Flags()
	return append(flags, access.Flags()...)
}
