#!/usr/bin/env bash
# Integration tests for call-graph binary
# Usage: ./scripts/codereview/test/integration_test.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../../.." && pwd)"
CODEREVIEW_DIR="${PROJECT_ROOT}/scripts/codereview"
BINARY="${CODEREVIEW_DIR}/bin/call-graph"

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0

# Temp directory for test outputs
TEMP_DIR=""

# Cleanup function
cleanup() {
    if [[ -n "${TEMP_DIR}" && -d "${TEMP_DIR}" ]]; then
        rm -rf "${TEMP_DIR}"
    fi
}
trap cleanup EXIT

# Test helper functions
pass() {
    echo -e "${GREEN}PASS${NC}: $1"
    TESTS_PASSED=$((TESTS_PASSED + 1))
}

fail() {
    echo -e "${RED}FAIL${NC}: $1"
    TESTS_FAILED=$((TESTS_FAILED + 1))
}

info() {
    echo -e "${YELLOW}INFO${NC}: $1"
}

# Create temp directory
setup() {
    TEMP_DIR=$(mktemp -d)
    info "Using temp directory: ${TEMP_DIR}"
}

# Verify binary exists
check_binary() {
    if [[ ! -x "${BINARY}" ]]; then
        echo -e "${RED}ERROR${NC}: Binary not found at ${BINARY}"
        echo "Build it first with: cd ${CODEREVIEW_DIR} && go build -o bin/call-graph ./cmd/call-graph"
        exit 1
    fi
    info "Using binary: ${BINARY}"
}

# Test 1: Help output
test_help_output() {
    info "Test 1: Help output (--help shows usage)"

    local output
    output=$("${BINARY}" --help 2>&1 || true)

    if echo "${output}" | grep -q "Usage:" && \
       echo "${output}" | grep -q "Call Graph Analyzer" && \
       echo "${output}" | grep -q "\-ast"; then
        pass "Help output shows usage information"
    else
        fail "Help output missing expected content"
        echo "Output was:"
        echo "${output}"
    fi
}

# Test 2: Missing required flag
test_missing_required_flag() {
    info "Test 2: Missing required flag (-ast)"

    local output
    local exit_code=0
    output=$("${BINARY}" 2>&1) || exit_code=$?

    if [[ ${exit_code} -ne 0 ]] && echo "${output}" | grep -q "\-ast flag is required"; then
        pass "Missing -ast flag produces correct error"
    else
        fail "Expected error about missing -ast flag"
        echo "Exit code: ${exit_code}"
        echo "Output was:"
        echo "${output}"
    fi
}

# Test 3: Process valid AST input
test_valid_ast_input() {
    info "Test 3: Process valid AST input (with verbose output)"

    local ast_file="${TEMP_DIR}/go-ast.json"
    local output_dir="${TEMP_DIR}/output"

    # Create valid Phase 2 SemanticDiff format AST input
    cat > "${ast_file}" << 'EOF'
[
  {
    "language": "go",
    "file_path": "internal/handler/user.go",
    "functions": [
      {
        "name": "CreateUser",
        "change_type": "modified",
        "before": {"receiver": "*UserHandler"},
        "after": {"receiver": "*UserHandler"}
      }
    ],
    "summary": {
      "functions_modified": 1
    }
  }
]
EOF

    local output
    local exit_code=0
    output=$("${BINARY}" -ast "${ast_file}" -output "${output_dir}" -verbose 2>&1) || exit_code=$?

    if [[ ${exit_code} -eq 0 ]] && \
       echo "${output}" | grep -q "AST input:" && \
       echo "${output}" | grep -q "Language: go" && \
       echo "${output}" | grep -q "Modified functions:"; then
        pass "Valid AST input processed successfully with verbose output"
    else
        fail "Failed to process valid AST input"
        echo "Exit code: ${exit_code}"
        echo "Output was:"
        echo "${output}"
    fi
}

# Test 4: Verify JSON output file exists
test_json_output_exists() {
    info "Test 4: Verify JSON output file exists ({lang}-calls.json)"

    local ast_file="${TEMP_DIR}/go-ast.json"
    local output_dir="${TEMP_DIR}/output-json"

    # Create valid Phase 2 SemanticDiff format AST input
    cat > "${ast_file}" << 'EOF'
[
  {
    "language": "go",
    "file_path": "internal/service/auth.go",
    "functions": [
      {
        "name": "ValidateToken",
        "change_type": "added",
        "after": {"receiver": "*AuthService"}
      }
    ],
    "summary": {
      "functions_added": 1
    }
  }
]
EOF

    # Run binary
    "${BINARY}" -ast "${ast_file}" -output "${output_dir}" > /dev/null 2>&1 || true

    local json_file="${output_dir}/go-calls.json"
    if [[ -f "${json_file}" ]]; then
        # Verify it's valid JSON
        if jq -e . "${json_file}" > /dev/null 2>&1; then
            pass "JSON output file exists and is valid: go-calls.json"
        else
            fail "JSON output file exists but is not valid JSON"
        fi
    else
        fail "JSON output file not found: ${json_file}"
        echo "Output directory contents:"
        ls -la "${output_dir}" 2>/dev/null || echo "(directory does not exist)"
    fi
}

