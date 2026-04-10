#!/usr/bin/env python3
"""
Rename PDF statement files to a consistent 'PaypalStatement-YYYY-MM.pdf' format.

Supported input patterns:
  1. statement-Apr-2024.pdf   →  PaypalStatement-2024-04.pdf
  2. Statement_201406.pdf     →  PaypalStatement-2014-06.pdf

Usage:
    python rename_files.py                        # renames in current directory
    python rename_files.py /path/to/dir           # renames in specified directory
    python rename_files.py /path/to/dir --dry-run # preview changes without renaming
"""

import argparse
import re
import sys
from pathlib import Path
from dataclasses import dataclass

MONTH_MAP = {
    "Jan": "01", "Feb": "02", "Mar": "03", "Apr": "04",
    "May": "05", "Jun": "06", "Jul": "07", "Aug": "08",
    "Sep": "09", "Oct": "10", "Nov": "11", "Dec": "12",
}


@dataclass
class RenamePattern:
    """A pattern that matches filenames and extracts year/month for renaming."""
    description: str
    regex: re.Pattern
    # 'alpha'   → month is a 3-letter abbreviation (Jan, Feb, …)
    # 'numeric' → month is already a zero-padded number (01–12)
    month_type: str


PATTERNS: list[RenamePattern] = [
    RenamePattern(
        description="statement-Mon-YYYY.pdf",
        regex=re.compile(
            r"^statement-(?P<month>[A-Za-z]{3})-(?P<year>\d{4})\.pdf$",
            re.IGNORECASE,
        ),
        month_type="alpha",
    ),
    RenamePattern(
        description="Statement_YYYYMM.pdf",
        regex=re.compile(
            r"^statement_(?P<year>\d{4})(?P<month>\d{2})\.pdf$",
            re.IGNORECASE,
        ),
        month_type="numeric",
    ),
]


def resolve_month(month_str: str, month_type: str) -> str | None:
    """
    Return a zero-padded month number string, or None on failure.
      month_type='alpha'   : 'Apr' → '04'
      month_type='numeric' : '06'  → '06' (validated to be 01–12)
    """
    if month_type == "alpha":
        return MONTH_MAP.get(month_str.capitalize())
    if month_type == "numeric":
        if 1 <= int(month_str) <= 12:
            return month_str.zfill(2)
        return None
    return None


def match_file(filename: str) -> tuple[str, str] | None:
    """
    Try every pattern against filename.
    Returns (year, month_num) on the first match, or None if nothing matches.
    """
    for pattern in PATTERNS:
        m = pattern.regex.match(filename)
        if not m:
            continue
        month_num = resolve_month(m.group("month"), pattern.month_type)
        if month_num is None:
            continue
        return m.group("year"), month_num
    return None


def build_new_name(year: str, month_num: str) -> str:
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
        result = match_file(filepath.name)
        if result is None:
            print(f"  [skip]    {filepath.name!r}  — doesn't match any pattern")
            skipped += 1
            continue

        year, month_num = result
        new_name = build_new_name(year, month_num)
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
        description="Rename PDF statement files to 'PaypalStatement-YYYY-MM.pdf'.",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="Supported patterns:\n" + "\n".join(
            f"  {p.description}" for p in PATTERNS
        ),
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
    