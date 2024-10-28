package cmd

import (
	"context"
	"github.com/charmbracelet/huh"
	"log/slog"
	"os"
	"vmuser/config"
)

func TUI(appCtx context.Context, cfg *config.VMUserConfig) error {
	var function string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("XBRL-Go").Description("Select an option").
				Options(
					huh.NewOption("Home", "home"),
					huh.NewOption("Start server", "server"),
				).
				Value(&function),
		),
	).WithTheme(huh.ThemeBase16())

	err := form.Run()
	if err != nil {
		slog.Error("Error running form", "error", err)
		os.Exit(1)
	}

	switch function {
	case "home":
		slog.Info("Displaying home")
	case "server":
		err = Server(appCtx, cfg)
		slog.Error("Error starting server", "error", err)
	case "exit":
		slog.Info("Exiting application")
		os.Exit(0)
	default:
		slog.Error("No valid option selected")
	}
	return nil
}
