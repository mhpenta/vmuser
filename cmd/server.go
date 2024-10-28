package cmd

import (
	"context"
	"log/slog"
	"vmuser/config"
	"vmuser/server"
)

func Server(appCtx context.Context, cfg *config.VMUserConfig) error {
	serverCfg := server.Config{
		Port: cfg.Server.Port,
	}
	s := server.NewServer(&serverCfg)

	err := s.Start(appCtx)
	if err != nil {
		slog.Error("Error starting server", "err", err)
	}
	return err
}
