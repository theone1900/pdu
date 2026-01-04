// Package fileio provides file I/O operations for PostgreSQL data files.
package fileio

import (
	"os"
	"path/filepath"

	"github.com/wublabdubdub/pdu/pkg/pgtypes"
)

// FileReader is an interface for reading PostgreSQL data files.
type FileReader interface {
	// Open opens a file for reading
	Open(path string) error
	
	// Close closes the file
	Close() error
	
	// ReadPage reads a single page from the file
	ReadPage(pageNumber int64) ([]byte, error)
	
	// ReadBytes reads a specified number of bytes from the file at a given offset
	ReadBytes(offset int64, length int) ([]byte, error)
	
	// GetFileSize returns the size of the file
	GetFileSize() int64
	
	// GetPageCount returns the number of pages in the file
	GetPageCount() int64
}

// PgFileReader implements FileReader for PostgreSQL data files.
type PgFileReader struct {
	file     *os.File
	fileSize int64
}

// NewPgFileReader creates a new PgFileReader instance.
func NewPgFileReader() *PgFileReader {
	return &PgFileReader{
		file:     nil,
		fileSize: 0,
	}
}

// Open opens a file for reading.
func (r *PgFileReader) Open(path string) error {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}

	// Open the file
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	// Get file size
	fileInfo, err := file.Stat()
	if err != nil {
		file.Close()
		return err
	}

	// Set file and file size
	r.file = file
	r.fileSize = fileInfo.Size()

	return nil
}

// Close closes the file.
func (r *PgFileReader) Close() error {
	if r.file != nil {
		return r.file.Close()
	}
	return nil
}

// ReadPage reads a single page from the file.
func (r *PgFileReader) ReadPage(pageNumber int64) ([]byte, error) {
	// Calculate offset
	offset := pageNumber * int64(pgtypes.BLCKSZ)

	// Check if offset is within file bounds
	if offset >= r.fileSize {
		return nil, os.ErrInvalid
	}

	// Calculate bytes to read (may be less than BLCKSZ for last page)
	bytesToRead := int64(pgtypes.BLCKSZ)
	if offset+bytesToRead > r.fileSize {
		bytesToRead = r.fileSize - offset
	}

	// Allocate buffer
	buffer := make([]byte, bytesToRead)

	// Read the page
	n, err := r.file.ReadAt(buffer, offset)
	if err != nil {
		return nil, err
	}

	// Check if we read the expected number of bytes
	if int64(n) != bytesToRead {
		return nil, os.ErrInvalid
	}

	return buffer, nil
}

// ReadBytes reads a specified number of bytes from the file at a given offset.
func (r *PgFileReader) ReadBytes(offset int64, length int) ([]byte, error) {
	// Check if offset and length are within file bounds
	if offset < 0 || int64(length) <= 0 || offset+int64(length) > r.fileSize {
		return nil, os.ErrInvalid
	}

	// Allocate buffer
	buffer := make([]byte, length)

	// Read the bytes
	n, err := r.file.ReadAt(buffer, offset)
	if err != nil {
		return nil, err
	}

	// Check if we read the expected number of bytes
	if n != length {
		return nil, os.ErrInvalid
	}

	return buffer, nil
}

// GetFileSize returns the size of the file.
func (r *PgFileReader) GetFileSize() int64 {
	return r.fileSize
}

// GetPageCount returns the number of pages in the file.
func (r *PgFileReader) GetPageCount() int64 {
	if r.fileSize == 0 {
		return 0
	}

	pageCount := r.fileSize / int64(pgtypes.BLCKSZ)
	if r.fileSize%int64(pgtypes.BLCKSZ) > 0 {
		pageCount++
	}

	return pageCount
}

// FindDataFiles finds all PostgreSQL data files in a directory.
func FindDataFiles(dir string) ([]string, error) {
	var dataFiles []string

	// Walk through the directory
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if the file is a PostgreSQL data file
		// Data files have numeric names without extensions
		name := info.Name()
		if IsNumeric(name) {
			dataFiles = append(dataFiles, path)
		}

		return nil
	})

	return dataFiles, err
}

// IsNumeric checks if a string is numeric.
func IsNumeric(s string) bool {
	if s == "" {
		return false
	}

	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}

	return true
}
