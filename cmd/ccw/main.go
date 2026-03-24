package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

func main() {
	// Read stdin with 1-second timeout
	done := make(chan []byte, 1)
	go func() {
		data, _ := io.ReadAll(os.Stdin)
		done <- data
	}()

	var data []byte
	select {
	case data = <-done:
	case <-time.After(1 * time.Second):
		os.Exit(0)
	}

	if len(data) == 0 {
		os.Exit(0)
	}

	var input protocol.StatusLineInput
	if err := json.Unmarshal(data, &input); err != nil {
		os.Exit(0)
	}

	fmt.Print("ccw: ok")
}
