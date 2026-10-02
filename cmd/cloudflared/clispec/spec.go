// Package clispec converts cloudflared's urfave/cli command tree into a
// deterministic, versioned JSON manifest.
package clispec

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
)

// SchemaVersion is the version of the JSON contract emitted by this package.
const SchemaVersion = 1

// Manifest describes the command-line interface exposed by one cloudflared
// release. It intentionally contains no executable behavior.
type Manifest struct {
	SchemaVersion      int       `json:"schemaVersion"`
	CloudflaredVersion string    `json:"cloudflaredVersion"`
	GlobalOptions      []Option  `json:"globalOptions,omitempty"`
	Commands           []Command `json:"commands"`
}

// Command describes one command at Path. Options contains only options declared
// directly on the command; consumers can inherit options from its parents.
type Command struct {
	Path              []string `json:"path"`
	Aliases           []string `json:"aliases,omitempty"`
	Usage             string   `json:"usage,omitempty"`
	UsageText         string   `json:"usageText,omitempty"`
	Description       string   `json:"description,omitempty"`
	ArgsUsage         string   `json:"argsUsage,omitempty"`
	Hidden            bool     `json:"hidden,omitempty"`
	SkipFlagParsing   bool     `json:"skipFlagParsing,omitempty"`
	ShortOptionGroups bool     `json:"shortOptionGroups,omitempty"`
	Options           []Option `json:"options,omitempty"`
}

// Option describes a urfave/cli flag without serializing its value. Default
// values are deliberately excluded because several cloudflared defaults depend
// on the machine generating the manifest.
type Option struct {
	Name       string   `json:"name"`
	Aliases    []string `json:"aliases,omitempty"`
	Type       string   `json:"type"`
	Repeatable bool     `json:"repeatable,omitempty"`
	Required   bool     `json:"required,omitempty"`
	Hidden     bool     `json:"hidden,omitempty"`
	Usage      string   `json:"usage,omitempty"`
	EnvVars    []string `json:"envVars,omitempty"`
}

// Build constructs a manifest from the same command and flag definitions used
// by the cloudflared executable.
func Build(version string, globalFlags []cli.Flag, commands []*cli.Command) (Manifest, error) {
	if strings.TrimSpace(version) == "" {
		return Manifest{}, fmt.Errorf("cloudflared version must not be empty")
	}

	globalOptions, err := buildOptions(globalFlags)
	if err != nil {
		return Manifest{}, fmt.Errorf("global options: %w", err)
	}

	manifest := Manifest{
		SchemaVersion:      SchemaVersion,
		CloudflaredVersion: version,
		GlobalOptions:      globalOptions,
	}
	seen := make(map[string]struct{})
	for _, command := range commands {
		if err := appendCommand(&manifest.Commands, seen, nil, command); err != nil {
			return Manifest{}, err
		}
	}
	sort.Slice(manifest.Commands, func(i, j int) bool {
		return strings.Join(manifest.Commands[i].Path, "\x00") < strings.Join(manifest.Commands[j].Path, "\x00")
	})
	return manifest, nil
}

// WriteJSON writes a human-readable manifest with a trailing newline.
func WriteJSON(w io.Writer, manifest Manifest) error {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(manifest); err != nil {
		return fmt.Errorf("encode CLI manifest: %w", err)
	}
	return nil
}

func appendCommand(target *[]Command, seen map[string]struct{}, parent []string, source *cli.Command) error {
	if source == nil {
		return fmt.Errorf("command beneath %q is nil", strings.Join(parent, " "))
	}
	if source.Name == "" {
		return fmt.Errorf("command beneath %q has no name", strings.Join(parent, " "))
	}

	path := append(append([]string(nil), parent...), source.Name)
	key := strings.Join(path, "\x00")
	if _, ok := seen[key]; ok {
		return fmt.Errorf("duplicate command path %q", strings.Join(path, " "))
	}
	seen[key] = struct{}{}

	options, err := buildOptions(source.Flags)
	if err != nil {
		return fmt.Errorf("command %q: %w", strings.Join(path, " "), err)
	}
	*target = append(*target, Command{
		Path:              path,
		Aliases:           append([]string(nil), source.Aliases...),
		Usage:             source.Usage,
		UsageText:         source.UsageText,
		Description:       source.Description,
		ArgsUsage:         source.ArgsUsage,
		Hidden:            source.Hidden,
		SkipFlagParsing:   source.SkipFlagParsing,
		ShortOptionGroups: source.UseShortOptionHandling,
		Options:           options,
	})

	for _, child := range source.Subcommands {
		if err := appendCommand(target, seen, path, child); err != nil {
			return err
		}
	}
	return nil
}

func buildOptions(flags []cli.Flag) ([]Option, error) {
	options := make([]Option, 0, len(flags))
	seen := make(map[string]struct{}, len(flags))
	for _, flag := range flags {
		option, err := buildOption(flag)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[option.Name]; ok {
			return nil, fmt.Errorf("duplicate option --%s", option.Name)
		}
		seen[option.Name] = struct{}{}
		options = append(options, option)
	}
	sort.Slice(options, func(i, j int) bool {
		return options[i].Name < options[j].Name
	})
	return options, nil
}

