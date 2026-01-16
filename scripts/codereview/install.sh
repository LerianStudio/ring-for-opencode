#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_DIR="${SCRIPT_DIR}/bin"
CHECKSUMS_FILE="${SCRIPT_DIR}/checksums.txt"
INSTALL_TARGET="${1:-all}"

# List of all codereview binaries
BINARIES=(
    "scope-detector"
    "static-analysis"
    "ast-extractor"
    "call-graph"
    "data-flow"
    "compile-context"
    "run-all"
)

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[OK]${NC} $1"
}

# Verify SHA256 checksum of a binary
# Usage: verify_checksum "binary_path" "tool_name"
# Returns 0 if checksum matches or checksums file doesn't exist, 1 on mismatch
verify_checksum() {
    local binary="$1"
    local tool_name="$2"
    
    if [[ ! -f "$CHECKSUMS_FILE" ]]; then
        log_warn "Checksums file not found, skipping verification for $tool_name"
        return 0
    fi
    
    local expected
    expected=$(grep "^${tool_name}:" "$CHECKSUMS_FILE" 2>/dev/null | cut -d':' -f2 | tr -d ' ')
    
    if [[ -z "$expected" ]]; then
        log_warn "No checksum found for $tool_name in checksums.txt"
        return 0
    fi
    
    if [[ "$expected" == "PLACEHOLDER" ]]; then
        log_warn "Checksum for $tool_name is a placeholder, skipping verification"
        return 0
    fi
    
    local actual
    if command -v sha256sum &> /dev/null; then
        actual=$(sha256sum "$binary" 2>/dev/null | cut -d' ' -f1)
    elif command -v shasum &> /dev/null; then
        actual=$(shasum -a 256 "$binary" 2>/dev/null | cut -d' ' -f1)
    else
        log_warn "No sha256sum or shasum available, skipping checksum verification"
        return 0
    fi
    
    if [[ "$actual" != "$expected" ]]; then
        log_error "Checksum mismatch for $tool_name"
        log_error "  Expected: $expected"
        log_error "  Actual:   $actual"
        return 1
    fi
    
    log_success "Checksum verified for $tool_name"
    return 0
}

# Get binary path for a Go tool
get_go_binary_path() {
    local tool="$1"
    local gopath="${GOPATH:-$HOME/go}"
    echo "${gopath}/bin/${tool}"
}

# Install a tool with error handling
# Usage: install_tool "name" command args...
install_tool() {
    local name="$1"
    shift
    log_info "Installing $name..."
    if ! "$@"; then
        log_error "Failed to install $name"
        return 1
    fi
    log_success "$name installed"
}

install_go_tools() {
    log_info "Installing Go analysis tools..."

    # staticcheck - Go static analysis (pinned version)
    if command -v staticcheck &> /dev/null; then
        log_info "staticcheck already installed: $(staticcheck --version 2>&1 | head -1)"
    else
        install_tool "staticcheck" go install honnef.co/go/tools/cmd/staticcheck@2024.1.1 || return 1
        verify_checksum "$(get_go_binary_path staticcheck)" "staticcheck" || return 1
    fi

    # gosec - Go security scanner (pinned version)
    if command -v gosec &> /dev/null; then
        log_info "gosec already installed: $(gosec --version 2>&1 | head -1)"
    else
        install_tool "gosec" go install github.com/securego/gosec/v2/cmd/gosec@v2.22.0 || return 1
        verify_checksum "$(get_go_binary_path gosec)" "gosec" || return 1
    fi

    # golangci-lint - Go linters aggregator (pinned version)
    if command -v golangci-lint &> /dev/null; then
        log_info "golangci-lint already installed: $(golangci-lint --version 2>&1 | head -1)"
    else
        install_tool "golangci-lint" go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.63.4 || return 1
        verify_checksum "$(get_go_binary_path golangci-lint)" "golangci-lint" || return 1
    fi

    log_info "Go tools installation complete."
}

