package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not load .env file: %v\n", err)
	}

	// Get configuration from environment variables
	sourceDir := os.Getenv("BACKUP_SOURCE")
	destDir := os.Getenv("BACKUP_DEST")
	transferRateStr := os.Getenv("TRANSFER_RATE_MB")

	// Validate required fields
	if sourceDir == "" {
		fmt.Fprintf(os.Stderr, "Error: BACKUP_SOURCE not set in .env\n")
		os.Exit(1)
	}
	if destDir == "" {
		fmt.Fprintf(os.Stderr, "Error: BACKUP_DEST not set in .env\n")
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

	// Verify source directory exists
	srcStat, err := os.Stat(sourceDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Source directory not found: %s\n", sourceDir)
		os.Exit(1)
	}
	if !srcStat.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: Source is not a directory: %s\n", sourceDir)
		os.Exit(1)
	}

	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(destDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Could not create destination directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Backup started:\n")
	fmt.Printf("  Source: %s\n", sourceDir)
	fmt.Printf("  Destination: %s\n", destDir)
	if transferRateMB > 0 {
		fmt.Printf("  Transfer Rate Limit: %.2f MB/s\n", transferRateMB)
	} else {
		fmt.Printf("  Transfer Rate Limit: Unlimited\n")
	}
	fmt.Println()

	// Perform recursive backup
	bytesTransferred := int64(0)
	startTime := time.Now()

	if err := backupRecursive(sourceDir, destDir, sourceDir, transferRateMB, &bytesTransferred); err != nil {
		fmt.Fprintf(os.Stderr, "Error during backup: %v\n", err)
		os.Exit(1)
	}

	elapsed := time.Since(startTime)
	fmt.Printf("\nBackup completed!\n")
	fmt.Printf("  Total files transferred: %d bytes\n", bytesTransferred)
	fmt.Printf("  Time elapsed: %v\n", elapsed)
}

func backupRecursive(src, dest, srcRoot string, transferRateMB float64, bytesTransferred *int64) error {
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
			if err := backupRecursive(srcPath, dest, srcRoot, transferRateMB, bytesTransferred); err != nil {
				return err
			}
		} else {
			// Copy file
			if err := copyFile(srcPath, destPath, transferRateMB, bytesTransferred); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Could not copy file %s: %v\n", srcPath, err)
				continue
			}
			// fmt.Printf("Copied: %s\n", relPath)
		}
	}

	return nil
}

func copyFile(src, dest string, transferRateMB float64, bytesTransferred *int64) error {
	// Get source file info
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
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

	return nil
}
