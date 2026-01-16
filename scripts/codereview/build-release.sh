#!/bin/bash
#
# build-release.sh - Multi-platform build script for codereview binaries
#
# Builds all codereview tools for darwin/linux on amd64/arm64 architectures.
# Output is placed in default/lib/codereview/bin/{os}_{arch}/
#
# Usage:
#   ./build-release.sh [--clean] [--platform=<os/arch>]
#
# Options:
#   --clean             Remove existing binaries before building
#   --platform=<os/arch>  Build only for specific platform (e.g., --platform=darwin/arm64)
#
# Examples:
#   ./build-release.sh                    # Build all platforms
#   ./build-release.sh --clean            # Clean and rebuild all
#   ./build-release.sh --platform=linux/amd64  # Build only linux/amd64
#

set -e

# Script directory (scripts/codereview/)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Repository root
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# Output directory for release binaries
OUTPUT_DIR="${SCRIPT_DIR}/bin/releases"

# List of all codereview binaries to build
BINARIES=(
    "run-all"
    "scope-detector"
    "static-analysis"
    "ast-extractor"
    "call-graph"
    "data-flow"
    "compile-context"
)

# Target platforms: os/arch
PLATFORMS=(
    "darwin/amd64"
    "darwin/arm64"
    "linux/amd64"
    "linux/arm64"
)

# Build flags for smaller binaries (strip debug symbols)
LDFLAGS="-s -w"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Counters for summary
TOTAL_BINARIES=0
SUCCESSFUL_BUILDS=0
FAILED_BUILDS=0

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[OK]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_header() {
    echo -e "\n${BOLD}${CYAN}=== $1 ===${NC}\n"
}

# Parse command line arguments
CLEAN=false
SINGLE_PLATFORM=""

for arg in "$@"; do
    case $arg in
        --clean)
            CLEAN=true
            shift
            ;;
        --platform=*)
            SINGLE_PLATFORM="${arg#*=}"
            shift
            ;;
        --help|-h)
            echo "Usage: $0 [--clean] [--platform=<os/arch>]"
            echo ""
            echo "Options:"
            echo "  --clean               Remove existing binaries before building"
            echo "  --platform=<os/arch>  Build only for specific platform"
            echo ""
            echo "Platforms: darwin/amd64, darwin/arm64, linux/amd64, linux/arm64"
            echo ""
            echo "Examples:"
            echo "  $0                           # Build all platforms"
            echo "  $0 --clean                   # Clean and rebuild all"
            echo "  $0 --platform=linux/amd64    # Build only linux/amd64"
            exit 0
            ;;
        *)
            log_error "Unknown argument: $arg"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# If single platform specified, validate it
if [[ -n "$SINGLE_PLATFORM" ]]; then
    valid_platform=false
    for platform in "${PLATFORMS[@]}"; do
        if [[ "$platform" == "$SINGLE_PLATFORM" ]]; then
            valid_platform=true
            break
        fi
    done
    if [[ "$valid_platform" == "false" ]]; then
        log_error "Invalid platform: $SINGLE_PLATFORM"
        echo "Valid platforms: ${PLATFORMS[*]}"
        exit 1
    fi
    PLATFORMS=("$SINGLE_PLATFORM")
fi

# Check Go is installed
check_prerequisites() {
    log_header "Checking Prerequisites"

    if ! command -v go &> /dev/null; then
        log_error "Go is not installed. Please install Go 1.21+ first."
        exit 1
    fi

    go_version=$(go version | grep -oE 'go[0-9]+\.[0-9]+' | head -1)
    log_success "Go installed: $go_version"

    # Verify we can find the source code
    if [[ ! -d "${SCRIPT_DIR}/cmd" ]]; then
        log_error "Source directory not found: ${SCRIPT_DIR}/cmd"
        exit 1
    fi

    log_success "Source directory found: ${SCRIPT_DIR}/cmd"
}

