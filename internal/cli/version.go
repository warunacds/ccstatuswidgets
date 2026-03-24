package cli

import "fmt"

// Version is set at build time via -ldflags. Falls back to "dev".
var Version = "dev"

// RunVersion prints the version string.
func RunVersion() {
	fmt.Printf("ccstatuswidgets %s\n", Version)
}
