#!/usr/bin/env bash
set -euo pipefail

# Ring → OpenCode installer
#
# Installs Ring's unified plugin architecture into ~/.config/opencode
# WITHOUT deleting any existing user content.
#
# Architecture:
# - Unified plugin system (RingUnifiedPlugin) with hook-based architecture
# - Config injection for agents/skills/commands via ring-config.json
# - Background task management with schema validation
#
# Behavior:
# - Copies (overwrites) only the Ring-managed files that share exact paths
# - NEVER deletes unknown files in the target directory
# - Backs up any overwritten files into ~/.config/opencode/.ring-backups/<timestamp>/
# - Merges required dependencies into ~/.config/opencode/package.json (preserving existing fields)
# - Copies JSON schema files for IDE autocomplete support

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SOURCE_ASSETS="$SCRIPT_DIR/assets"
SOURCE_PLUGIN="$SCRIPT_DIR/plugin"
TARGET_ROOT="${OPENCODE_CONFIG_DIR:-"$HOME/.config/opencode"}"

# Node version check - require 18-24, warn if 25+
check_node_version() {
  if ! command -v node >/dev/null 2>&1; then
    echo "WARN: Node.js not found. Will attempt to use bun for installation." >&2
    return 0
  fi

  local node_version
  node_version=$(node -v | sed 's/^v//' | cut -d. -f1)

  if [[ "$node_version" -lt 18 ]]; then
    echo "ERROR: Node.js version $node_version is too old. Requires Node 18-24 (LTS)." >&2
    echo "Please install Node 22 LTS: https://nodejs.org/" >&2
    exit 1
  fi

  if [[ "$node_version" -ge 25 ]]; then
    echo "WARN: Node.js $node_version detected. This installer is tested with Node 18-24." >&2
    echo "WARN: better-sqlite3 may fail to build on Node 25+." >&2
    echo "WARN: Consider using Node 22 LTS for best compatibility." >&2
    echo ""
  fi
}

check_node_version

if [[ ! -d "$SOURCE_ASSETS" ]]; then
  echo "ERROR: Source directory not found: $SOURCE_ASSETS" >&2
  echo "Expected this script to live at: ring-for-opencode/installer.sh" >&2
  echo "And source at: ring-for-opencode/assets/" >&2
  exit 1
fi

if [[ ! -d "$SOURCE_PLUGIN" ]]; then
  echo "ERROR: Plugin directory not found: $SOURCE_PLUGIN" >&2
  echo "Expected plugin at: ring-for-opencode/plugin/" >&2
  exit 1
fi

mkdir -p "$TARGET_ROOT"

STAMP="$(date -u +"%Y%m%dT%H%M%SZ")"
BACKUP_DIR="$TARGET_ROOT/.ring-backups/$STAMP"
mkdir -p "$BACKUP_DIR"

backup_if_exists() {
  local rel="$1"
  local source_base="${2:-$SOURCE_ASSETS}"
  local src="$source_base/$rel"
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
  local source_base="${2:-$SOURCE_ASSETS}"

  # Ensure destination exists
  mkdir -p "$TARGET_ROOT/$rel_dir"

  # rsync WITHOUT --delete: preserves user content
  # -a: archive (permissions, times)
  # --checksum: safer overwrites when timestamps differ
  rsync -a --checksum "$source_base/$rel_dir/" "$TARGET_ROOT/$rel_dir/"
}

copy_file() {
  local rel="$1"
  local source_base="${2:-$SOURCE_ASSETS}"
  local src="$source_base/$rel"
  local dst="$TARGET_ROOT/$rel"

  if [[ -e "$src" ]]; then
    mkdir -p "$(dirname "$dst")"
    cp -a "$src" "$dst"
    echo "Copied: $rel"
  fi
}

backup_root_config() {
  local filename="$1"
  local dest_name="${2:-$1}"
  local src="$SCRIPT_DIR/$filename"
  local dst="$TARGET_ROOT/ring/$dest_name"

  [[ -e "$src" ]] || return 0

  if [[ -e "$dst" ]]; then
    mkdir -p "$BACKUP_DIR/ring"
    cp -a "$dst" "$BACKUP_DIR/ring/$dest_name"
  fi
}

