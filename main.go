package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/k-leveller/grocr/api"
	"github.com/k-leveller/grocr/config"
	"github.com/k-leveller/grocr/keybind"
	"github.com/k-leveller/grocr/locale"
	"github.com/k-leveller/grocr/logger"
)

func main() {
	testMode := flag.Bool("test", false, "Test mode: skip all Grocy API calls")
	consumeMode := flag.Bool("consume", false, "Start in consume mode")
	flag.Parse()

	if err := logger.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not initialize syslog: %v\n", err)
	}
	defer logger.Close()

	// Keybinds always load: a missing or corrupt file falls back to the
	// built-in defaults.
	keybind.Load()

	var grocy *api.GrocyClient
	if !*testMode {
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		locale.Load(cfg.Language)
		grocy = api.NewGrocyClient(cfg)
	} else {
		// Dummy client for test mode
		grocy = api.NewGrocyClient(&config.Config{BaseURL: "http://localhost", APIKey: "test"})
	}

	off := api.NewOFFClient()
	m := NewModel(grocy, off, *testMode)
	if *consumeMode {
		m.mode = "consume"
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