# Clean existing binaries
clean_binaries() {
    log_header "Cleaning Existing Binaries"

    if [[ -d "$OUTPUT_DIR" ]]; then
        rm -rf "$OUTPUT_DIR"
        log_success "Removed: $OUTPUT_DIR"
    else
        log_info "Nothing to clean"
    fi
}

# Build binaries for a specific platform
build_platform() {
    local os=$1
    local arch=$2
    local platform_dir="${OUTPUT_DIR}/${os}_${arch}"

    log_header "Building for ${os}/${arch}"

    mkdir -p "$platform_dir"

    # Change to script directory for Go module context
    pushd "$SCRIPT_DIR" > /dev/null

    for binary in "${BINARIES[@]}"; do
        TOTAL_BINARIES=$((TOTAL_BINARIES + 1))

        local src_dir="./cmd/${binary}"
        local output_path="${platform_dir}/${binary}"

        if [[ ! -d "$src_dir" ]]; then
            log_error "Source not found: $src_dir"
            FAILED_BUILDS=$((FAILED_BUILDS + 1))
            continue
        fi

        echo -n "  Building ${binary}... "

        if GOOS="$os" GOARCH="$arch" go build \
            -ldflags="$LDFLAGS" \
            -o "$output_path" \
            "$src_dir" 2>/dev/null; then

            # Make executable (for same-platform builds)
            chmod +x "$output_path"

            # Get file size
            local size
            if [[ "$OSTYPE" == "darwin"* ]]; then
                size=$(stat -f%z "$output_path" 2>/dev/null || echo "?")
            else
                size=$(stat --printf="%s" "$output_path" 2>/dev/null || echo "?")
            fi

            # Human readable size
            local human_size
            if [[ "$size" != "?" ]]; then
                if [[ $size -ge 1048576 ]]; then
                    human_size="$(echo "scale=1; $size/1048576" | bc)M"
                elif [[ $size -ge 1024 ]]; then
                    human_size="$(echo "scale=1; $size/1024" | bc)K"
                else
                    human_size="${size}B"
                fi
            else
                human_size="?"
            fi

            echo -e "${GREEN}OK${NC} (${human_size})"
            SUCCESSFUL_BUILDS=$((SUCCESSFUL_BUILDS + 1))
        else
            echo -e "${RED}FAILED${NC}"
            FAILED_BUILDS=$((FAILED_BUILDS + 1))
        fi
    done

    # Return to original directory
    popd > /dev/null
}

