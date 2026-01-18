#!/usr/bin/env python3
"""
Python AST Extractor for Semantic Diffs.

Extracts functions, classes, and imports from Python files
and compares before/after versions to generate semantic diffs.
"""

import ast
import hashlib
import json
import sys
from dataclasses import dataclass, field, asdict
from pathlib import Path
from typing import Optional


@dataclass
class Param:
    name: str
    type: str = ""


@dataclass
class FuncSig:
    params: list[Param]
    returns: list[str]
    is_async: bool = False
    decorators: list[str] = field(default_factory=list)
    is_exported: bool = True
    start_line: int = 0
    end_line: int = 0


@dataclass
class FunctionDiff:
    name: str
    change_type: str  # added, removed, modified, renamed
    before: Optional[FuncSig] = None
    after: Optional[FuncSig] = None
    body_diff: str = ""


@dataclass
class FieldDiff:
    name: str
    change_type: str
    old_type: str = ""
    new_type: str = ""


@dataclass
class TypeDiff:
    name: str
    kind: str  # class, dataclass
    change_type: str
    fields: list[FieldDiff] = field(default_factory=list)
    start_line: int = 0
    end_line: int = 0


@dataclass
class ImportDiff:
    path: str
    alias: str = ""
    change_type: str = ""


@dataclass
class ChangeSummary:
    functions_added: int = 0
    functions_removed: int = 0
    functions_modified: int = 0
    types_added: int = 0
    types_removed: int = 0
    types_modified: int = 0
    variables_added: int = 0
    variables_removed: int = 0
    variables_modified: int = 0
    imports_added: int = 0
    imports_removed: int = 0


@dataclass
class SemanticDiff:
    language: str
    file_path: str
    functions: list[FunctionDiff]
    types: list[TypeDiff]
    imports: list[ImportDiff]
    summary: ChangeSummary
    error: str = ""


@dataclass
class ParsedFunc:
    name: str
    params: list[Param]
    returns: list[str]
    is_async: bool
    decorators: list[str]
    is_exported: bool
    start_line: int
    end_line: int
    body_hash: str


@dataclass
class ParsedClass:
    name: str
    is_dataclass: bool
    fields: dict[str, str]  # name -> type
    methods: list[str]
    is_exported: bool
    start_line: int
    end_line: int


@dataclass
class ParsedFile:
    functions: dict[str, ParsedFunc]
    classes: dict[str, ParsedClass]
    imports: dict[str, str]  # module -> alias
    error: str = ""


def get_annotation_str(node: Optional[ast.expr]) -> str:
    """Convert an annotation AST node to string."""
    if node is None:
        return ""
    return ast.unparse(node)


def parse_file(file_path: str) -> ParsedFile:
    """Parse a Python file and extract semantic information."""
    result = ParsedFile(functions={}, classes={}, imports={})

    if not file_path or not Path(file_path).exists():
        return result

    content = Path(file_path).read_text()
    try:
        tree = ast.parse(content)
    except SyntaxError as e:
        result.error = f"Syntax error: {e.msg} at line {e.lineno}"
        return result

    for node in ast.walk(tree):
        # Extract imports
        if isinstance(node, ast.Import):
            for alias in node.names:
                result.imports[alias.name] = alias.asname or ""

        elif isinstance(node, ast.ImportFrom):
            module = node.module or ""
            for alias in node.names:
                key = f"{module}.{alias.name}" if module else alias.name
                result.imports[key] = alias.asname or ""

    # Process top-level definitions
    for node in ast.iter_child_nodes(tree):
        if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef)):
            func = _parse_function(node, content)
            result.functions[func.name] = func

        elif isinstance(node, ast.ClassDef):
            cls = _parse_class(node, content)
            result.classes[cls.name] = cls

            # Extract methods as functions
            for item in node.body:
                if isinstance(item, (ast.FunctionDef, ast.AsyncFunctionDef)):
                    method = _parse_function(item, content)
                    method.name = f"{cls.name}.{method.name}"
                    result.functions[method.name] = method

    return result


def _parse_function(
    node: ast.FunctionDef | ast.AsyncFunctionDef, content: str
) -> ParsedFunc:
    """Parse a function definition."""
    params = []
    for arg in node.args.args:
        params.append(Param(name=arg.arg, type=get_annotation_str(arg.annotation)))

    returns = []
    if node.returns:
        returns.append(get_annotation_str(node.returns))

    decorators = []
    for dec in node.decorator_list:
        if isinstance(dec, ast.Name):
            decorators.append(dec.id)
        elif isinstance(dec, ast.Call) and isinstance(dec.func, ast.Name):
            decorators.append(dec.func.id)
        elif isinstance(dec, ast.Attribute):
            decorators.append(ast.unparse(dec))

    # Hash the body for change detection
    end_line = node.end_lineno if node.end_lineno else node.lineno
    body_lines = content.split("\n")[node.lineno - 1 : end_line]
    body_hash = hashlib.sha256("\n".join(body_lines).encode("utf-8")).hexdigest()

    return ParsedFunc(
        name=node.name,
        params=params,
        returns=returns,
        is_async=isinstance(node, ast.AsyncFunctionDef),
        decorators=decorators,
        is_exported=not node.name.startswith("_"),
        start_line=node.lineno,
        end_line=node.end_lineno or node.lineno,
        body_hash=body_hash,
    )