install_ts_tools() {
    log_info "Installing TypeScript analysis tools (local to project)..."

    # Check if npm is available
    if ! command -v npm &> /dev/null; then
        log_error "npm is not installed. Please install Node.js first."
        exit 1
    fi

    (  # START SUBSHELL - automatically restores directory on exit
        cd "${SCRIPT_DIR}"

        # Initialize package.json if it doesn't exist
        if [[ ! -f "package.json" ]]; then
            log_info "Initializing package.json..."
            npm init -y > /dev/null 2>&1
        fi

        # typescript - TypeScript compiler (pinned version, local install)
        if [[ -f "node_modules/.bin/tsc" ]]; then
            log_info "typescript already installed locally"
        else
            install_tool "typescript" npm install --save-dev typescript@5.7.2 || exit 1
        fi

        # dependency-cruiser - Dependency analysis (pinned version, local install)
        if [[ -f "node_modules/.bin/depcruise" ]]; then
            log_info "dependency-cruiser already installed locally"
        else
            install_tool "dependency-cruiser" npm install --save-dev dependency-cruiser@16.8.0 || exit 1
        fi

        # madge - Circular dependency detection (pinned version, local install)
        if [[ -f "node_modules/.bin/madge" ]]; then
            log_info "madge already installed locally"
        else
            install_tool "madge" npm install --save-dev madge@8.0.0 || exit 1
        fi

        log_info "TypeScript tools installation complete."
        log_info "Note: TypeScript tools are installed locally. Use npx or ./node_modules/.bin/ to run them."
    )  # END SUBSHELL
}

install_py_tools() {
    log_info "Installing Python analysis tools..."

    # Check if pip is available
    if ! command -v pip3 &> /dev/null && ! command -v pip &> /dev/null; then
        log_error "pip is not installed. Please install Python first."
        exit 1
    fi

    PIP_CMD="pip3"
    if ! command -v pip3 &> /dev/null; then
        PIP_CMD="pip"
    fi

    # Use --user flag when not in a virtual environment to avoid permission issues
    PIP_ARGS=()
    if [[ -z "${VIRTUAL_ENV:-}" ]]; then
        PIP_ARGS+=("--user")
    fi

    # ruff - Fast Python linter (pinned version)
    if command -v ruff &> /dev/null; then
        log_info "ruff already installed: $(ruff --version 2>&1)"
    else
        install_tool "ruff" "$PIP_CMD" install "${PIP_ARGS[@]}" ruff==0.8.4 || return 1
    fi

    # mypy - Static type checker (pinned version)
    if command -v mypy &> /dev/null; then
        log_info "mypy already installed: $(mypy --version 2>&1)"
    else
        install_tool "mypy" "$PIP_CMD" install "${PIP_ARGS[@]}" mypy==1.14.1 || return 1
    fi

    # pylint - Python linter (pinned version)
    if command -v pylint &> /dev/null; then
        log_info "pylint already installed: $(pylint --version 2>&1 | head -1)"
    else
        install_tool "pylint" "$PIP_CMD" install "${PIP_ARGS[@]}" pylint==3.3.3 || return 1
    fi

    # bandit - Security linter (pinned version)
    if command -v bandit &> /dev/null; then
        log_info "bandit already installed: $(bandit --version 2>&1 | head -1)"
    else
        install_tool "bandit" "$PIP_CMD" install "${PIP_ARGS[@]}" bandit==1.8.0 || return 1
    fi

    log_info "Python tools installation complete."
}

