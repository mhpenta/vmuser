package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"vmuser/cmd"
	"vmuser/config"
)

func main() {
	configFile := flag.String("config", "vmuser.toml", "Path to the configuration file")
	tui := flag.Bool("tui", false, "Run TUI")
	addReport := flag.String("add-report", "", "Path to the report file to add")

	flag.Parse()

	appContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill, syscall.SIGTERM)
	defer stop()

	cfg := config.GetVMUserConfig(*configFile)

	// Handle add-report command first
	if *addReport != "" {
		if err := cmd.AddReport(appContext, cfg, *addReport); err != nil {
			slog.Error("Error adding report", "error", err, "file", *addReport)
			os.Exit(1)
		}
		return
	}

	if *tui {
		if err := cmd.TUI(appContext, cfg); err != nil {
			slog.Error("Error running application", "error", err)
			os.Exit(1)
		}
	}

	if err := cmd.Server(appContext, cfg); err != nil {
		slog.Error("Error running application", "error", err)
		os.Exit(1)
	}
}