copy_root_config() {
  local filename="$1"
  local dest_name="${2:-$1}"
  local src="$SCRIPT_DIR/$filename"
  local dst="$TARGET_ROOT/ring/$dest_name"

  if [[ -e "$src" ]]; then
    mkdir -p "$(dirname "$dst")"
    cp -a "$src" "$dst"
    echo "Copied: ring/$dest_name"
  fi
}

# Backup any files we might overwrite
# Plugin files (from root plugin/)
backup_if_exists "plugin/index.ts" "$SOURCE_PLUGIN/.."
backup_if_exists "plugin/ring-plugin.ts" "$SOURCE_PLUGIN/.."
backup_if_exists "plugin/ring-unified.ts" "$SOURCE_PLUGIN/.."
backup_if_exists "package.json"
backup_root_config "ring.jsonc" "config.jsonc"

# Back up all Ring plugin .ts files (best-effort)
if [[ -d "$SOURCE_PLUGIN" ]]; then
  while IFS= read -r -d '' f; do
    rel="${f#"$SCRIPT_DIR/"}"
    backup_if_exists "$rel" "$SCRIPT_DIR"
  done < <(find "$SOURCE_PLUGIN" -type f -name "*.ts" -print0)
fi

# Copy plugin directory from root level
echo "Copying plugin directory..."
copy_tree_no_delete "plugin" "$SCRIPT_DIR"

# Copy skill/command/agent/standards/templates from assets
echo "Copying skill/command/agent/standards/templates directories..."
for d in skill command agent standards templates; do
  if [[ -d "$SOURCE_ASSETS/$d" ]]; then
    copy_tree_no_delete "$d" "$SOURCE_ASSETS"
  fi
done

# Copy schema files for IDE autocomplete
echo "Copying schema files..."
copy_file "ring-config.schema.json" "$SOURCE_ASSETS"
copy_file "background-tasks.schema.json" "$SOURCE_ASSETS"

# Copy config templates
copy_root_config "ring.jsonc" "config.jsonc"

# Ensure global state dir exists in user config (no overwrite)
# Note: Project-level state is in <project>/.opencode/state/ and created dynamically
mkdir -p "$TARGET_ROOT/state"

# Merge package.json deps (preserves existing user package.json fields)
REQUIRED_DEPS_JSON='{
  "dependencies": {
    "@opencode-ai/plugin": "1.1.3",
    "better-sqlite3": "12.6.0",
    "zod": "^4.1.8",
    "jsonc-parser": "^3.3.1",
    "@clack/prompts": "^0.11.0",
    "picocolors": "^1.1.1",
    "commander": "^14.0.2"
  },
  "devDependencies": {
    "@types/better-sqlite3": "7.6.13",
    "@types/node": "22.19.5",
    "typescript": "5.9.3",
    "@biomejs/biome": "^1.9.4"
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
  if ! (cd "$TARGET_ROOT" && CXXFLAGS='-std=c++20' bun install); then
    echo "" >&2
    echo "ERROR: bun install failed." >&2
    echo "" >&2
    echo "Common causes:" >&2
    echo "  - Node.js version incompatibility (better-sqlite3 requires Node 18-24)" >&2
    echo "  - Missing C++ build tools" >&2
    echo "" >&2
    echo "Recommended fix:" >&2
    echo "  1. Install Node 22 LTS: https://nodejs.org/" >&2
    echo "  2. Ensure you have build tools installed:" >&2
    echo "     - macOS: xcode-select --install" >&2
    echo "     - Linux: apt-get install build-essential python3" >&2
    echo "  3. Re-run this installer" >&2
    exit 1
  fi

else
  echo "WARN: bun not found; skipping dependency install." >&2
fi

echo ""
echo "=========================================="
echo "  Ring for OpenCode - Install Complete"
echo "=========================================="
echo ""
echo "Installed components:"
echo "  - RingUnifiedPlugin (unified plugin with hook-based architecture)"
echo "  - Skills, commands, agents, standards, and templates from assets/"
echo "  - JSON schemas for IDE autocomplete"
echo ""
echo "Backup location (if any): $BACKUP_DIR"
echo ""
echo "To verify installation:"
echo "  1. Start OpenCode in your project directory"
echo "  2. Check that Ring skills appear in the command palette"
echo "  3. Create a ring-config.json for custom configuration"
echo ""
echo "For more info, see: https://github.com/fredcamaral/ring-for-opencode"
echo ""
