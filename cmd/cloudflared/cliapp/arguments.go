package cliapp

import "github.com/cloudflare/cloudflared/cmd/cloudflared/clispec"

// CommandArguments returns structured positional argument annotations for the
// command tree. cloudflared's urfave/cli fork only models positional arguments
// as free-form help text, so the manifest validates and merges this companion
// metadata by command path.
func CommandArguments() []clispec.CommandArguments {
	return []clispec.CommandArguments{
		{
			Path: []string{"access", "login"},
			Arguments: []clispec.Argument{
				{Name: "url", Usage: "URL of the Access application"},
			},
		},
		{
			Path: []string{"access", "curl"},
			Arguments: []clispec.Argument{
				{Name: "url", Usage: "URL of the Access application", Required: true},
				{Name: "curl-args", Usage: "Arguments forwarded to curl", Variadic: true},
			},
		},
		{
			Path: []string{"access", "token"},
			Arguments: []clispec.Argument{
				{Name: "url", Usage: "URL of the Access application"},
			},
		},
		{
			Path: []string{"management", "token"},
			Arguments: []clispec.Argument{
				{Name: "tunnel-id", Usage: "Tunnel UUID", Required: true},
			},
		},
		{
			Path: []string{"tail"},
			Arguments: []clispec.Argument{
				{Name: "tunnel-id", Usage: "Tunnel UUID to stream logs from"},
			},
		},
		{
			Path: []string{"tail", "token"},
			Arguments: []clispec.Argument{
				{Name: "tunnel-id", Usage: "Tunnel UUID", Required: true},
			},
		},
		{
			Path: []string{"tunnel", "cleanup"},
			Arguments: []clispec.Argument{
				{Name: "tunnel", Usage: "Tunnel name or UUID", Required: true, Variadic: true},
			},
		},
		{
			Path: []string{"tunnel", "create"},
			Arguments: []clispec.Argument{
				{Name: "name", Usage: "Name of the tunnel", Required: true},
			},
		},
		{
			Path: []string{"tunnel", "delete"},
			Arguments: []clispec.Argument{
				{Name: "tunnel", Usage: "Tunnel name or UUID", Required: true, Variadic: true},
			},
		},
		{
			Path: []string{"tunnel", "info"},
			Arguments: []clispec.Argument{
				{Name: "tunnel", Usage: "Tunnel name or UUID", Required: true},
			},
		},
		{
			Path: []string{"tunnel", "ingress", "rule"},
			Arguments: []clispec.Argument{
				{Name: "url", Usage: "URL to test against the ingress rules", Required: true},
			},
		},
		{
			Path: []string{"tunnel", "route", "dns"},
			Arguments: []clispec.Argument{
				{Name: "tunnel", Usage: "Tunnel name or UUID", Required: true},
				{Name: "hostname", Usage: "Hostname to route", Required: true},
			},
		},
		{
			Path: []string{"tunnel", "route", "ip", "add"},
			Arguments: []clispec.Argument{
				{Name: "cidr", Usage: "Private network in CIDR notation", Required: true},
				{Name: "tunnel", Usage: "Tunnel name or UUID", Required: true},
				{Name: "comment", Usage: "Comment describing the route"},
			},
		},
		{
			Path: []string{"tunnel", "route", "ip", "delete"},
			Arguments: []clispec.Argument{
				{Name: "route", Usage: "Route UUID or CIDR", Required: true},
			},
		},
		{
			Path: []string{"tunnel", "route", "ip", "get"},
			Arguments: []clispec.Argument{
				{Name: "ip", Usage: "IP address whose matching route should be returned", Required: true},
			},
		},
		{
			Path: []string{"tunnel", "route", "lb"},
			Arguments: []clispec.Argument{
				{Name: "tunnel", Usage: "Tunnel name or UUID", Required: true},
				{Name: "hostname", Usage: "Load balancer hostname", Required: true},
				{Name: "pool", Usage: "Load balancer pool name", Required: true},
			},
		},
		{
			Path: []string{"tunnel", "run"},
			Arguments: []clispec.Argument{
				{Name: "tunnel", Usage: "Tunnel name or UUID"},
			},
		},
		{
			Path: []string{"tunnel", "token"},
			Arguments: []clispec.Argument{
				{Name: "tunnel", Usage: "Tunnel name or UUID", Required: true},
			},
		},
		{
			Path: []string{"tunnel", "vnet", "add"},
			Arguments: []clispec.Argument{
				{Name: "name", Usage: "Name of the virtual network", Required: true},
				{Name: "comment", Usage: "Comment describing the virtual network"},
			},
		},
		{
			Path: []string{"tunnel", "vnet", "delete"},
			Arguments: []clispec.Argument{
				{Name: "virtual-network", Usage: "Virtual network name or UUID", Required: true},
			},
		},
		{
			Path: []string{"tunnel", "vnet", "update"},
			Arguments: []clispec.Argument{
				{Name: "virtual-network", Usage: "Virtual network name or UUID", Required: true},
			},
		},
	}
}