func buildOption(flag cli.Flag) (Option, error) {
	if flag == nil {
		return Option{}, fmt.Errorf("nil option")
	}
	base := unwrapAlternativeSourceFlag(flag)
	optionType, repeatable, err := classifyFlag(base)
	if err != nil {
		return Option{}, err
	}

	names := base.Names()
	if len(names) == 0 || names[0] == "" {
		return Option{}, fmt.Errorf("option %T has no name", flag)
	}
	documentation, ok := base.(cli.DocGenerationFlag)
	if !ok {
		return Option{}, fmt.Errorf("option --%s (%T) does not expose documentation", names[0], flag)
	}
	required := false
	if requiredFlag, ok := base.(cli.RequiredFlag); ok {
		required = requiredFlag.IsRequired()
	}
	hidden, envVars, err := flagPresentation(base)
	if err != nil {
		return Option{}, err
	}

	return Option{
		Name:       names[0],
		Aliases:    append([]string(nil), names[1:]...),
		Type:       optionType,
		Repeatable: repeatable,
		Required:   required,
		Hidden:     hidden,
		Usage:      documentation.GetUsage(),
		EnvVars:    envVars,
	}, nil
}

func unwrapAlternativeSourceFlag(flag cli.Flag) cli.Flag {
	switch flag := flag.(type) {
	case *altsrc.BoolFlag:
		return flag.BoolFlag
	case *altsrc.DurationFlag:
		return flag.DurationFlag
	case *altsrc.Float64Flag:
		return flag.Float64Flag
	case *altsrc.GenericFlag:
		return flag.GenericFlag
	case *altsrc.IntFlag:
		return flag.IntFlag
	case *altsrc.IntSliceFlag:
		return flag.IntSliceFlag
	case *altsrc.PathFlag:
		return flag.PathFlag
	case *altsrc.StringFlag:
		return flag.StringFlag
	case *altsrc.StringSliceFlag:
		return flag.StringSliceFlag
	default:
		return flag
	}
}

func classifyFlag(flag cli.Flag) (optionType string, repeatable bool, err error) {
	switch flag.(type) {
	case *cli.BoolFlag:
		return "boolean", false, nil
	case *cli.StringFlag:
		return "string", false, nil
	case *cli.StringSliceFlag:
		return "string", true, nil
	case *cli.PathFlag:
		return "path", false, nil
	case *cli.IntFlag, *cli.Int64Flag, *cli.UintFlag, *cli.Uint64Flag:
		return "integer", false, nil
	case *cli.IntSliceFlag, *cli.Int64SliceFlag:
		return "integer", true, nil
	case *cli.Float64Flag:
		return "number", false, nil
	case *cli.Float64SliceFlag:
		return "number", true, nil
	case *cli.DurationFlag:
		return "duration", false, nil
	case *cli.TimestampFlag:
		return "timestamp", false, nil
	default:
		return "", false, fmt.Errorf("unsupported option type %T", flag)
	}
}

func flagPresentation(flag cli.Flag) (hidden bool, envVars []string, err error) {
	switch flag := flag.(type) {
	case *cli.BoolFlag:
		return flag.Hidden, append([]string(nil), flag.EnvVars...), nil
	case *cli.StringFlag:
		return flag.Hidden, append([]string(nil), flag.EnvVars...), nil
	case *cli.StringSliceFlag:
		return flag.Hidden, append([]string(nil), flag.EnvVars...), nil
	case *cli.PathFlag:
		return flag.Hidden, append([]string(nil), flag.EnvVars...), nil
	case *cli.IntFlag:
		return flag.Hidden, append([]string(nil), flag.EnvVars...), nil
	case *cli.Int64Flag:
		return flag.Hidden, append([]string(nil), flag.EnvVars...), nil
	case *cli.IntSliceFlag:
		return flag.Hidden, append([]string(nil), flag.EnvVars...), nil
	case *cli.Int64SliceFlag:
		return flag.Hidden, append([]string(nil), flag.EnvVars...), nil
	case *cli.UintFlag:
		return flag.Hidden, append([]string(nil), flag.EnvVars...), nil
	case *cli.Uint64Flag:
		return flag.Hidden, append([]string(nil), flag.EnvVars...), nil
	case *cli.Float64Flag:
		return flag.Hidden, append([]string(nil), flag.EnvVars...), nil
	case *cli.Float64SliceFlag:
		return flag.Hidden, append([]string(nil), flag.EnvVars...), nil
	case *cli.DurationFlag:
		return flag.Hidden, append([]string(nil), flag.EnvVars...), nil
	case *cli.TimestampFlag:
		return flag.Hidden, append([]string(nil), flag.EnvVars...), nil
	default:
		return false, nil, fmt.Errorf("unsupported option presentation type %T", flag)
	}
}