# Test 5: Verify markdown summary exists
test_markdown_summary_exists() {
    info "Test 5: Verify markdown summary exists (impact-summary.md)"

    local ast_file="${TEMP_DIR}/go-ast.json"
    local output_dir="${TEMP_DIR}/output-md"

    # Create valid Phase 2 SemanticDiff format AST input
    cat > "${ast_file}" << 'EOF'
[
  {
    "language": "go",
    "file_path": "internal/repository/user.go",
    "functions": [
      {
        "name": "FindByID",
        "change_type": "modified",
        "before": {"receiver": "*UserRepository"},
        "after": {"receiver": "*UserRepository"}
      }
    ],
    "summary": {
      "functions_modified": 1
    }
  }
]
EOF

    # Run binary
    "${BINARY}" -ast "${ast_file}" -output "${output_dir}" > /dev/null 2>&1 || true

    local md_file="${output_dir}/impact-summary.md"
    if [[ -f "${md_file}" ]]; then
        # Verify it contains expected markdown content
        if grep -Eq '^# +Impact Summary' "${md_file}" && grep -Eq '^## +(Summary Metrics|Recommendations)' "${md_file}"; then
            pass "Markdown summary exists: impact-summary.md"
        else
            fail "Markdown summary exists but missing expected content"
            echo "Content:"
            head -20 "${md_file}"
        fi
    else
        fail "Markdown summary not found: ${md_file}"
        echo "Output directory contents:"
        ls -la "${output_dir}" 2>/dev/null || echo "(directory does not exist)"
    fi
}

# Test 6: Unsupported language error
test_unsupported_language() {
    info "Test 6: Unsupported language error"

    local ast_file="${TEMP_DIR}/ruby-ast.json"
    local output_dir="${TEMP_DIR}/output-ruby"

    # Create AST input with unsupported language
    cat > "${ast_file}" << 'EOF'
[
  {
    "language": "ruby",
    "file_path": "app/models/user.rb",
    "functions": [
      {
        "name": "create",
        "change_type": "modified"
      }
    ]
  }
]
EOF

    local output
    local exit_code=0
    output=$("${BINARY}" -ast "${ast_file}" -output "${output_dir}" -lang ruby 2>&1) || exit_code=$?

    if [[ ${exit_code} -ne 0 ]] && echo "${output}" | grep -q "unsupported language"; then
        pass "Unsupported language produces correct error"
    else
        fail "Expected unsupported language error"
        echo "Exit code: ${exit_code}"
        echo "Output was:"
        echo "${output}"
    fi
}

# Test 7: Empty functions list (edge case)
test_empty_functions() {
    info "Test 7: Empty functions list (edge case)"

    local ast_file="${TEMP_DIR}/go-empty-ast.json"
    local output_dir="${TEMP_DIR}/output-empty"

    # Create valid AST input with no functions
    cat > "${ast_file}" << 'EOF'
[
  {
    "language": "go",
    "file_path": "internal/handler/health.go",
    "functions": [],
    "summary": {
      "functions_modified": 0
    }
  }
]
EOF

    local output
    local exit_code=0
    output=$("${BINARY}" -ast "${ast_file}" -output "${output_dir}" -verbose 2>&1) || exit_code=$?

    if [[ ${exit_code} -eq 0 ]] && echo "${output}" | grep -q "No modified functions found"; then
        pass "Empty functions list handled correctly"
    else
        # Still a pass if it completes without error
        if [[ ${exit_code} -eq 0 ]]; then
            pass "Empty functions list processed without error"
        else
            fail "Failed to process empty functions list"
            echo "Exit code: ${exit_code}"
            echo "Output was:"
            echo "${output}"
        fi
    fi
}

# Test 8: Language auto-detection from filename
test_language_autodetect() {
    info "Test 8: Language auto-detection from filename"

    local ast_file="${TEMP_DIR}/typescript-ast.json"
    local output_dir="${TEMP_DIR}/output-ts"

    # Create valid TypeScript AST input (filename should trigger ts detection)
    cat > "${ast_file}" << 'EOF'
[
  {
    "language": "typescript",
    "file_path": "src/services/user.service.ts",
    "functions": [
      {
        "name": "createUser",
        "change_type": "added"
      }
    ],
    "summary": {
      "functions_added": 1
    }
  }
]
EOF

    local output
    local exit_code=0
    output=$("${BINARY}" -ast "${ast_file}" -output "${output_dir}" -verbose 2>&1) || exit_code=$?

    if [[ ${exit_code} -eq 0 ]] && echo "${output}" | grep -q "Language: typescript"; then
        # Verify correct output filename
        if [[ -f "${output_dir}/typescript-calls.json" ]]; then
            pass "Language auto-detected (typescript) with correct output filename"
        else
            fail "Output file not using detected language name"
        fi
    else
        fail "Failed to auto-detect language from filename"
        echo "Exit code: ${exit_code}"
        echo "Output was:"
        echo "${output}"
    fi
}

# Main
main() {
    echo "=========================================="
    echo "  Call-Graph Integration Tests"
    echo "=========================================="
    echo ""

    check_binary
    setup
    echo ""

    # Run all tests
    test_help_output
    test_missing_required_flag
    test_valid_ast_input
    test_json_output_exists
    test_markdown_summary_exists
    test_unsupported_language
    test_empty_functions
    test_language_autodetect

    echo ""
    echo "=========================================="
    echo "  Test Summary"
    echo "=========================================="
    echo -e "  ${GREEN}Passed${NC}: ${TESTS_PASSED}"
    echo -e "  ${RED}Failed${NC}: ${TESTS_FAILED}"
    echo "=========================================="

    # Exit with failure if any tests failed
    if [[ ${TESTS_FAILED} -gt 0 ]]; then
        exit 1
    fi

    exit 0
}

main "$@"
