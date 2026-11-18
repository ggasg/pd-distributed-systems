"""Print the planned core dates from Obsidian frontmatter."""
import argparse
from datetime import date, timedelta
from pathlib import Path


def frontmatter(path):
    lines = path.read_text(encoding="utf-8").splitlines()
    if not lines or lines[0] != "---":
        return {}
    end = lines.index("---", 1)
    return dict(line.split(": ", 1) for line in lines[1:end] if ": " in line)


def main():
    vault = Path(__file__).resolve().parent.parent
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--start", type=date.fromisoformat)
    args = parser.parse_args()
    start = args.start or date.fromisoformat(
        frontmatter(vault / "config.md")["start_date"].strip('"')
    )
    core = []
    for note in (vault / "weeks").glob("*.md"):
        fields = frontmatter(note)
        if fields.get("track") == "core" and fields.get("week_number"):
            core.append((int(fields["week_number"]), fields["title"].strip('"')))
    for slot, title in sorted(core):
        first = start + timedelta(weeks=slot - 1)
        last = first + timedelta(days=6)
        print(f"{first} to {last}  {title}")


if __name__ == "__main__":
    main()
