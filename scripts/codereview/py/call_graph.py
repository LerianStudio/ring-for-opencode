#!/usr/bin/env python3
"""
Python call graph analyzer using AST.

Analyzes Python files and builds function-level call relationships.
Outputs JSON for consumption by the Go wrapper.

Usage:
    python3 call_graph.py <files...> [--functions func1,func2,...]
"""

import ast
import json
import os
import sys
from dataclasses import dataclass, field
from collections.abc import Iterator
from typing import Optional

# Maximum file size to analyze (10MB) - prevents memory exhaustion
MAX_FILE_SIZE = 10 * 1024 * 1024


@dataclass
class CallSite:
    """Represents a call site within a function."""

    target: str
    line: int
    column: int
    is_method: bool


@dataclass
class CallerInfo:
    """Represents information about a function caller."""

    function: str
    file: str
    line: int


@dataclass
class FunctionInfo:
    """Comprehensive function information including call relationships."""

    name: str
    file: str
    line: int
    end_line: int
    call_sites: list[CallSite] = field(default_factory=list)
    called_by: list[CallerInfo] = field(default_factory=list)


@dataclass
class CallGraphOutput:
    """Output format for the call graph analysis."""

    functions: list[FunctionInfo] = field(default_factory=list)
    error: Optional[str] = None


class CallGraphVisitor(ast.NodeVisitor):
    """AST visitor that extracts function definitions and their call sites."""

    def __init__(self, file_path: str):
        self.file_path = file_path
        self.functions: list[FunctionInfo] = []
        self._class_stack: list[str] = []
        self._current_function: Optional[FunctionInfo] = None

    def visit_ClassDef(self, node: ast.ClassDef) -> None:
        """Visit a class definition and track class context for methods."""
        self._class_stack.append(node.name)
        self.generic_visit(node)
        self._class_stack.pop()

    def visit_FunctionDef(self, node: ast.FunctionDef) -> None:
        """Visit a function definition."""
        self._process_function(node)

    def visit_AsyncFunctionDef(self, node: ast.AsyncFunctionDef) -> None:
        """Visit an async function definition."""
        self._process_function(node)

    def _process_function(
        self, node: ast.FunctionDef | ast.AsyncFunctionDef
    ) -> None:
        """Process a function or async function definition."""
        # Build qualified name with class prefix if within a class
        if self._class_stack:
            name = f"{self._class_stack[-1]}.{node.name}"
        else:
            name = node.name

        func_info = FunctionInfo(
            name=name,
            file=self.file_path,
            line=node.lineno,
            end_line=node.end_lineno or node.lineno,
            call_sites=[],
            called_by=[],
        )

        # Store previous function context and set current
        prev_function = self._current_function
        self._current_function = func_info

        # Extract call sites from the function body
        # Use walk_without_nested_functions to prevent double-counting
        # calls that appear in nested functions
        call_extractor = CallSiteExtractor()
        for child in walk_without_nested_functions(node):
            if isinstance(child, ast.Call):
                call_site = call_extractor.extract_call(child)
                if call_site:
                    func_info.call_sites.append(call_site)

        self.functions.append(func_info)

        # Visit nested functions
        self.generic_visit(node)

        # Restore previous function context
        self._current_function = prev_function


def walk_without_nested_functions(node: ast.AST) -> Iterator[ast.AST]:
    """
    Walk AST nodes without descending into nested function definitions.

    This prevents double-counting calls that appear in nested functions.
    """
    from collections import deque
    todo = deque([node])
    while todo:
        current = todo.popleft()
        # Skip the root node itself if it's a function (we want its children)
        if current is not node and isinstance(
            current, (ast.FunctionDef, ast.AsyncFunctionDef)
        ):
            # Don't descend into nested functions
            continue
        yield current
        todo.extend(ast.iter_child_nodes(current))


