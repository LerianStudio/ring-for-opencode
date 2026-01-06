#!/usr/bin/env bash
set -euo pipefail

# Ring → OpenCode installer
#
# Installs Ring's OpenCode plugins/skills/commands/agents into ~/.config/opencode
# WITHOUT deleting any existing user content.
#
# Behavior:
# - Copies (overwrites) only the Ring-managed files that share exact paths
# - NEVER deletes unknown files in the target directory
# - Backs up any overwritten files into ~/.config/opencode/.ring-backups/<timestamp>/
# - Merges required dependencies into ~/.config/opencode/package.json (preserving existing fields)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SOURCE_ROOT="$SCRIPT_DIR/.opencode"
TARGET_ROOT="${OPENCODE_CONFIG_DIR:-"$HOME/.config/opencode"}"

if [[ ! -d "$SOURCE_ROOT" ]]; then
  echo "ERROR: Source directory not found: $SOURCE_ROOT" >&2
  echo "Expected this script to live at: opencode/installer.sh" >&2
  echo "And source at: opencode/.opencode/" >&2
  exit 1
fi

mkdir -p "$TARGET_ROOT"

STAMP="$(date -u +"%Y%m%dT%H%M%SZ")"
BACKUP_DIR="$TARGET_ROOT/.ring-backups/$STAMP"
mkdir -p "$BACKUP_DIR"

backup_if_exists() {
  local rel="$1"
  local src="$SOURCE_ROOT/$rel"
  local dst="$TARGET_ROOT/$rel"

  # Only consider files we manage (exist in source)
  [[ -e "$src" ]] || return 0

  if [[ -e "$dst" ]]; then
    mkdir -p "$(dirname "$BACKUP_DIR/$rel")"
    cp -a "$dst" "$BACKUP_DIR/$rel"
  fi
}

copy_tree_no_delete() {
  local rel_dir="$1"

  # Ensure destination exists
  mkdir -p "$TARGET_ROOT/$rel_dir"

  # rsync WITHOUT --delete: preserves user content
  # -a: archive (permissions, times)
  # --checksum: safer overwrites when timestamps differ
  rsync -a --checksum "$SOURCE_ROOT/$rel_dir/" "$TARGET_ROOT/$rel_dir/"
}

# Backup any files we might overwrite
backup_if_exists "plugin/index.ts"
backup_if_exists "plugin/README.md"
backup_if_exists "plugin/test-plugins.ts"
backup_if_exists "package.json"

# Back up all Ring plugin .ts files (best-effort)
if [[ -d "$SOURCE_ROOT/plugin" ]]; then
  while IFS= read -r -d '' f; do
    rel="${f#"$SOURCE_ROOT/"}"
    backup_if_exists "$rel"
  done < <(find "$SOURCE_ROOT/plugin" -type f -name "*.ts" -print0)
fi

# Copy trees (no deletes)
for d in plugin skill command agent; do
  if [[ -d "$SOURCE_ROOT/$d" ]]; then
    copy_tree_no_delete "$d"
  fi
done

# Ensure state dir exists (no overwrite)
mkdir -p "$TARGET_ROOT/state"

# Merge package.json deps (preserves existing user package.json fields)
REQUIRED_DEPS_JSON='{
  "dependencies": {
    "@opencode-ai/plugin": "1.1.3",
    "better-sqlite3": "9.0.0"
  },
  "devDependencies": {
    "@types/better-sqlite3": "7.6.13",
    "@types/node": "25.0.3",
    "typescript": "5.9.3"
  }
}'

REQUIRED_DEPS_JSON="$REQUIRED_DEPS_JSON" node - <<'NODE'
const fs = require('fs');
const path = require('path');

const targetRoot = process.env.OPENCODE_CONFIG_DIR || path.join(process.env.HOME, '.config/opencode');
const pkgPath = path.join(targetRoot, 'package.json');
const required = JSON.parse(process.env.REQUIRED_DEPS_JSON);

function mergeSection(target, sectionName) {
  const src = required[sectionName] || {};
  const dst = target[sectionName] || {};
  target[sectionName] = { ...dst, ...src };
}

let pkg = {};
if (fs.existsSync(pkgPath)) {
  try {
    pkg = JSON.parse(fs.readFileSync(pkgPath, 'utf8'));
  } catch (e) {
    console.error(`ERROR: Failed to parse existing ${pkgPath}: ${e}`);
    process.exit(1);
  }
}

mergeSection(pkg, 'dependencies');
mergeSection(pkg, 'devDependencies');

// Ensure package.json is valid even if it didn't exist
pkg.name ??= 'opencode-config';
pkg.private ??= true;

fs.writeFileSync(pkgPath, JSON.stringify(pkg, null, 2) + '\n', { encoding: 'utf8', mode: 0o600 });
console.log(`Updated ${pkgPath}`);
NODE

# Install deps
if command -v bun >/dev/null 2>&1; then
  echo "Installing dependencies with bun..."
  (cd "$TARGET_ROOT" && CXXFLAGS='-std=c++20' bun install)

  # Run tests if present
  if [[ -f "$TARGET_ROOT/plugin/test-plugins.ts" ]]; then
    echo "Running plugin tests..."
    (cd "$TARGET_ROOT" && bun plugin/test-plugins.ts)
  fi
else
  echo "WARN: bun not found; skipping dependency install and tests." >&2
fi

echo "Install complete. Backup (if any) at: $BACKUP_DIR"