def _parse_class(node: ast.ClassDef, content: str) -> ParsedClass:
    """Parse a class definition."""
    is_dataclass = any(
        (isinstance(d, ast.Name) and d.id == "dataclass")
        or (
            isinstance(d, ast.Call)
            and isinstance(d.func, ast.Name)
            and d.func.id == "dataclass"
        )
        for d in node.decorator_list
    )

    fields: dict[str, str] = {}
    methods: list[str] = []

    for item in node.body:
        if isinstance(item, ast.AnnAssign) and isinstance(item.target, ast.Name):
            fields[item.target.id] = get_annotation_str(item.annotation)
        elif isinstance(item, (ast.FunctionDef, ast.AsyncFunctionDef)):
            methods.append(item.name)

    return ParsedClass(
        name=node.name,
        is_dataclass=is_dataclass,
        fields=fields,
        methods=methods,
        is_exported=not node.name.startswith("_"),
        start_line=node.lineno,
        end_line=node.end_lineno or node.lineno,
    )


def compare_functions(
    before: dict[str, ParsedFunc], after: dict[str, ParsedFunc]
) -> list[FunctionDiff]:
    """Compare functions between before and after versions."""
    diffs = []

    # Find removed and modified
    for name, before_func in before.items():
        after_func = after.get(name)
        if after_func is None:
            diffs.append(
                FunctionDiff(
                    name=name,
                    change_type="removed",
                    before=FuncSig(
                        params=before_func.params,
                        returns=before_func.returns,
                        is_async=before_func.is_async,
                        decorators=before_func.decorators,
                        is_exported=before_func.is_exported,
                        start_line=before_func.start_line,
                        end_line=before_func.end_line,
                    ),
                )
            )
            continue

        # Check for modifications
        changes = []
        if before_func.params != after_func.params:
            changes.append("parameters changed")
        if before_func.returns != after_func.returns:
            changes.append("return type changed")
        if before_func.is_async != after_func.is_async:
            changes.append("async modifier changed")
        if before_func.decorators != after_func.decorators:
            changes.append("decorators changed")
        if before_func.body_hash != after_func.body_hash:
            changes.append("implementation changed")

        if changes:
            diffs.append(
                FunctionDiff(
                    name=name,
                    change_type="modified",
                    before=FuncSig(
                        params=before_func.params,
                        returns=before_func.returns,
                        is_async=before_func.is_async,
                        decorators=before_func.decorators,
                        is_exported=before_func.is_exported,
                        start_line=before_func.start_line,
                        end_line=before_func.end_line,
                    ),
                    after=FuncSig(
                        params=after_func.params,
                        returns=after_func.returns,
                        is_async=after_func.is_async,
                        decorators=after_func.decorators,
                        is_exported=after_func.is_exported,
                        start_line=after_func.start_line,
                        end_line=after_func.end_line,
                    ),
                    body_diff=", ".join(changes),
                )
            )

    # Find added
    for name, after_func in after.items():
        if name not in before:
            diffs.append(
                FunctionDiff(
                    name=name,
                    change_type="added",
                    after=FuncSig(
                        params=after_func.params,
                        returns=after_func.returns,
                        is_async=after_func.is_async,
                        decorators=after_func.decorators,
                        is_exported=after_func.is_exported,
                        start_line=after_func.start_line,
                        end_line=after_func.end_line,
                    ),
                )
            )

    return diffs


def compare_classes(
    before: dict[str, ParsedClass], after: dict[str, ParsedClass]
) -> list[TypeDiff]:
    """Compare classes between before and after versions."""
    diffs = []

    for name, before_cls in before.items():
        after_cls = after.get(name)
        if after_cls is None:
            diffs.append(
                TypeDiff(
                    name=name,
                    kind="dataclass" if before_cls.is_dataclass else "class",
                    change_type="removed",
                    start_line=before_cls.start_line,
                    end_line=before_cls.end_line,
                )
            )
            continue

        # Compare fields
        field_diffs = []
        for field_name, field_type in before_cls.fields.items():
            after_type = after_cls.fields.get(field_name)
            if after_type is None:
                field_diffs.append(
                    FieldDiff(
                        name=field_name,
                        change_type="removed",
                        old_type=field_type,
                    )
                )
            elif after_type != field_type:
                field_diffs.append(
                    FieldDiff(
                        name=field_name,
                        change_type="modified",
                        old_type=field_type,
                        new_type=after_type,
                    )
                )

        for field_name, field_type in after_cls.fields.items():
            if field_name not in before_cls.fields:
                field_diffs.append(
                    FieldDiff(
                        name=field_name,
                        change_type="added",
                        new_type=field_type,
                    )
                )

        if field_diffs or before_cls.is_dataclass != after_cls.is_dataclass:
            diffs.append(
                TypeDiff(
                    name=name,
                    kind="dataclass" if after_cls.is_dataclass else "class",
                    change_type="modified",
                    fields=field_diffs,
                    start_line=after_cls.start_line,
                    end_line=after_cls.end_line,
                )
            )

    for name, after_cls in after.items():
        if name not in before:
            diffs.append(
                TypeDiff(
                    name=name,
                    kind="dataclass" if after_cls.is_dataclass else "class",
                    change_type="added",
                    start_line=after_cls.start_line,
                    end_line=after_cls.end_line,
                )
            )

    return diffs


