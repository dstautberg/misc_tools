#!/usr/bin/env python3
"""
Rename files from 'statement-Mon-YYYY.pdf' to 'PaypalStatement-YYYY-MM.pdf'

Usage:
    python rename_files.py                        # renames in current directory
    python rename_files.py /path/to/dir           # renames in specified directory
    python rename_files.py /path/to/dir --dry-run # preview changes without renaming
"""

import argparse
import re
import sys
from pathlib import Path

MONTH_MAP = {
    "Jan": "01", "Feb": "02", "Mar": "03", "Apr": "04",
    "May": "05", "Jun": "06", "Jul": "07", "Aug": "08",
    "Sep": "09", "Oct": "10", "Nov": "11", "Dec": "12",
}

# Matches: statement-Apr-2024.pdf  (case-insensitive prefix)
PATTERN = re.compile(
    r"^statement-(?P<month>[A-Za-z]{3})-(?P<year>\d{4})\.pdf$",
    re.IGNORECASE,
)


def build_new_name(month_str: str, year: str) -> str | None:
    """Return the new filename, or None if the month abbreviation is unrecognised."""
    # Normalise to title-case so the dict lookup works regardless of input case
    month_title = month_str.capitalize()
    month_num = MONTH_MAP.get(month_title)
    if month_num is None:
        return None
    return f"PaypalStatement-{year}-{month_num}.pdf"


def rename_files(directory: Path, dry_run: bool) -> None:
    if not directory.is_dir():
        print(f"Error: '{directory}' is not a valid directory.")
        sys.exit(1)

    files = sorted(directory.glob("*.pdf"))
    if not files:
        print(f"No PDF files found in '{directory}'.")
        return

    renamed = 0
    skipped = 0

    for filepath in files:
        match = PATTERN.match(filepath.name)
        if not match:
            print(f"  [skip]    {filepath.name!r}  — doesn't match pattern")
            skipped += 1
            continue

        new_name = build_new_name(match.group("month"), match.group("year"))
        if new_name is None:
            print(f"  [skip]    {filepath.name!r}  — unknown month abbreviation")
            skipped += 1
            continue

        new_path = filepath.with_name(new_name)

        if new_path.exists():
            print(f"  [skip]    {filepath.name!r}  — target '{new_name}' already exists")
            skipped += 1
            continue

        if dry_run:
            print(f"  [dry-run] {filepath.name!r}  →  {new_name!r}")
        else:
            filepath.rename(new_path)
            print(f"  [renamed] {filepath.name!r}  →  {new_name!r}")

        renamed += 1

    print(f"\nDone. {renamed} file(s) {'would be ' if dry_run else ''}renamed, {skipped} skipped.")


def main() -> None:
    parser = argparse.ArgumentParser(
        description="Rename 'statement-Mon-YYYY.pdf' files to 'PaypalStatement-YYYY-MM.pdf'."
    )
    parser.add_argument(
        "directory",
        nargs="?",
        default=".",
        help="Directory containing the files (default: current directory).",
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Preview renames without making any changes.",
    )
    args = parser.parse_args()

    directory = Path(args.directory).resolve()
    print(f"{'[DRY RUN] ' if args.dry_run else ''}Scanning: {directory}\n")
    rename_files(directory, dry_run=args.dry_run)


if __name__ == "__main__":
    main()