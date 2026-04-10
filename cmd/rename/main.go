package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Rename PDF statement files to a consistent 'PaypalStatement-YYYY-MM.pdf' format.
//
// Supported input patterns:
//   1. statement-Apr-2024.pdf   →  PaypalStatement-2024-04.pdf
//   2. Statement_201406.pdf     →  PaypalStatement-2014-06.pdf
//
// # Build the rename tool
// go build -o rename.exe ./cmd/rename

var monthMap = map[string]string{
	"Jan": "01", "Feb": "02", "Mar": "03", "Apr": "04",
	"May": "05", "Jun": "06", "Jul": "07", "Aug": "08",
	"Sep": "09", "Oct": "10", "Nov": "11", "Dec": "12",
}

// RenamePattern matches filenames and extracts year/month for renaming.
type RenamePattern struct {
	Description string
	Regex       *regexp.Regexp
	// 'alpha'   → month is a 3-letter abbreviation (Jan, Feb, …)
	// 'numeric' → month is already a zero-padded number (01–12)
	MonthType string
}

var patterns = []RenamePattern{
	{
		Description: "statement-Mon-YYYY.pdf",
		Regex: regexp.MustCompile(
			`(?i)^statement-(?P<month>[A-Za-z]{3})-(?P<year>\d{4})\.pdf$`,
		),
		MonthType: "alpha",
	},
	{
		Description: "Statement_YYYYMM.pdf",
		Regex: regexp.MustCompile(
			`(?i)^statement_(?P<year>\d{4})(?P<month>\d{2})\.pdf$`,
		),
		MonthType: "numeric",
	},
}

func resolveMonth(monthStr, monthType string) *string {
	// Return a zero-padded month number string, or nil on failure.
	//   monthType='alpha'   : 'Apr' → '04'
	//   monthType='numeric' : '06'  → '06' (validated to be 01–12)
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
		return nil
	}
	return nil
}

func matchFile(filename string) (string, string, bool) {
	// Try every pattern against filename.
	// Returns (year, month_num, ok) on the first match, or (_, _, false) if nothing matches.
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

func renameFiles(directory string, dryRun bool) {
	fi, err := os.Stat(directory)
	if err != nil || !fi.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: '%s' is not a valid directory.\n", directory)
		os.Exit(1)
	}

	entries, err := os.ReadDir(directory)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading directory: %v\n", err)
		os.Exit(1)
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

	renamed := 0
	skipped := 0

	for _, filename := range files {
		year, monthNum, ok := matchFile(filename)
		if !ok {
			fmt.Printf("  [skip]    %q  — doesn't match any pattern\n", filename)
			skipped++
			continue
		}

		newName := buildNewName(year, monthNum)
		oldPath := filepath.Join(directory, filename)
		newPath := filepath.Join(directory, newName)

		if _, err := os.Stat(newPath); err == nil {
			fmt.Printf("  [skip]    %q  — target '%s' already exists\n", filename, newName)
			skipped++
			continue
		}

		if dryRun {
			fmt.Printf("  [dry-run] %q  →  %q\n", filename, newName)
		} else {
			if err := os.Rename(oldPath, newPath); err != nil {
				fmt.Fprintf(os.Stderr, "Error renaming %s: %v\n", filename, err)
				skipped++
				continue
			}
			fmt.Printf("  [renamed] %q  →  %q\n", filename, newName)
		}

		renamed++
	}

	suffix := ""
	if dryRun {
		suffix = "would be "
	}
	fmt.Printf("\nDone. %d file(s) %srenamed, %d skipped.\n", renamed, suffix, skipped)
}

func main() {
	dryRun := flag.Bool("dry-run", false, "Preview renames without making any changes.")
	flag.Parse()

	directory := "."
	if flag.NArg() > 0 {
		directory = flag.Arg(0)
	}

	absDir, err := filepath.Abs(directory)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving directory: %v\n", err)
		os.Exit(1)
	}

	prefix := ""
	if *dryRun {
		prefix = "[DRY RUN] "
	}
	fmt.Printf("%sScanning: %s\n\n", prefix, absDir)

	renameFiles(absDir, *dryRun)
}
