package database

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

type VirtualFile struct {
	ID        string    `json:"id"`
	Path      string    `json:"path"`
	Content   []byte    `json:"content"`
	Metadata  Metadata  `json:"metadata"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Metadata struct {
	MimeType    string            `json:"mime_type"`
	Tags        []string          `json:"tags"`
	Permissions map[string]string `json:"permissions"`
}

// Schema definitions
var schemas = []string{
	`CREATE TABLE IF NOT EXISTS system_config (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`,

	`CREATE TABLE IF NOT EXISTS virtual_filesystem (
		id TEXT PRIMARY KEY,
		path TEXT NOT NULL UNIQUE,
		content BLOB,
		metadata JSON,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(path)
	)`,

	`CREATE TABLE IF NOT EXISTS operation_log (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		operation TEXT NOT NULL,
		details JSON,
		timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`,

	`CREATE INDEX IF NOT EXISTS idx_vfs_path ON virtual_filesystem(path)`,
}

// FileSystem interface that the LLM will interact with
type VirtualFileSystem interface {
	// Basic file operations
	CreateFile(path string, content []byte, metadata Metadata) error
	ReadFile(path string) (*VirtualFile, error)
	UpdateFile(path string, content []byte) error
	DeleteFile(path string) error

	// Directory operations
	ListFiles(path string) ([]VirtualFile, error)
	CreateDirectory(path string) error

	// Search and query
	SearchFiles(query string) ([]VirtualFile, error)

	// Metadata operations
	UpdateMetadata(path string, metadata Metadata) error
	GetMetadata(path string) (Metadata, error)
}

// Implementation for Turso
type TursoFileSystem struct {
	db *sql.DB
}

func NewTursoFileSystem(dsn string) (*TursoFileSystem, error) {
	db, err := sql.Open("libsql", dsn)
	if err != nil {
		return nil, err
	}

	fs := &TursoFileSystem{db: db}
	if err := fs.initialize(); err != nil {
		db.Close()
		return nil, err
	}

	return fs, nil
}

func (fs *TursoFileSystem) initialize() error {
	// Initialize schemas
	for _, schema := range schemas {
		if _, err := fs.db.Exec(schema); err != nil {
			return err
		}
	}
	return nil
}

func (fs *TursoFileSystem) CreateFile(path string, content []byte, metadata Metadata) error {
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	if len(content) > MaxFileSize {
		return fmt.Errorf("file exceeds maximum size of %d bytes", MaxFileSize)
	}
	if len(path) > MaxPathLength {
		return fmt.Errorf("path exceeds maximum length of %d characters", MaxPathLength)
	}

	_, err = fs.db.Exec(`
		INSERT INTO virtual_filesystem (id, path, content, metadata)
		VALUES (?, ?, ?, ?)
	`, generateUUID(), path, content, metadataJSON)

	return err
}

type ComputerUseContext struct {
	fs VirtualFileSystem
	db *sql.DB
}

func (ctx *ComputerUseContext) HandleOperation(op string, args map[string]interface{}) (interface{}, error) {
	// Log operation
	details, _ := json.Marshal(args)
	_, err := ctx.db.Exec(`
		INSERT INTO operation_log (operation, details)
		VALUES (?, ?)
	`, op, string(details))

	if err != nil {
		return nil, err
	}

	// Handle operation based on type
	switch op {
	case "write_file":
		return ctx.handleWriteFile(args)
	case "read_file":
		return ctx.handleReadFile(args)
		// ... other operations
	}

	return nil, nil
}

// ReadFile retrieves a file from the virtual filesystem
func (fs *TursoFileSystem) ReadFile(path string) (*VirtualFile, error) {
	var file VirtualFile
	var metadataStr string

	err := fs.db.QueryRow(`
		SELECT id, path, content, metadata, created_at, updated_at 
		FROM virtual_filesystem 
		WHERE path = ?
	`, path).Scan(
		&file.ID,
		&file.Path,
		&file.Content,
		&metadataStr,
		&file.CreatedAt,
		&file.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("file not found: %s", path)
	}
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Parse metadata JSON
	if err := json.Unmarshal([]byte(metadataStr), &file.Metadata); err != nil {
		return nil, fmt.Errorf("metadata parse error: %w", err)
	}

	return &file, nil
}

// UpdateFile modifies an existing file's content
func (fs *TursoFileSystem) UpdateFile(path string, content []byte) error {
	result, err := fs.db.Exec(`
		UPDATE virtual_filesystem 
		SET content = ?, updated_at = CURRENT_TIMESTAMP 
		WHERE path = ?
	`, content, path)

	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking update result: %w", err)
	}
	if rows == 0 {
		return errors.New("file not found")
	}

	return nil
}

// DeleteFile removes a file from the virtual filesystem
func (fs *TursoFileSystem) DeleteFile(path string) error {
	result, err := fs.db.Exec(`
		DELETE FROM virtual_filesystem 
		WHERE path = ?
	`, path)

	if err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking delete result: %w", err)
	}
	if rows == 0 {
		return errors.New("file not found")
	}

	return nil
}

// ListFiles retrieves all files in a directory
func (fs *TursoFileSystem) ListFiles(path string) ([]VirtualFile, error) {
	// Ensure path ends with / for directory matching
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	rows, err := fs.db.Query(`
		SELECT id, path, content, metadata, created_at, updated_at 
		FROM virtual_filesystem 
		WHERE path LIKE ? || '%'
	`, path)

	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var files []VirtualFile
	for rows.Next() {
		var file VirtualFile
		var metadataStr string

		err := rows.Scan(
			&file.ID,
			&file.Path,
			&file.Content,
			&metadataStr,
			&file.CreatedAt,
			&file.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("row scan failed: %w", err)
		}

		if err := json.Unmarshal([]byte(metadataStr), &file.Metadata); err != nil {
			return nil, fmt.Errorf("metadata parse error: %w", err)
		}

		files = append(files, file)
	}

	return files, nil
}

// CreateDirectory creates a new directory entry
func (fs *TursoFileSystem) CreateDirectory(path string) error {
	// Ensure path ends with /
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	metadata := Metadata{
		MimeType:    "directory",
		Tags:        []string{"directory"},
		Permissions: map[string]string{"type": "directory"},
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("metadata marshaling failed: %w", err)
	}

	_, err = fs.db.Exec(`
		INSERT INTO virtual_filesystem (id, path, metadata)
		VALUES (?, ?, ?)
	`, generateUUID(), path, metadataJSON)

	if err != nil {
		return fmt.Errorf("directory creation failed: %w", err)
	}

	return nil
}

// SearchFiles searches for files matching the query
func (fs *TursoFileSystem) SearchFiles(query string) ([]VirtualFile, error) {
	rows, err := fs.db.Query(`
		SELECT id, path, content, metadata, created_at, updated_at 
		FROM virtual_filesystem 
		WHERE path LIKE ? OR metadata LIKE ?
	`, "%"+query+"%", "%"+query+"%")

	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}
	defer rows.Close()

	var files []VirtualFile
	for rows.Next() {
		var file VirtualFile
		var metadataStr string

		err := rows.Scan(
			&file.ID,
			&file.Path,
			&file.Content,
			&metadataStr,
			&file.CreatedAt,
			&file.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("row scan failed: %w", err)
		}

		if err := json.Unmarshal([]byte(metadataStr), &file.Metadata); err != nil {
			return nil, fmt.Errorf("metadata parse error: %w", err)
		}

		files = append(files, file)
	}

	return files, nil
}

// UpdateMetadata updates a file's metadata
func (fs *TursoFileSystem) UpdateMetadata(path string, metadata Metadata) error {
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("metadata marshaling failed: %w", err)
	}

	result, err := fs.db.Exec(`
		UPDATE virtual_filesystem 
		SET metadata = ?, updated_at = CURRENT_TIMESTAMP 
		WHERE path = ?
	`, metadataJSON, path)

	if err != nil {
		return fmt.Errorf("metadata update failed: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking update result: %w", err)
	}
	if rows == 0 {
		return errors.New("file not found")
	}

	return nil
}

// GetMetadata retrieves a file's metadata
func (fs *TursoFileSystem) GetMetadata(path string) (Metadata, error) {
	var metadataStr string
	err := fs.db.QueryRow(`
		SELECT metadata 
		FROM virtual_filesystem 
		WHERE path = ?
	`, path).Scan(&metadataStr)

	if err == sql.ErrNoRows {
		return Metadata{}, fmt.Errorf("file not found: %s", path)
	}
	if err != nil {
		return Metadata{}, fmt.Errorf("database error: %w", err)
	}

	var metadata Metadata
	if err := json.Unmarshal([]byte(metadataStr), &metadata); err != nil {
		return Metadata{}, fmt.Errorf("metadata parse error: %w", err)
	}

	return metadata, nil
}

// ComputerUseContext handler implementations
func (ctx *ComputerUseContext) handleWriteFile(args map[string]interface{}) (interface{}, error) {
	path, ok := args["path"].(string)
	if !ok {
		return nil, errors.New("path must be a string")
	}

	content, ok := args["content"].([]byte)
	if !ok {
		if strContent, ok := args["content"].(string); ok {
			content = []byte(strContent)
		} else {
			return nil, errors.New("content must be bytes or string")
		}
	}

	_, err := ctx.fs.ReadFile(path)
	if err == nil {
		return nil, ctx.fs.UpdateFile(path, content)
	}

	// File doesn't exist, create it
	metadata := Metadata{
		MimeType:    detectMimeType(path, content),
		Tags:        []string{},
		Permissions: map[string]string{"access": "rw"},
	}

	return nil, ctx.fs.CreateFile(path, content, metadata)
}

func (ctx *ComputerUseContext) handleReadFile(args map[string]interface{}) (interface{}, error) {
	path, ok := args["path"].(string)
	if !ok {
		return nil, errors.New("path must be a string")
	}

	return ctx.fs.ReadFile(path)
}

// Helper function to detect MIME type based on file extension and content
func detectMimeType(path string, content []byte) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".txt":
		return "text/plain"
	case ".json":
		return "application/json"
	case ".md":
		return "text/markdown"
	case ".html":
		return "text/html"
	default:
		return "application/octet-stream"
	}
}

func generateUUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		// In case of error, create a timestamp-based fallback
		return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