build_binaries() {
    log_info "Building codereview binaries..."

    mkdir -p "${BIN_DIR}"

    (  # START SUBSHELL - automatically restores directory on exit
        cd "${SCRIPT_DIR}"

        for binary in "${BINARIES[@]}"; do
            log_info "Building ${binary}..."
            go build -o "${BIN_DIR}/${binary}" "./cmd/${binary}"
        done

        # Make all binaries executable
        chmod +x "${BIN_DIR}"/*

        log_info "All binaries built successfully in ${BIN_DIR}/"
    )  # END SUBSHELL
}

verify_installation() {
    log_info "Verifying installation..."

    local all_ok=true

    # Verify Go tools
    log_info "Checking Go tools..."
    for tool in staticcheck gosec golangci-lint; do
        if command -v $tool &> /dev/null; then
            echo "  [OK] $tool"
        else
            echo "  [MISSING] $tool"
            all_ok=false
        fi
    done

    # Verify TypeScript tools (local installation)
    log_info "Checking TypeScript tools (local)..."
    for tool in tsc depcruise madge; do
        if [[ -f "${SCRIPT_DIR}/node_modules/.bin/${tool}" ]]; then
            echo "  [OK] $tool (local)"
        elif command -v $tool &> /dev/null; then
            echo "  [OK] $tool (global - consider local install)"
        else
            echo "  [MISSING] $tool"
            all_ok=false
        fi
    done

    # Verify Python tools
    log_info "Checking Python tools..."
    for tool in ruff mypy pylint bandit; do
        if command -v $tool &> /dev/null; then
            echo "  [OK] $tool"
        else
            echo "  [MISSING] $tool"
            all_ok=false
        fi
    done

    # Verify binaries
    log_info "Checking codereview binaries..."
    for binary in "${BINARIES[@]}"; do
        if [[ -x "${BIN_DIR}/${binary}" ]]; then
            echo "  [OK] ${binary}"
        else
            echo "  [MISSING] ${binary}"
            all_ok=false
        fi
    done

    if $all_ok; then
        log_info "All tools and binaries verified successfully!"
        return 0
    else
        log_warn "Some tools or binaries are missing. Run './install.sh all' to install everything."
        return 1
    fi
}

generate_checksums() {
    log_info "Generating checksums for installed Go tools..."
    
    local checksum_cmd
    if command -v sha256sum &> /dev/null; then
        checksum_cmd="sha256sum"
    elif command -v shasum &> /dev/null; then
        checksum_cmd="shasum -a 256"
    else
        log_error "No sha256sum or shasum available"
        return 1
    fi
    
    echo "# Checksums for codereview tool dependencies" > "$CHECKSUMS_FILE"
    echo "# Generated on $(date -u +"%Y-%m-%dT%H:%M:%SZ")" >> "$CHECKSUMS_FILE"
    echo "# Format: tool_name: sha256_checksum" >> "$CHECKSUMS_FILE"
    echo "#" >> "$CHECKSUMS_FILE"
    echo "# To update: run './install.sh generate-checksums' after verifying tool integrity" >> "$CHECKSUMS_FILE"
    echo "" >> "$CHECKSUMS_FILE"
    
    for tool in staticcheck gosec golangci-lint; do
        local binary_path
        binary_path=$(get_go_binary_path "$tool")
        if [[ -f "$binary_path" ]]; then
            local checksum
            checksum=$($checksum_cmd "$binary_path" | cut -d' ' -f1)
            echo "${tool}: ${checksum}" >> "$CHECKSUMS_FILE"
            log_info "  $tool: $checksum"
        else
            log_warn "  $tool: not found, skipping"
        fi
    done
    
    log_success "Checksums written to $CHECKSUMS_FILE"
}

# Main execution
case "$INSTALL_TARGET" in
    go)
        install_go_tools
        ;;
    ts)
        install_ts_tools
        ;;
    py)
        install_py_tools
        ;;
    build)
        build_binaries
        ;;
    all)
        install_go_tools
        install_ts_tools
        install_py_tools
        build_binaries
        verify_installation
        ;;
    verify)
        verify_installation
        ;;
    generate-checksums)
        generate_checksums
        ;;
    *)
        echo "Usage: $0 [go|ts|py|build|all|verify|generate-checksums]"
        echo ""
        echo "Commands:"
        echo "  go                  Install Go analysis tools (staticcheck, gosec, golangci-lint)"
        echo "  ts                  Install TypeScript tools locally (typescript, dependency-cruiser, madge)"
        echo "  py                  Install Python tools (ruff, mypy, pylint, bandit)"
        echo "  build               Build all codereview binaries"
        echo "  all                 Install all tools and build binaries (default)"
        echo "  verify              Verify all tools and binaries are installed"
        echo "  generate-checksums  Generate checksums.txt from currently installed Go tools"
        exit 1
        ;;
esac
