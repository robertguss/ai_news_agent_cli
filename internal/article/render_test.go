package article

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderMarkdown(t *testing.T) {
	content := "# Test Article\n\nThis is **bold** text."

	t.Run("plain text rendering", func(t *testing.T) {
		var buf bytes.Buffer
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := RenderMarkdown(content, false, w)
		require.NoError(t, err)

		w.Close()
		os.Stdout = oldStdout

		_, _ = buf.ReadFrom(r)
		output := buf.String()
		assert.Equal(t, content, output)
	})

	t.Run("styled rendering fallback to glamour", func(t *testing.T) {
		var buf bytes.Buffer
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := RenderMarkdown(content, true, w)
		require.NoError(t, err)

		w.Close()
		os.Stdout = oldStdout

		_, _ = buf.ReadFrom(r)
		output := buf.String()

		assert.NotEmpty(t, output)
		assert.NotEqual(t, content, output)
	})
}

func TestTryGlowCLI(t *testing.T) {
	content := "# Test\n\nContent"

	t.Run("glow not available", func(t *testing.T) {
		originalPath := os.Getenv("PATH")
		os.Setenv("PATH", "")
		defer os.Setenv("PATH", originalPath)

		err := tryGlowCLI(content, &bytes.Buffer{})
		assert.Error(t, err)
	})

	t.Run("glow available", func(t *testing.T) {
		if _, err := exec.LookPath("glow"); err != nil {
			t.Skip("glow not available in PATH")
		}

		err := tryGlowCLI(content, &bytes.Buffer{})
		assert.NoError(t, err)
	})
}

func TestRenderWithGlamour(t *testing.T) {
	content := "# Test Article\n\nThis is **bold** text with [link](https://example.com)."

	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := renderWithGlamour(content, w)
	require.NoError(t, err)

	w.Close()
	os.Stdout = oldStdout

	_, _ = buf.ReadFrom(r)
	output := buf.String()

	assert.NotEmpty(t, output)
	assert.NotEqual(t, content, output)
	assert.Contains(t, strings.ToLower(output), "test article")
}
