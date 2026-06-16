#!/usr/bin/env bash
# sync.sh — Sync canonical skill to platform-specific directories
#
# Usage:
#   ./docs/skills/sync.sh              # sync all skills
#   ./docs/skills/sync.sh lgh-actiond  # sync one skill
#
# Canonical source: docs/skills/<name>/SKILL.md + references/*
# Targets:
#   - ~/.hermes/skills/devops/<name>/
#   - ~/.codex/skills/<name>/
#
# If a target is a symlink (preferred), it's skipped — the canonical source
# is already directly accessible. The script only copies when the target is
# a real directory (e.g. on a machine without symlink setup).
#
# Add new targets by appending to the TARGETS array below.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SKILLS_DIR="$SCRIPT_DIR"

# Platform targets: "label|base_dir"
TARGETS=(
  "hermes|$HOME/.hermes/skills/devops"
  "codex|$HOME/.codex/skills"
)

sync_skill() {
  local name="$1"
  local src="$SKILLS_DIR/$name"

  if [[ ! -f "$src/SKILL.md" ]]; then
    echo "⚠️  No SKILL.md found at $src/SKILL.md — skipping"
    return 1
  fi

  for target in "${TARGETS[@]}"; do
    local label="${target%%|*}"
    local base="${target##*|}"
    local dest="$base/$name"

    # If destination is a symlink, it already points to canonical — skip
    if [[ -L "$dest" ]]; then
      echo "⏭️  $name → $label (symlink, already in sync)"
      continue
    fi

    mkdir -p "$dest"

    # Copy SKILL.md
    cp "$src/SKILL.md" "$dest/SKILL.md"

    # Copy references/ if present
    if [[ -d "$src/references" ]]; then
      mkdir -p "$dest/references"
      cp -r "$src/references/"* "$dest/references/" 2>/dev/null || true
    fi

    echo "✅ $name → $label ($dest)"
  done
}

main() {
  echo "📦 Skill Sync — canonical source: $SKILLS_DIR"
  echo ""

  if [[ $# -gt 0 ]]; then
    for name in "$@"; do
      sync_skill "$name"
    done
  else
    local count=0
    for dir in "$SKILLS_DIR"/*/; do
      [[ -d "$dir" ]] || continue
      local name="$(basename "$dir")"
      sync_skill "$name" && ((count++))
    done
    if [[ $count -eq 0 ]]; then
      echo "⚠️  No skills found in $SKILLS_DIR"
    fi
  fi

  echo ""
  echo "Done."
}

main "$@"
