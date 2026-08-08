#!/usr/bin/env python3
"""PostToolUse(Bash) hook: when a release-ish command runs (git tag/push
--tags, goreleaser), remind the agent to run the update-architecture-docs
skill so the self-hosted docs ship current with the release."""
import json
import re
import sys

try:
    data = json.load(sys.stdin)
except Exception:
    sys.exit(0)

cmd = (data.get("tool_input") or {}).get("command", "")
release_patterns = [
    r"\bgit\s+tag\b.*\bv?\d+\.\d+",
    r"\bgit\s+push\b.*--tags\b",
    r"\bgit\s+push\b.*\brefs/tags/",
    r"\bgoreleaser\b",
]
if not any(re.search(p, cmd) for p in release_patterns):
    sys.exit(0)

print(json.dumps({
    "hookSpecificOutput": {
        "hookEventName": "PostToolUse",
        "additionalContext": (
            "Release activity detected. If not already done for this release, "
            "run the update-architecture-docs skill now: refresh "
            "testdata/demo/byakugan/architecture.html against the shipped code "
            "and add prd-<slug>.html pages for major features, so the "
            "self-hosted docs match the release being cut."
        ),
    }
}))
