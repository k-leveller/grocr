package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kevin/grocy-scanner/api"
	"github.com/kevin/grocy-scanner/config"
	"github.com/kevin/grocy-scanner/logger"
)

func main() {
	testMode := flag.Bool("test", false, "Test mode: skip all Grocy API calls")
	consumeMode := flag.Bool("consume", false, "Start in consume mode")
	flag.Parse()

	if err := logger.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not initialize syslog: %v\n", err)
	}
	defer logger.Close()

	var grocy *api.GrocyClient
	if !*testMode {
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
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
