package database

const (
	MaxFileSize          = 10 * 1024 * 1024 // 10MB max file size
	MaxFilesPerDirectory = 1000             // Prevent directory bombs
	MaxPathLength        = 256              // Reasonable path length limit
)
