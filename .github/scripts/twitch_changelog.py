#!/usr/bin/env python3
"""Find new entries in the Twitch developer changelog RSS feed.

Reads the feed, compares each item's guid with the guids already seen (one per
line in the state file), and writes the new entries as Markdown for an issue
body. The state file is rewritten with every guid in the feed plus those seen
before. On the first run (no state file) only entries published in the last
BOOTSTRAP_DAYS days count as new, so migrating to this monitor doesn't report
the whole history.

Outputs (GITHUB_OUTPUT when set): count=<new entries>.
"""

import argparse
import email.utils
import html
import os
import re
import sys
import urllib.request
import xml.etree.ElementTree as ET
from datetime import datetime, timedelta, timezone
from html.parser import HTMLParser

FEED_URL = "https://dev.twitch.tv/docs/rss/change-log.xml"
BOOTSTRAP_DAYS = 14
MAX_BODY_CHARS = 60000  # GitHub caps issue and comment bodies at 65536


class _Text(HTMLParser):
    """Renders the feed's description HTML as plain Markdown-ish text."""

    def __init__(self):
        super().__init__(convert_charrefs=True)
        self.out = []
        self.href = None

    def handle_starttag(self, tag, attrs):
        if tag == "li":
            self.out.append("\n- ")
        elif tag in ("p", "br", "ul", "ol"):
            self.out.append("\n")
        elif tag == "code":
            self.out.append("`")
        elif tag == "a":
            self.href = dict(attrs).get("href")
            self.out.append("[")

    def handle_endtag(self, tag):
        if tag == "code":
            self.out.append("`")
        elif tag == "a":
            self.out.append(f"]({self.href})" if self.href else "]")
            self.href = None
        elif tag == "p":
            self.out.append("\n")

    def handle_data(self, data):
        self.out.append(re.sub(r"\s+", " ", data))

    def text(self):
        lines = (" ".join(line.split()) for line in "".join(self.out).splitlines())
        return "\n".join(line for line in lines if line)


def to_text(fragment):
    parser = _Text()
    parser.feed(html.unescape(fragment) if "<" not in fragment else fragment)
    parser.close()
    return parser.text()


def parse_feed(xml_bytes):
    root = ET.fromstring(xml_bytes)
    items = []
    for item in root.iter("item"):
        guid = (item.findtext("guid") or item.findtext("link") or "").strip()
        if not guid:
            continue
        published = None
        if item.findtext("pubDate"):
            try:
                published = email.utils.parsedate_to_datetime(item.findtext("pubDate"))
            except (TypeError, ValueError):
                published = None
        items.append({
            "guid": guid,
            "title": (item.findtext("title") or guid).strip(),
            "link": (item.findtext("link") or "").strip(),
            "published": published,
            "text": to_text(item.findtext("description") or ""),
        })
    return items


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--state", required=True, help="file of seen guids, one per line")
    ap.add_argument("--body", required=True, help="where to write the Markdown for new entries")
    ap.add_argument("--feed-file", help="read the feed from this file instead of fetching it")
    args = ap.parse_args()

    if args.feed_file:
        with open(args.feed_file, "rb") as f:
            data = f.read()
    else:
        req = urllib.request.Request(FEED_URL, headers={"User-Agent": "kappopher-changelog-monitor"})
        with urllib.request.urlopen(req, timeout=30) as resp:
            data = resp.read()
    items = parse_feed(data)
    if not items:
        sys.exit("feed has no items; refusing to update state")

    seen = None
    if os.path.exists(args.state):
        with open(args.state) as f:
            seen = {line.strip() for line in f if line.strip()}

    if seen is None:
        cutoff = datetime.now(timezone.utc) - timedelta(days=BOOTSTRAP_DAYS)
        new = [i for i in items if i["published"] and i["published"] >= cutoff]
        seen = set()
    else:
        new = [i for i in items if i["guid"] not in seen]

    parts = []
    for item in new:
        heading = f"### [{item['title']}]({item['link']})" if item["link"] else f"### {item['title']}"
        parts.append(f"{heading}\n\n{item['text']}\n")
    body = "\n".join(parts)
    if len(body) > MAX_BODY_CHARS:
        body = body[:MAX_BODY_CHARS] + "\n\n…(truncated; see the full changelog)\n"
    with open(args.body, "w") as f:
        f.write(body)

    os.makedirs(os.path.dirname(os.path.abspath(args.state)), exist_ok=True)
    with open(args.state, "w") as f:
        f.write("\n".join(sorted(seen | {i["guid"] for i in items})) + "\n")

    print(f"{len(items)} feed items, {len(new)} new")
    if os.environ.get("GITHUB_OUTPUT"):
        with open(os.environ["GITHUB_OUTPUT"], "a") as f:
            f.write(f"count={len(new)}\n")


if __name__ == "__main__":
    main()
