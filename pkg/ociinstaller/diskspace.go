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
// This is a practical estimate that accounts for:
// - Downloading compressed image layers
// - Extracting/unzipping archives (typically 2-3x compressed size)
// - Temporary files during installation
//
// Actual measured OCI image sizes (as of DB 18.4.0-rc.1 / FDW 2.3.0-rc.0; one platform layer is downloaded):
// - DB platform layer compressed: 26-27 MB (ghcr.io/turbot/steampipe/db:18.4.0-rc.1)
// - DB installed (uncompressed, incl. bundled ICU + OpenSSL): ~136 MB
// - FDW platform layer compressed: 19-34 MB (ghcr.io/turbot/steampipe/fdw:2.3.0-rc.0)
// - FDW uncompressed: ~62 MB darwin-arm64, larger on linux (~100-130 MB)
// - Peak disk usage (compressed + uncompressed during extraction): ~330 MB worst-case platform
//
// This function returns 500MB which:
// - Covers the measured ~330 MB peak with ~170 MB headroom
// - Avoids blocking installations that have adequate space (600-700 MB available)
// - Balances safety against false rejections in constrained environments
// - May fail if filesystem overhead or temp files exceed expectations, but will catch
//   the primary failure case (truly insufficient disk space)
func estimateRequiredSpace(imageRef string) uint64 {
	// Practical estimate: 500MB for Postgres/FDW installations - measured peak ~330 MB plus buffer.
	return 500 * 1024 * 1024 // 500MB
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
