# Get Report Feature Design

## Overview
Add functionality to retrieve reports from the database, both individual reports by ID and listing all reports.

## Current State
- The database layer already has `GetReport` and `ListReports` functions implemented
- The CLI only exposes the `add-report` functionality
- Reports are stored with content, filename, and timestamps

## Implementation Plan

### 1. Command Line Interface
Add two new commands:
- `vmuser get-report <id>`: Retrieve a specific report by ID
- `vmuser list-reports`: List all available reports

### 2. Command Layer Implementation
Create new functions in `cmd/report.go`:

```go
func GetReportByID(ctx context.Context, cfg *config.VMUserConfig, id int64) (*reports.Report, error)
func ListAllReports(ctx context.Context, cfg *config.VMUserConfig) ([]reports.Report, error)
```

### 3. Output Formatting
For `get-report`:
- Display report ID, filename, and content
- Show creation and update timestamps
- Option to save content to a file

For `list-reports`:
- Display a table with columns: ID, Filename, Created At
- Omit content for brevity
- Sort by creation date (newest first)

### 4. Error Handling
- Handle "report not found" scenarios
- Validate input ID format
- Handle database connection errors
- Provide user-friendly error messages

### 5. Testing
- Add unit tests for new command functions
- Add integration tests with the database
- Test edge cases (invalid IDs, empty database)

## Usage Examples

```bash
# Get a specific report
vmuser get-report 123

# List all reports
vmuser list-reports

# Get a report and save to file
vmuser get-report 123 --output report.txt
```

## Future Enhancements
- Add filtering options for list-reports (by date range, filename)
- Add search functionality
- Add output format options (JSON, CSV)
- Add pagination for list-reports