def compare_imports(before: dict[str, str], after: dict[str, str]) -> list[ImportDiff]:
    """Compare imports between before and after versions."""
    diffs = []

    for path, alias in before.items():
        if path not in after:
            diffs.append(ImportDiff(path=path, alias=alias, change_type="removed"))

    for path, alias in after.items():
        if path not in before:
            diffs.append(ImportDiff(path=path, alias=alias, change_type="added"))

    return diffs


def extract_diff(before_path: str, after_path: str) -> SemanticDiff:
    """Extract semantic diff between two Python files."""
    before = parse_file(before_path)
    after = parse_file(after_path)

    functions = compare_functions(before.functions, after.functions)
    types = compare_classes(before.classes, after.classes)
    imports = compare_imports(before.imports, after.imports)

    summary = ChangeSummary(
        functions_added=sum(1 for f in functions if f.change_type == "added"),
        functions_removed=sum(1 for f in functions if f.change_type == "removed"),
        functions_modified=sum(1 for f in functions if f.change_type == "modified"),
        types_added=sum(1 for t in types if t.change_type == "added"),
        types_removed=sum(1 for t in types if t.change_type == "removed"),
        types_modified=sum(1 for t in types if t.change_type == "modified"),
        imports_added=sum(1 for i in imports if i.change_type == "added"),
        imports_removed=sum(1 for i in imports if i.change_type == "removed"),
    )

    return SemanticDiff(
        language="python",
        file_path=after_path or before_path,
        functions=functions,
        types=types,
        imports=imports,
        summary=summary,
    )


def dataclass_to_dict(obj):
    """Recursively convert dataclass to dict, omitting empty values."""
    if hasattr(obj, "__dataclass_fields__"):
        result = {}
        for key, value in asdict(obj).items():
            if value is None:
                continue
            if isinstance(value, list) and not value:
                continue
            if isinstance(value, str) and not value:
                continue
            result[key] = value
        return result
    elif isinstance(obj, list):
        return [dataclass_to_dict(item) for item in obj]
    elif isinstance(obj, dict):
        return {k: dataclass_to_dict(v) for k, v in obj.items()}
    return obj


def main():
    """CLI entry point."""
    # Parse command line arguments
    before_path = ""
    after_path = ""

    i = 1
    while i < len(sys.argv):
        arg = sys.argv[i]
        if arg == "--before":
            if i + 1 < len(sys.argv):
                before_path = sys.argv[i + 1]
                i += 2
            else:
                print("Error: --before requires a path argument", file=sys.stderr)
                sys.exit(1)
        elif arg == "--after":
            if i + 1 < len(sys.argv):
                after_path = sys.argv[i + 1]
                i += 2
            else:
                print("Error: --after requires a path argument", file=sys.stderr)
                sys.exit(1)
        elif arg in ("--help", "-h"):
            print("Usage: ast_extractor.py --before <path> --after <path>")
            print()
            print("Options:")
            print("  --before <path>  Path to the before version of the file")
            print("  --after <path>   Path to the after version of the file")
            print()
            print('Use empty string "" for new/deleted files')
            sys.exit(0)
        else:
            # Positional arguments for backward compatibility
            if not before_path:
                before_path = arg
            elif not after_path:
                after_path = arg
            i += 1

    # Handle empty string markers
    if before_path in ('""', "''"):
        before_path = ""
    if after_path in ('""', "''"):
        after_path = ""

    if not before_path and not after_path:
        print("Usage: ast_extractor.py --before <path> --after <path>", file=sys.stderr)
        print('Use empty string "" for new/deleted files', file=sys.stderr)
        sys.exit(1)

    try:
        diff = extract_diff(before_path, after_path)
        output = dataclass_to_dict(diff)
        print(json.dumps(output, indent=2))
    except Exception as e:
        error_diff = SemanticDiff(
            language="python",
            file_path=after_path or before_path,
            functions=[],
            types=[],
            imports=[],
            summary=ChangeSummary(),
            error=str(e),
        )
        print(json.dumps(dataclass_to_dict(error_diff), indent=2))
        sys.exit(1)


if __name__ == "__main__":
    main()