class CallSiteExtractor:
    """Extracts call information from AST Call nodes."""

    def extract_call(self, node: ast.Call) -> Optional[CallSite]:
        """Extract call information from a Call node."""
        target = self._get_call_target(node.func)
        if not target:
            return None

        # Skip common built-ins that aren't useful for call graph
        skip_targets = {
            "print",
            "len",
            "str",
            "int",
            "float",
            "list",
            "dict",
            "set",
            "tuple",
            "range",
            "enumerate",
            "zip",
            "map",
            "filter",
            "sorted",
            "reversed",
            "isinstance",
            "issubclass",
            "hasattr",
            "getattr",
            "setattr",
            "delattr",
            "type",
            "super",
            "open",
            "input",
        }

        base_target = target.split(".")[-1]
        if base_target in skip_targets:
            return None

        is_method = isinstance(node.func, ast.Attribute)

        return CallSite(
            target=target,
            line=node.lineno,
            column=node.col_offset + 1,
            is_method=is_method,
        )

    def _get_call_target(self, func: ast.expr) -> Optional[str]:
        """Get the target name from a function call expression."""
        if isinstance(func, ast.Name):
            # Direct function call: function_name()
            return func.id
        elif isinstance(func, ast.Attribute):
            # Method call: obj.method() or module.function()
            value = self._get_expr_name(func.value)
            if value:
                return f"{value}.{func.attr}"
            return func.attr
        elif isinstance(func, ast.Subscript):
            # Subscript call: obj[key]()
            return "<subscript>"
        elif isinstance(func, ast.Call):
            # Chained call: func()()
            return "<chained>"
        return None

    def _get_expr_name(self, expr: ast.expr) -> Optional[str]:
        """Get a name representation for an expression."""
        if isinstance(expr, ast.Name):
            return expr.id
        elif isinstance(expr, ast.Attribute):
            value = self._get_expr_name(expr.value)
            if value:
                return f"{value}.{expr.attr}"
            return expr.attr
        elif expr.__class__.__name__ == "Constant":
            # ast.Constant for literals
            return None
        return None


def is_test_function(name: str) -> bool:
    """Check if a function name indicates a test function."""
    # Remove class prefix if present
    base_name = name.split(".")[-1]

    # Python unittest/pytest patterns
    if base_name.startswith("test_"):
        return True
    if base_name.startswith("Test"):
        return True
    if base_name.endswith("_test"):
        return True

    return False


def is_test_file(file_path: str) -> bool:
    """Check if a file path indicates a test file."""
    base = os.path.basename(file_path)

    # Common test file patterns
    if base.startswith("test_"):
        return True
    if base.endswith("_test.py"):
        return True
    if "tests/" in file_path or "/tests/" in file_path:
        return True
    if "test/" in file_path or "/test/" in file_path:
        return True

    return False


def analyze_file(file_path: str) -> list[FunctionInfo]:
    """Analyze a single Python file and extract function information."""
    # Check file size before reading to prevent memory exhaustion
    try:
        file_size = os.path.getsize(file_path)
        if file_size > MAX_FILE_SIZE:
            return []  # Skip files that are too large
    except OSError:
        return []

    try:
        with open(file_path, encoding="utf-8") as f:
            source = f.read()
    except OSError:
        return []

    try:
        tree = ast.parse(source, filename=file_path)
    except SyntaxError:
        return []

    visitor = CallGraphVisitor(file_path)
    visitor.visit(tree)

    return visitor.functions


def build_caller_map(
    all_functions: list[FunctionInfo],
) -> dict[str, list[CallerInfo]]:
    """Build a map of function names to their callers."""
    caller_map: dict[str, list[CallerInfo]] = {}

    # Initialize map for all functions
    for fn in all_functions:
        caller_map[fn.name] = []

    # Build caller relationships
    for caller_fn in all_functions:
        for call_site in caller_fn.call_sites:
            target_name = call_site.target
            base_name = target_name.split(".")[-1]

            # Try to match with known functions
            for target_fn in all_functions:
                target_base = target_fn.name.split(".")[-1]
                if (
                    target_fn.name == target_name
                    or target_fn.name == base_name
                    or target_base == base_name
                ):
                    if target_fn.name in caller_map:
                        caller_map[target_fn.name].append(
                            CallerInfo(
                                function=caller_fn.name,
                                file=caller_fn.file,
                                line=call_site.line,
                            )
                        )
                    break

    return caller_map


