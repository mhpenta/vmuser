package cmd

import (
	"context"
	"fmt"
	"os"
	"vmuser/config"
	"vmuser/database"
	"vmuser/pkg/reports"
)

func AddReport(ctx context.Context, cfg *config.VMUserConfig, filePath string) error {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("report file does not exist: %s", filePath)
	}

	db, err := database.GetConnection(&cfg.Turso)
	if err != nil {
		return fmt.Errorf("error getting database connection: %w", err)
	}

	err = reports.AddReportToDatabase(ctx, db, filePath)
	if err != nil {
		return fmt.Errorf("error adding report to database: %w", err)
	}

	return nil
}
