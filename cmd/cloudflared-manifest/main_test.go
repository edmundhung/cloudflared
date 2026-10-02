package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cloudflare/cloudflared/cmd/cloudflared/clispec"
)

func TestGenerateToStdout(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	require.NoError(t, generate("2026.9.3", "-", &output))

	var manifest clispec.Manifest
	require.NoError(t, json.Unmarshal(output.Bytes(), &manifest))
	assert.Equal(t, "2026.9.3", manifest.CloudflaredVersion)
	assert.NotEmpty(t, manifest.Commands)
}

func TestGenerateToFile(t *testing.T) {
	t.Parallel()

	output := filepath.Join(t.TempDir(), "cloudflared-cli-manifest-v1.json")
	require.NoError(t, generate("2026.9.3", output, nil))

	info, err := os.Stat(output)
	require.NoError(t, err)
	assert.Positive(t, info.Size())
}