def analyze_call_graph(
    file_paths: list[str], target_functions: Optional[list[str]] = None
) -> CallGraphOutput:
    """
    Analyze Python files and build function-level call relationships.

    Args:
        file_paths: List of file paths to analyze
        target_functions: Optional list of function names to focus on

    Returns:
        CallGraphOutput with analysis results
    """
    result = CallGraphOutput()

    if not file_paths:
        return result

    # Filter out non-existent files
    existing_files = [fp for fp in file_paths if os.path.isfile(fp)]

    if not existing_files:
        result.error = "No valid files to analyze"
        return result

    # Collect all functions from all files
    all_functions: list[FunctionInfo] = []
    for file_path in existing_files:
        functions = analyze_file(file_path)
        all_functions.extend(functions)

    # Build caller map
    caller_map = build_caller_map(all_functions)

    # Filter functions if target list provided
    functions_to_report = all_functions
    if target_functions:
        target_set = set(target_functions)
        functions_to_report = [
            fn
            for fn in all_functions
            if fn.name in target_set or fn.name.split(".")[-1] in target_set
        ]

    # Build output with caller information
    for fn in functions_to_report:
        callers = caller_map.get(fn.name, [])
        fn.called_by = callers
        result.functions.append(fn)

    return result


def to_dict(obj) -> dict:
    """Convert a dataclass object to a dictionary for JSON serialization."""
    if hasattr(obj, "__dataclass_fields__"):
        result = {}
        for field_name in obj.__dataclass_fields__:
            value = getattr(obj, field_name)
            if isinstance(value, list):
                result[field_name] = [to_dict(item) for item in value]
            elif value is None:
                continue  # Skip None values
            else:
                result[field_name] = value
        return result
    return obj


def output_to_json(output: CallGraphOutput) -> str:
    """Convert CallGraphOutput to JSON string."""
    data = {
        "functions": [to_dict(fn) for fn in output.functions],
    }
    if output.error:
        data["error"] = output.error
    return json.dumps(data, indent=2)


def parse_args(args: list[str]) -> tuple[list[str], list[str]]:
    """Parse CLI arguments."""
    files: list[str] = []
    functions: list[str] = []

    i = 0
    while i < len(args):
        arg = args[i]
        if arg == "--functions" and i + 1 < len(args):
            # Parse comma-separated function names
            func_list = args[i + 1]
            functions.extend(
                f.strip() for f in func_list.split(",") if f.strip()
            )
            i += 2
        elif not arg.startswith("--"):
            files.append(arg)
            i += 1
        else:
            i += 1

    return files, functions


def main() -> None:
    """Main CLI entry point."""
    args = sys.argv[1:]

    if not args:
        print(
            "Usage: call_graph.py <files...> [--functions func1,func2,...]",
            file=sys.stderr,
        )
        print("", file=sys.stderr)
        print(
            "Analyzes Python files and outputs function call relationships.",
            file=sys.stderr,
        )
        print("", file=sys.stderr)
        print("Options:", file=sys.stderr)
        print(
            "  --functions    Comma-separated list of function names to focus on",
            file=sys.stderr,
        )
        sys.exit(1)

    files, functions = parse_args(args)

    if not files:
        print("Error: No files specified", file=sys.stderr)
        sys.exit(1)

    # Resolve file paths
    resolved_files = [os.path.abspath(f) for f in files]

    try:
        target_functions = functions if functions else None
        result = analyze_call_graph(resolved_files, target_functions)
        print(output_to_json(result))
    except Exception as e:
        output = CallGraphOutput(error=f"{type(e).__name__}: {e}")
        print(output_to_json(output))
        sys.exit(1)


if __name__ == "__main__":
    main()
