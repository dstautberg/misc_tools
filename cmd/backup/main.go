package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not load .env file: %v\n", err)
	}

	// Get configuration from environment variables
	// Support multiple sources/dests via BACKUP_SOURCES / BACKUP_DESTS (comma or semicolon separated)
	// Fall back to singular BACKUP_SOURCE / BACKUP_DEST for compatibility
	// Read new GO_ names; fall back to legacy names if present
	sourcesEnv := os.Getenv("GO_BACKUP_SOURCES")
	if sourcesEnv == "" {
		sourcesEnv = os.Getenv("GO_BACKUP_SOURCE")
		if sourcesEnv == "" {
			// legacy fallback
			sourcesEnv = os.Getenv("BACKUP_SOURCES")
			if sourcesEnv == "" {
				sourcesEnv = os.Getenv("BACKUP_SOURCE")
			}
		}
	}

	destsEnv := os.Getenv("GO_BACKUP_DESTINATIONS")
	if destsEnv == "" {
		destsEnv = os.Getenv("GO_BACKUP_DESTINATION")
		if destsEnv == "" {
			// legacy fallback
			destsEnv = os.Getenv("BACKUP_DESTS")
			if destsEnv == "" {
				destsEnv = os.Getenv("BACKUP_DEST")
			}
		}
	}
	transferRateStr := os.Getenv("TRANSFER_RATE_MB")

	// Parse sources and destinations
	splitPaths := func(s string) []string {
		if s == "" {
			return nil
		}
		// split on comma or semicolon
		parts := strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == ';' })
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			t := strings.TrimSpace(p)
			if t != "" {
				out = append(out, t)
			}
		}
		return out
	}

	sources := splitPaths(sourcesEnv)
	dests := splitPaths(destsEnv)

	if len(sources) == 0 {
		fmt.Fprintf(os.Stderr, "Error: GO_BACKUP_SOURCES or GO_BACKUP_SOURCE not set in .env\n")
		os.Exit(1)
	}
	if len(dests) == 0 {
		fmt.Fprintf(os.Stderr, "Error: GO_BACKUP_DESTINATIONS or GO_BACKUP_DESTINATION not set in .env\n")
		os.Exit(1)
	}

	// Parse transfer rate (MB/s, 0 = unlimited)
	var transferRateMB float64
	if transferRateStr == "" {
		transferRateMB = 0 // unlimited
	} else {
		rate, err := strconv.ParseFloat(transferRateStr, 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Invalid TRANSFER_RATE_MB value: %s\n", transferRateStr)
			os.Exit(1)
		}
		transferRateMB = rate
	}

	// We'll validate sources/destinations and run backups per pair below

	fmt.Printf("Backup started:\n")
	fmt.Printf("  Sources: %s\n", strings.Join(sources, ", "))
	fmt.Printf("  Destinations: %s\n", strings.Join(dests, ", "))
	if transferRateMB > 0 {
		fmt.Printf("  Transfer Rate Limit: %.2f MB/s\n", transferRateMB)
	} else {
		fmt.Printf("  Transfer Rate Limit: Unlimited\n")
	}
	fmt.Println()

	// Perform recursive backup for each source/destination pair
	bytesTransferred := int64(0)
	filesCopied := int64(0)
	filesSkipped := int64(0)
	startTime := time.Now()

	for i, src := range sources {
		// Determine destination for this source
		var destRoot string
		if len(dests) == 1 {
			// Single destination: create subdir using basename of source
			destRoot = filepath.Join(dests[0], filepath.Base(src))
		} else {
			// Multiple destinations: must match by index
			if i >= len(dests) {
				fmt.Fprintf(os.Stderr, "Error: not enough destinations provided for source %s\n", src)
				os.Exit(1)
			}
			destRoot = dests[i]
		}

		fmt.Printf("\nBacking up source: %s -> %s\n", src, destRoot)

		// Verify source exists and is a directory
		srcStat, err := os.Stat(src)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Source not found, skipping: %s (%v)\n", src, err)
			continue
		}
		if !srcStat.IsDir() {
			fmt.Fprintf(os.Stderr, "Warning: Source is not a directory, skipping: %s\n", src)
			continue
		}

		// Create destination directory if it doesn't exist
		if err := os.MkdirAll(destRoot, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not create destination directory %s: %v\n", destRoot, err)
			continue
		}

		if err := backupRecursive(src, destRoot, src, transferRateMB, &bytesTransferred, &filesCopied, &filesSkipped); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Error during backup of %s: %v\n", src, err)
			continue
		}
	}

	elapsed := time.Since(startTime)
	fmt.Printf("\nBackup completed!\n")
	fmt.Printf("  Files copied: %d\n", filesCopied)
	fmt.Printf("  Files skipped: %d\n", filesSkipped)
	dataMB := float64(bytesTransferred) / (1024 * 1024)
	if dataMB >= 1024 {
		fmt.Printf("  Total data transferred: %.2f GB\n", dataMB/1024)
	} else {
		fmt.Printf("  Total data transferred: %.2f MB\n", dataMB)
	}
	fmt.Printf("  Time elapsed: %v\n", elapsed)
}

