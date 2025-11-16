package ociinstaller

import (
	"fmt"

	"github.com/dustin/go-humanize"
	"golang.org/x/sys/unix"
)

// getAvailableDiskSpace returns the available disk space in bytes for the given path.
// It uses the unix.Statfs system call to get filesystem statistics.
func getAvailableDiskSpace(path string) (uint64, error) {
	var stat unix.Statfs_t
	err := unix.Statfs(path, &stat)
	if err != nil {
		return 0, fmt.Errorf("failed to get disk space for %s: %w", path, err)
	}

	// Available blocks * block size = available bytes
	// Use Bavail (available to unprivileged user) rather than Bfree (total free)
	availableBytes := stat.Bavail * uint64(stat.Bsize)
	return availableBytes, nil
}

// estimateRequiredSpace estimates the disk space required for installing an OCI image.
// This is a conservative estimate that accounts for:
// - Downloading compressed image layers
// - Extracting/unzipping archives (typically 2-3x compressed size)
// - Temporary files during installation
// - A safety buffer
//
// Actual measured OCI image sizes (as of DB 14.19.0 / FDW 2.1.3):
// - DB image compressed: 37 MB (ghcr.io/turbot/steampipe/db:14.19.0)
// - FDW image compressed: 91 MB (ghcr.io/turbot/steampipe/fdw:2.1.3)
// - Total compressed: ~128 MB
// - Typical uncompressed size: 2-3x compressed = ~350-450 MB
// - Peak disk usage (compressed + uncompressed during extraction): ~530 MB
//
// This function returns a conservative estimate of 1GB which provides:
// - ~50% safety buffer above peak usage
// - Protection against future image size increases
// - Sufficient margin for filesystem overhead and temp files
func estimateRequiredSpace(imageRef string) uint64 {
	// Conservative estimate: 1GB for Postgres/FDW installations
	// This accounts for:
	// - Download: ~130MB compressed
	// - Extraction: ~400MB uncompressed
	// - Safety buffer: ~470MB (to protect against future image growth and temp files)
	return 1 * 1024 * 1024 * 1024 // 1GB
}

// validateDiskSpace checks if sufficient disk space is available before installation.
// Returns an error if insufficient space is available, with a clear message indicating
// how much space is needed and how much is available.
func validateDiskSpace(path string, imageRef string) error {
	required := estimateRequiredSpace(imageRef)
	available, err := getAvailableDiskSpace(path)
	if err != nil {
		return fmt.Errorf("could not check disk space: %w", err)
	}

	if available < required {
		return fmt.Errorf(
			"insufficient disk space: need ~%s, have %s available at %s",
			humanize.Bytes(required),
			humanize.Bytes(available),
			path,
		)
	}

	return nil
}
