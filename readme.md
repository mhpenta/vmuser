# VMUser Repository

A Go application that provides a configurable server with CLI capabilities, report management, and a TUI (Terminal User Interface) mode. The application uses SQLite/Turso for data storage and includes robust HTTP client utilities.

## Features

### Core Functionality
- **Server Mode**: HTTP server with configurable ports and API endpoints
- **TUI Mode**: Terminal-based user interface for interactive usage
- **Report Management**: Add and manage reports with database storage
- **Configuration Management**: TOML-based configuration with environment variable support

### Command Line Interface
```bash
# Run server mode
go run . 

# Run TUI mode
go run . --tui

# Add a report
go run . --add-report path/to/report.md

# Specify config file
go run . --config custom_config.toml
```

### Configuration
Supports configuration for multiple services and components:
- Elastic Search
- PostgreSQL Database
- Turso Database
- Server Settings
- LLM Integration
- Custom LLM Library Settings

## Project Structure

### Core Packages
- `cmd/`: Command handlers for different modes (server, TUI, report management)
- `config/`: Configuration structs and loading logic
- `database/`: Database connection and virtual filesystem implementation
- `pkg/reports/`: Report management and storage
- `server/`: HTTP server implementation

### Extended HTTP Utilities (`ext/httpext/`)
- **Headers**: Predefined HTTP headers for various services
- **Requests**: Sophisticated HTTP client implementations with:
    - Retry logic
    - Rate limiting
    - Error handling
    - Redirect following
    - Special SEC API handling
- **Responses**: Response helpers for:
    - JSON responses
    - HTML responses
    - Text responses
    - Server-Sent Events (SSE)
    - Error responses

### Virtual Filesystem
Implements a database-backed virtual filesystem with features:
- Basic file operations (Create, Read, Update, Delete)
- Directory operations
- Search capabilities
- Metadata management
- Security limits and constraints

## Technical Details

### Database Schema
The application uses SQLite/Turso with the following main tables:
- `reports`: Stores report content and metadata
- `system_config`: System configuration storage
- `virtual_filesystem`: Virtual file system storage
- `operation_log`: Logging of system operations

### HTTP Client Features
- Configurable retry mechanisms
- Rate limiting
- Custom backoff strategies
- Network availability detection
- Comprehensive error handling
- Support for various content types and encodings

### Security Features
- File size limits
- Directory entry limits
- Path length restrictions
- Configurable permissions
- Rate limiting for API requests

## Development

### Prerequisites
- Go 1.21 or higher
- SQLite/Turso database
- TOML configuration file

### Setup
1. Clone the repository
2. Create a `vmuser.toml` configuration file
3. Run `go mod tidy` to install dependencies

### Configuration Example
```toml
[Server]
Port = "10101"

[Turso]
DBName = "turso"
URL = "http://localhost:8080"

[Elastic]
Addresses = "https://localhost:9200"
Username = "elastic"
# Add other configuration sections as needed
```

### Running Tests
```bash
go test ./...
```

## Error Handling
The application implements comprehensive error handling with:
- Custom error types for specific scenarios
- Detailed error logging
- HTTP error responses
- Network failure recovery
- Rate limit handling

## Contributing
- Fork the repository
- Create a feature branch
- Submit a pull request

## License
[Add License Information]