package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// --- Configuration and Data ---

var monthMap = map[string]string{
	"Jan": "01", "Feb": "02", "Mar": "03", "Apr": "04",
	"May": "05", "Jun": "06", "Jul": "07", "Aug": "08",
	"Sep": "09", "Oct": "10", "Nov": "11", "Dec": "12",
}

type RenamePattern struct {
	Description string
	Regex       *regexp.Regexp
	MonthType   string
}

var patterns = []RenamePattern{
	{
		Description: "statement-Mon-YYYY.pdf",
		Regex:       regexp.MustCompile(`(?i)^statement-(?P<month>[A-Za-z]{3})-(?P<year>\d{4})\.pdf$`),
		MonthType:   "alpha",
	},
	{
		Description: "Statement_YYYYMM.pdf",
		Regex:       regexp.MustCompile(`(?i)^statement_(?P<year>\d{4})(?P<month>\d{2})\.pdf$`),
		MonthType:   "numeric",
	},
}

// --- Logic Functions ---

func resolveMonth(monthStr, monthType string) *string {
	if monthType == "alpha" {
		if val, ok := monthMap[strings.Title(strings.ToLower(monthStr))]; ok {
			return &val
		}
		return nil
	}
	if monthType == "numeric" {
		var monthNum int
		_, err := fmt.Sscanf(monthStr, "%d", &monthNum)
		if err == nil && monthNum >= 1 && monthNum <= 12 {
			result := fmt.Sprintf("%02d", monthNum)
			return &result
		}
	}
	return nil
}

func matchFile(filename string) (string, string, bool) {
	for _, pattern := range patterns {
		matches := pattern.Regex.FindStringSubmatchIndex(filename)
		if matches == nil {
			continue
		}
		groupNames := pattern.Regex.SubexpNames()
		var year, month string
		for i, name := range groupNames {
			if name == "year" {
				year = filename[matches[2*i]:matches[2*i+1]]
			} else if name == "month" {
				month = filename[matches[2*i]:matches[2*i+1]]
			}
		}
		monthNum := resolveMonth(month, pattern.MonthType)
		if monthNum == nil {
			continue
		}
		return year, *monthNum, true
	}
	return "", "", false
}

func buildNewName(year, monthNum string) string {
	return fmt.Sprintf("PaypalStatement-%s-%s.pdf", year, monthNum)
}

// --- Execution Handlers ---

func renameFiles(directory string, dryRun bool) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		fmt.Printf("Error reading directory: %v\n", err)
		return
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".pdf") {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)

	if len(files) == 0 {
		fmt.Printf("No PDF files found in '%s'.\n", directory)
		return
	}

	for _, filename := range files {
		processFile(directory, filename, dryRun)
	}
}

func renameSingleFile(path string, dryRun bool) {
	dir := filepath.Dir(path)
	filename := filepath.Base(path)
	processFile(dir, filename, dryRun)
}

func processFile(directory, filename string, dryRun bool) {
	year, monthNum, ok := matchFile(filename)
	if !ok {
		fmt.Printf("  [skip]    %q  — no match\n", filename)
		return
	}

	newName := buildNewName(year, monthNum)
	oldPath := filepath.Join(directory, filename)
	newPath := filepath.Join(directory, newName)

	if _, err := os.Stat(newPath); err == nil {
		fmt.Printf("  [skip]    %q  — target exists\n", filename)
		return
	}

	if dryRun {
		fmt.Printf("  [dry-run] %q  →  %q\n", filename, newName)
	} else {
		if err := os.Rename(oldPath, newPath); err != nil {
			fmt.Printf("  [error]   %q: %v\n", filename, err)
		} else {
			fmt.Printf("  [renamed] %q  →  %q\n", filename, newName)
		}
	}
}

// --- Main Entry Point ---

func main() {
	dryRun := flag.Bool("dry-run", false, "Preview renames.")
	flag.Parse()

	target := "."
	if flag.NArg() > 0 {
		target = flag.Arg(0)
	}

	absPath, err := filepath.Abs(target)
	if err != nil {
		fmt.Printf("Path error: %v\n", err)
		waitForExit()
		return
	}

	fi, err := os.Stat(absPath)
	if err != nil {
		fmt.Printf("Access error: %v\n", err)
		waitForExit()
		return
	}

	if fi.IsDir() {
		fmt.Printf("Scanning Directory: %s\n\n", absPath)
		renameFiles(absPath, *dryRun)
	} else {
		fmt.Printf("Processing File: %s\n\n", absPath)
		renameSingleFile(absPath, *dryRun)
	}

	waitForExit()
}

func waitForExit() {
	fmt.Println("\nExecution finished. Press 'Enter' to close...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}