func backupRecursive(src, dest, srcRoot string, transferRateMB float64, bytesTransferred *int64, filesCopied *int64, filesSkipped *int64) error {
	files, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}

	for _, file := range files {
		srcPath := filepath.Join(src, file.Name())
		relPath, _ := filepath.Rel(srcRoot, srcPath)
		destPath := filepath.Join(dest, relPath)

		if file.IsDir() {
			// Create destination directory
			if err := os.MkdirAll(destPath, 0755); err != nil {
				return err
			}
			// Recursively copy contents
			if err := backupRecursive(srcPath, dest, srcRoot, transferRateMB, bytesTransferred, filesCopied, filesSkipped); err != nil {
				return err
			}
		} else {
			// Copy file
			if err := copyFile(srcPath, destPath, transferRateMB, bytesTransferred, filesCopied, filesSkipped); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Could not copy file %s: %v\n", srcPath, err)
				continue
			}
			// fmt.Printf("Copied: %s\n", relPath)
		}
	}

	return nil
}

func copyFile(src, dest string, transferRateMB float64, bytesTransferred *int64, filesCopied *int64, filesSkipped *int64) error {
	// Get source file info
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// If destination exists and is newer or same, skip copying
	if destInfo, err := os.Stat(dest); err == nil {
		if !srcInfo.ModTime().After(destInfo.ModTime()) {
			fmt.Printf("[%s] Skipping (dest up-to-date): %s\n", time.Now().Format("2006-01-02 15:04:05"), filepath.Base(src))
			(*filesSkipped)++
			return nil
		}
	}

	fileSizeMB := float64(srcInfo.Size()) / (1024 * 1024)
	fmt.Printf("[%s] Copying: %s (%.2f MB)", time.Now().Format("2006-01-02 15:04:05"), filepath.Base(src), fileSizeMB)

	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create destination file
	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	fileStartTime := time.Now()
	fileBytes := int64(0)
	lastUpdateTime := fileStartTime
	lastUpdateBytes := int64(0)

	// Copy with rate limiting
	if transferRateMB > 0 {
		// Transfer with rate limiting
		bytesPerSecond := int64(transferRateMB * 1024 * 1024)
		bufferSize := int64(32 * 1024) // 32 KB chunks
		if bufferSize > bytesPerSecond {
			bufferSize = bytesPerSecond
		}

		buffer := make([]byte, bufferSize)
		for {
			n, err := srcFile.Read(buffer)
			if err != nil && err != io.EOF {
				return err
			}
			if n == 0 {
				break
			}

			if _, err := destFile.Write(buffer[:n]); err != nil {
				return err
			}

			*bytesTransferred += int64(n)
			fileBytes += int64(n)

			// Update bandwidth display every 5 seconds
			now := time.Now()
			if now.Sub(lastUpdateTime) >= 5*time.Second {
				bytesSinceUpdate := fileBytes - lastUpdateBytes
				timeSinceUpdate := now.Sub(lastUpdateTime).Seconds()
				// Convert bytes to MB and divide by seconds: (MB) / (Seconds) = MB/s
				mbSinceUpdate := float64(bytesSinceUpdate) / (1024 * 1024)
				bandwidth := mbSinceUpdate / timeSinceUpdate
				fmt.Printf("\r[%s] Copying: %s (%.2f MB) - %.2f MB/s", now.Format("2006-01-02 15:04:05"), filepath.Base(src), fileSizeMB, bandwidth)
				lastUpdateTime = now
				lastUpdateBytes = fileBytes
			}

			// Rate limiting: sleep if necessary
			if transferRateMB > 0 {
				time.Sleep(time.Duration(int64(n)) * time.Second / time.Duration(bytesPerSecond))
			}
		}
	} else {
		// Copy without rate limiting
		buffer := make([]byte, 1024*1024) // 1 MB buffer
		for {
			n, err := srcFile.Read(buffer)
			if err != nil && err != io.EOF {
				return err
			}
			if n == 0 {
				break
			}

			if _, err := destFile.Write(buffer[:n]); err != nil {
				return err
			}

			*bytesTransferred += int64(n)
			fileBytes += int64(n)

			// Update bandwidth display every 5 seconds
			now := time.Now()
			if now.Sub(lastUpdateTime) >= 5*time.Second {
				bytesSinceUpdate := fileBytes - lastUpdateBytes
				timeSinceUpdate := now.Sub(lastUpdateTime).Seconds()
				// Convert bytes to MB and divide by seconds: (MB) / (Seconds) = MB/s
				mbSinceUpdate := float64(bytesSinceUpdate) / (1024 * 1024)
				bandwidth := mbSinceUpdate / timeSinceUpdate
				fmt.Printf("\r[%s] Copying: %s (%.2f MB) - %.2f MB/s", now.Format("2006-01-02 15:04:05"), filepath.Base(src), fileSizeMB, bandwidth)
				lastUpdateTime = now
				lastUpdateBytes = fileBytes
			}
		}
	}

	// Preserve file permissions
	if err := os.Chmod(dest, srcInfo.Mode()); err != nil {
		return err
	}

	// Calculate and display final bandwidth
	elapsed := time.Since(fileStartTime)
	// Convert bytes to MB and divide by seconds: (MB) / (Seconds) = MB/s
	mbTotal := float64(fileBytes) / (1024 * 1024)
	bandwidth := mbTotal / elapsed.Seconds()
	fmt.Printf("\r[%s] Copying: %s (%.2f MB) - %.2f MB/s\n", time.Now().Format("2006-01-02 15:04:05"), filepath.Base(src), fileSizeMB, bandwidth)

	// Count successful copy
	(*filesCopied)++
	return nil
}
