package reports

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"
)

type Report struct {
	ID        int64
	Content   string
	Filename  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// AddReportToDatabase adds a new report to the database
func AddReportToDatabase(ctx context.Context, db *sql.DB, reportPath string) error {
	if err := ensureReportTable(ctx, db); err != nil {
		return err
	}

	return insertReport(ctx, db, reportPath)
}

// ensureReportTable creates the reports table if it doesn't exist
func ensureReportTable(ctx context.Context, db *sql.DB) error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS reports (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		content TEXT NOT NULL,
		filename TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := db.ExecContext(ctx, createTableSQL)
	if err != nil {
		return fmt.Errorf("error creating reports table: %w", err)
	}

	return nil
}

// insertReport handles the actual insertion of a report
func insertReport(ctx context.Context, db *sql.DB, reportPath string) error {
	content, err := os.ReadFile(reportPath)
	if err != nil {
		return fmt.Errorf("error reading report file: %w", err)
	}

	insertSQL := `
	INSERT INTO reports (content, filename, created_at, updated_at)
	VALUES (?, ?, ?, ?);`

	now := time.Now().UTC()

	result, err := db.ExecContext(ctx, insertSQL, string(content), reportPath, now, now)
	if err != nil {
		return fmt.Errorf("error inserting report into database: %w", err)
	}

	_, err = result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert ID: %w", err)
	}

	return nil
}

// GetReport retrieves a report by ID
func GetReport(ctx context.Context, db *sql.DB, id int64) (*Report, error) {
	query := `
	SELECT id, content, filename, created_at, updated_at
	FROM reports
	WHERE id = ?;`

	report := &Report{}
	err := db.QueryRowContext(ctx, query, id).Scan(
		&report.ID,
		&report.Content,
		&report.Filename,
		&report.CreatedAt,
		&report.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting report: %w", err)
	}

	return report, nil
}

// ListReports returns all reports
func ListReports(ctx context.Context, db *sql.DB) ([]Report, error) {
	query := `
	SELECT id, content, filename, created_at, updated_at
	FROM reports
	ORDER BY created_at DESC;`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error querying reports: %w", err)
	}
	defer rows.Close()

	var reports []Report
	for rows.Next() {
		var r Report
		err := rows.Scan(
			&r.ID,
			&r.Content,
			&r.Filename,
			&r.CreatedAt,
			&r.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning report row: %w", err)
		}
		reports = append(reports, r)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating report rows: %w", err)
	}

	return reports, nil
}