# Print summary of all builds
print_summary() {
    log_header "Build Summary"

    echo -e "Total binaries:     ${TOTAL_BINARIES}"
    echo -e "Successful builds:  ${GREEN}${SUCCESSFUL_BUILDS}${NC}"

    if [[ $FAILED_BUILDS -gt 0 ]]; then
        echo -e "Failed builds:      ${RED}${FAILED_BUILDS}${NC}"
    else
        echo -e "Failed builds:      ${FAILED_BUILDS}"
    fi

    echo ""

    # Show output directory structure
    log_info "Output directory: ${OUTPUT_DIR}"
    echo ""

    if [[ -d "$OUTPUT_DIR" ]]; then
        # List all platforms and their sizes
        for platform_dir in "$OUTPUT_DIR"/*/; do
            if [[ -d "$platform_dir" ]]; then
                local platform_name
                platform_name=$(basename "$platform_dir")
                local total_size=0
                local file_count=0

                echo -e "  ${BOLD}${platform_name}/${NC}"

                for binary in "$platform_dir"*; do
                    if [[ -f "$binary" ]]; then
                        local name
                        name=$(basename "$binary")
                        local size
                        if [[ "$OSTYPE" == "darwin"* ]]; then
                            size=$(stat -f%z "$binary" 2>/dev/null || echo "0")
                        else
                            size=$(stat --printf="%s" "$binary" 2>/dev/null || echo "0")
                        fi

                        # Human readable size
                        local human_size
                        if [[ $size -ge 1048576 ]]; then
                            human_size="$(echo "scale=1; $size/1048576" | bc)M"
                        elif [[ $size -ge 1024 ]]; then
                            human_size="$(echo "scale=1; $size/1024" | bc)K"
                        else
                            human_size="${size}B"
                        fi

                        printf "    %-20s %8s\n" "$name" "$human_size"
                        total_size=$((total_size + size))
                        file_count=$((file_count + 1))
                    fi
                done

                # Platform total
                local total_human
                if [[ $total_size -ge 1048576 ]]; then
                    total_human="$(echo "scale=1; $total_size/1048576" | bc)M"
                else
                    total_human="$(echo "scale=1; $total_size/1024" | bc)K"
                fi
                echo -e "    ${CYAN}Total: ${file_count} files, ${total_human}${NC}"
                echo ""
            fi
        done
    fi

    # Final status
    if [[ $FAILED_BUILDS -eq 0 ]]; then
        echo -e "${GREEN}${BOLD}Build completed successfully!${NC}"
        return 0
    else
        echo -e "${RED}${BOLD}Build completed with errors.${NC}"
        return 1
    fi
}

# Main execution
main() {
    echo -e "${BOLD}${CYAN}"
    echo "=============================================="
    echo "  Ring Codereview Multi-Platform Build"
    echo "=============================================="
    echo -e "${NC}"

    check_prerequisites

    if [[ "$CLEAN" == "true" ]]; then
        clean_binaries
    fi

    # Build for each platform
    for platform in "${PLATFORMS[@]}"; do
        IFS='/' read -r os arch <<< "$platform"
        build_platform "$os" "$arch"
    done

    # Generate checksums for release verification
    log_info "=== Generating Checksums ==="

    # Determine checksum command (macOS uses shasum, Linux uses sha256sum)
    if command -v sha256sum &> /dev/null; then
        CHECKSUM_CMD="sha256sum"
    elif command -v shasum &> /dev/null; then
        CHECKSUM_CMD="shasum -a 256"
    else
        log_error "No sha256sum or shasum available for checksum generation"
        exit 1
    fi
    log_info "Using checksum command: $CHECKSUM_CMD"

    for platform_dir in "${OUTPUT_DIR}"/*/; do
        platform=$(basename "$platform_dir")
        checksum_file="${platform_dir}CHECKSUMS.sha256"

        # Delete existing checksum file before regenerating to avoid self-inclusion
        rm -f "$checksum_file"

        # Generate checksums for all binaries in this platform directory
        if ! (cd "$platform_dir" && $CHECKSUM_CMD * > CHECKSUMS.sha256); then
            log_error "Failed to generate checksums for $platform"
            FAILED_BUILDS=$((FAILED_BUILDS + 1))
            continue
        fi

        if [[ -f "$checksum_file" ]]; then
            log_success "Generated checksums for $platform"
        else
            log_warn "Failed to generate checksums for $platform"
        fi
    done

    # Generate combined checksums file at root
    root_checksum="${OUTPUT_DIR}/CHECKSUMS.sha256"

    # Delete existing root checksum file before regenerating
    rm -f "$root_checksum"

    # Generate checksums for all platform binaries
    root_checksum_failed=false
    {
        for platform_dir in "${OUTPUT_DIR}"/*/; do
            platform=$(basename "$platform_dir")
            for binary in "${platform_dir}"*; do
                if [[ -f "$binary" && -x "$binary" ]]; then
                    # Include platform in path for uniqueness
                    if ! (cd "${OUTPUT_DIR}" && $CHECKSUM_CMD "${platform}/$(basename "$binary")"); then
                        root_checksum_failed=true
                    fi
                fi
            done
        done
    } > "$root_checksum"

    if [[ "$root_checksum_failed" == "true" ]]; then
        log_error "Some checksums failed to generate for root CHECKSUMS.sha256"
    elif [[ -f "$root_checksum" ]]; then
        log_success "Generated root CHECKSUMS.sha256 with $(wc -l < "$root_checksum" | tr -d ' ') entries"
    else
        log_error "Failed to generate root CHECKSUMS.sha256"
    fi

    print_summary
}

# Run main
main
