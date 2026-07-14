#!/usr/bin/env python3
"""
Replace image filenames in markdown links with UUID-based URLs.

Example:
    (../assets/image_1773192791000_0.png)
        -> (https://img.assets.internal/img/019f5e57-f103-76f4-b0b9-188e4ebf1e2d)

Usage:
    python replace_image_paths.py img.csv TARGET_DIR [--dry-run]
"""

import argparse
import csv
import re
import sys
from pathlib import Path

BASE_URL = "https://img.assets.internal/img"


def load_mapping(csv_path: Path) -> dict[str, str]:
    """Load filename -> uuid mapping from the CSV file."""
    mapping = {}
    with csv_path.open(newline="", encoding="utf-8") as f:
        reader = csv.DictReader(f)
        for row in reader:
            mapping[row["filename"]] = row["uuid"]
    return mapping


def build_pattern(filenames: list[str]) -> re.Pattern:
    """Build a single regex matching any of the target filenames inside a
    markdown link path, e.g. (../assets/<filename>)."""
    # Escape filenames for safe use in regex, then join as alternatives.
    escaped = "|".join(re.escape(name) for name in filenames)
    # Match "(" + any path prefix (non-greedy) + filename + ")"
    return re.compile(r"\(([^()]*?/)?(" + escaped + r")\)")


def replace_in_text(text: str, mapping: dict[str, str], pattern: re.Pattern) -> tuple[str, int]:
    """Replace all matching image links in the given text. Returns the new
    text and the number of replacements made."""
    count = 0

    def _sub(m: re.Match) -> str:
        nonlocal count
        filename = m.group(2)
        uuid = mapping[filename]
        count += 1
        return f"({BASE_URL}/{uuid})"

    new_text = pattern.sub(_sub, text)
    return new_text, count


def process_file(path: Path, mapping: dict[str, str], pattern: re.Pattern, dry_run: bool) -> int:
    text = path.read_text(encoding="utf-8")
    new_text, count = replace_in_text(text, mapping, pattern)

    if count > 0:
        action = "Would update" if dry_run else "Updated"
        print(f"{action} {path} ({count} replacement(s))")
        if not dry_run:
            path.write_text(new_text, encoding="utf-8")

    return count


def main() -> None:
    parser = argparse.ArgumentParser(description="Replace image filenames with UUID URLs in markdown files.")
    parser.add_argument("csv_path", type=Path, help="Path to img.csv (columns: filename,uuid)")
    parser.add_argument("target_dir", type=Path, help="Directory to search recursively for .md files")
    parser.add_argument("--dry-run", action="store_true", help="Show what would change without writing files")
    args = parser.parse_args()

    if not args.csv_path.is_file():
        sys.exit(f"CSV file not found: {args.csv_path}")
    if not args.target_dir.is_dir():
        sys.exit(f"Target directory not found: {args.target_dir}")

    mapping = load_mapping(args.csv_path)
    if not mapping:
        sys.exit("No entries found in CSV file.")

    pattern = build_pattern(list(mapping.keys()))

    total_files = 0
    total_replacements = 0
    for md_path in sorted(args.target_dir.rglob("*.md")):
        count = process_file(md_path, mapping, pattern, args.dry_run)
        if count > 0:
            total_files += 1
            total_replacements += count

    mode = "[dry-run] " if args.dry_run else ""
    print(f"{mode}Done. {total_replacements} replacement(s) in {total_files} file(s).")


if __name__ == "__main__":
    main()