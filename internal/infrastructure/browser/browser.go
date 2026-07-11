package browser

import (
	"fmt"
	"os/exec"
	"runtime"
)

// startCommand launches a prepared command without waiting for it to exit.
// Tests replace it to observe the launch without opening a real browser.
var startCommand = func(cmd *exec.Cmd) error { return cmd.Start() }

// Open opens path with the platform's default browser and returns without
// waiting for the browser to exit.
func Open(path string) error {
	name, args := OpenCommand(runtime.GOOS, path)
	if err := startCommand(exec.Command(name, args...)); err != nil {
		return fmt.Errorf("open %s in browser: %w", path, err)
	}

	return nil
}

// OpenCommand returns the command name and arguments that open path with
// the default browser on the given platform. Unknown platforms fall back to
// the freedesktop opener.
func OpenCommand(goos, path string) (name string, args []string) {
	switch goos {
	case "darwin":
		return "open", []string{path}
	case "windows":
		return "rundll32", []string{"url.dll,FileProtocolHandler", path}
	default:
		return "xdg-open", []string{path}
	}
}
