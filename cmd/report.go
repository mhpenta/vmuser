package cmd

import (
        "context"
        "database/sql"
        "fmt"
        "os"
        "text/tabwriter"
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

// GetReportByID retrieves a specific report by its ID
func GetReportByID(ctx context.Context, cfg *config.VMUserConfig, id int64) (*reports.Report, error) {
        db, err := database.GetConnection(&cfg.Turso)
        if err != nil {
                return nil, fmt.Errorf("error getting database connection: %w", err)
        }

        report, err := reports.GetReport(ctx, db, id)
        if err != nil {
                if err == sql.ErrNoRows {
                        return nil, fmt.Errorf("report with ID %d not found", id)
                }
                return nil, fmt.Errorf("error retrieving report: %w", err)
        }

        return report, nil
}

// ListAllReports retrieves all reports from the database
func ListAllReports(ctx context.Context, cfg *config.VMUserConfig) ([]reports.Report, error) {
        db, err := database.GetConnection(&cfg.Turso)
        if err != nil {
                return nil, fmt.Errorf("error getting database connection: %w", err)
        }

        reportList, err := reports.ListReports(ctx, db)
        if err != nil {
                return nil, fmt.Errorf("error retrieving reports: %w", err)
        }

        return reportList, nil
}

// DisplayReport formats and prints a single report
func DisplayReport(w *tabwriter.Writer, report *reports.Report) {
        fmt.Fprintf(w, "Report ID:\t%d\n", report.ID)
        fmt.Fprintf(w, "Filename:\t%s\n", report.Filename)
        fmt.Fprintf(w, "Created At:\t%s\n", report.CreatedAt.Format("2006-01-02 15:04:05"))
        fmt.Fprintf(w, "Updated At:\t%s\n", report.UpdatedAt.Format("2006-01-02 15:04:05"))
        fmt.Fprintf(w, "Content:\n%s\n", report.Content)
}
