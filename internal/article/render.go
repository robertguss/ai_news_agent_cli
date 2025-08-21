package article

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/glamour"
)

func RenderMarkdown(content string, styled bool, writer io.Writer) error {
	if !styled {
		fmt.Fprint(writer, content)
		return nil
	}

	// Skip glow CLI for now to avoid potential hanging issues
	// if err := tryGlowCLI(content, writer); err == nil {
	//	return nil
	// }

	return renderWithGlamour(content, writer)
}

func tryGlowCLI(content string, writer io.Writer) error {
	if _, err := exec.LookPath("glow"); err != nil {
		return err
	}

	cmd := exec.Command("glow", "-")
	cmd.Stdin = strings.NewReader(content)
	cmd.Stdout = writer
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func renderWithGlamour(content string, writer io.Writer) error {
	// Use simpler, faster Glamour settings
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"), // Use a predefined style instead of auto-detection
		glamour.WithWordWrap(80),
	)
	if err != nil {
		fmt.Fprint(writer, content)
		return nil
	}

	rendered, err := renderer.Render(content)
	if err != nil {
		fmt.Fprint(writer, content)
		return nil
	}

	fmt.Fprint(writer, rendered)
	return nil
}
