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
// For Postgres images, typical sizes are:
// - Compressed: 300-400 MB
// - Uncompressed: 1-1.2 GB
// - With extraction overhead and temp files: ~2x uncompressed size
//
// This function returns a conservative estimate of 2GB for database installations.
func estimateRequiredSpace(imageRef string) uint64 {
	// Conservative estimate: 2GB for Postgres/FDW installations
	// This accounts for:
	// - Download: ~400MB compressed
	// - Extraction: ~1.2GB uncompressed
	// - Temp files and safety buffer: additional ~400MB
	return 2 * 1024 * 1024 * 1024 // 2GB
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
