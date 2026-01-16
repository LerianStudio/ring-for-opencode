#!/usr/bin/env python3
"""
Data flow vulnerability analyzer for Python and TypeScript.

Detects sources (user input, external data) and sinks (dangerous operations)
and tracks data flow between them to identify potential security vulnerabilities.

Usage:
    python3 data_flow.py <language> <file1> [file2] ...

Languages:
    python, typescript

Output:
    JSON to stdout with sources, sinks, flows, and nil_sources
"""

import hashlib
import json
import os
import re
import sys
from dataclasses import dataclass, field
from typing import Optional

# Maximum file size to analyze (10MB) - prevents memory exhaustion
MAX_FILE_SIZE = 10 * 1024 * 1024


@dataclass
class Source:
    """Represents a data source (user input, external data)."""

    source_type: str  # http_body, http_query, env_var, etc.
    variable: str
    file: str
    line: int
    column: int
    pattern: str  # The matched pattern


@dataclass
class Sink:
    """Represents a data sink (dangerous operation)."""

    sink_type: str  # database, command_exec, http_response, etc.
    variable: str
    file: str
    line: int
    column: int
    pattern: str  # The matched pattern
    context: str = ""  # The line content for sanitization checking


@dataclass
class Flow:
    """Represents a data flow from source to sink."""

    id: str  # Unique hash of source+sink
    source: Source
    sink: Sink
    risk: str  # critical, high, medium, low
    sanitized: bool
    sanitizer: Optional[str] = None


@dataclass
class NilSource:
    """Represents a potentially null/nil variable source."""

    variable: str
    file: str
    line: int
    column: int
    pattern: str
    reason: str  # database_query, map_lookup, json_parse, etc.


@dataclass
class AnalysisOutput:
    """Output format for the data flow analysis."""

    sources: list[Source] = field(default_factory=list)
    sinks: list[Sink] = field(default_factory=list)
    flows: list[Flow] = field(default_factory=list)
    nil_sources: list[NilSource] = field(default_factory=list)
    error: Optional[str] = None


# =============================================================================
# Python Source Patterns (Flask, Django, FastAPI)
# =============================================================================

PYTHON_SOURCE_PATTERNS: dict[str, list[tuple[str, str]]] = {
    "http_body": [
        (r"request\.get_json\s*\(\s*\)", "request.get_json()"),
        (r"request\.json\b", "request.json"),
        (r"request\.data\b", "request.data"),
        (r"request\.form\b", "request.form"),
        (r"request\.POST\b", "request.POST"),
        (r"await\s+request\.json\s*\(\s*\)", "await request.json()"),
        (r"await\s+request\.body\s*\(\s*\)", "await request.body()"),
    ],
    "http_query": [
        (r"request\.args\.get\s*\(", "request.args.get"),
        (r"request\.GET\.get\s*\(", "request.GET.get"),
        (r"request\.GET\[", "request.GET[]"),
        (r"request\.query_params\.get\s*\(", "request.query_params.get"),
        (r"Query\s*\(", "Query()"),
    ],
    "http_header": [
        (r"request\.headers\.get\s*\(", "request.headers.get"),
        (r"request\.headers\[", "request.headers[]"),
        (r"request\.META\.get\s*\(", "request.META.get"),
        (r"Header\s*\(", "Header()"),
    ],
    "http_path": [
        (r"@app\.route\s*\([^)]*<([^>]+)>", "route_param"),
        (r"@router\.(get|post|put|delete|patch)\s*\([^)]*\{([^}]+)\}", "path_param"),
        (r"Path\s*\(", "Path()"),
    ],
    "env_var": [
        (r"os\.getenv\s*\(", "os.getenv"),
        (r"os\.environ\.get\s*\(", "os.environ.get"),
        (r"os\.environ\[", "os.environ[]"),
        (r"environ\.get\s*\(", "environ.get"),
    ],
    "file_read": [
        (r"open\s*\([^)]*\)\.read\s*\(", "open().read()"),
        (r"\.read_text\s*\(\s*\)", ".read_text()"),
        (r"\.read_bytes\s*\(\s*\)", ".read_bytes()"),
        (r"Path\s*\([^)]*\)\.read", "Path().read"),
        (r"with\s+open\s*\(", "with open()"),
    ],
    "database": [
        (r"\.execute\s*\(", ".execute()"),
        (r"\.fetchone\s*\(\s*\)", ".fetchone()"),
        (r"\.fetchall\s*\(\s*\)", ".fetchall()"),
        (r"\.fetchmany\s*\(", ".fetchmany()"),
        (r"cursor\.", "cursor operation"),
        (r"\.objects\.raw\s*\(", ".objects.raw()"),
        (r"\.objects\.get\s*\(", ".objects.get()"),
        (r"\.objects\.filter\s*\(", ".objects.filter()"),
    ],
    "external_api": [
        (r"requests\.get\s*\(", "requests.get"),
        (r"requests\.post\s*\(", "requests.post"),
        (r"requests\.put\s*\(", "requests.put"),
        (r"requests\.delete\s*\(", "requests.delete"),
        (r"requests\.patch\s*\(", "requests.patch"),
        (r"httpx\.(get|post|put|delete|patch)\s*\(", "httpx request"),
        (r"aiohttp\.ClientSession\s*\(", "aiohttp session"),
        (r"await\s+session\.(get|post|put|delete|patch)\s*\(", "aiohttp request"),
        (r"urllib\.request\.urlopen\s*\(", "urllib.request.urlopen"),
    ],
}


# =============================================================================
# Python Sink Patterns
# =============================================================================

PYTHON_SINK_PATTERNS: dict[str, list[tuple[str, str]]] = {
    "database": [
        (r"cursor\.execute\s*\([^)]*%", "cursor.execute with % formatting"),
        (r"cursor\.execute\s*\([^)]*\.format\s*\(", "cursor.execute with .format()"),
        (r'cursor\.execute\s*\(\s*f["\']', "cursor.execute with f-string"),
        (r"\.raw\s*\(", ".raw() query"),
        (r"\.extra\s*\(", ".extra() query"),
        (r"connection\.execute\s*\(", "connection.execute"),
        (r"RawSQL\s*\(", "RawSQL()"),
    ],
    "command_exec": [
        (r"subprocess\.run\s*\(", "subprocess.run"),
        (r"subprocess\.call\s*\(", "subprocess.call"),
        (r"subprocess\.Popen\s*\(", "subprocess.Popen"),
        (r"os\.system\s*\(", "os.system"),
        (r"os\.popen\s*\(", "os.popen"),
        (r"os\.exec[lv]p?e?\s*\(", "os.exec*"),
        (r"\beval\s*\(", "eval()"),
        (r"\bexec\s*\(", "exec()"),
        (r"compile\s*\(", "compile()"),
        (r"__import__\s*\(", "__import__()"),
    ],
    "http_response": [
        (r"jsonify\s*\(", "jsonify()"),
        (r"make_response\s*\(", "make_response()"),
        (r"render_template\s*\(", "render_template()"),
        (r"HttpResponse\s*\(", "HttpResponse()"),
        (r"JsonResponse\s*\(", "JsonResponse()"),
        (r"Response\s*\(", "Response()"),
        (r"return\s+\{", "return dict"),
    ],
    "logging": [
        (r"logging\.(debug|info|warning|error|critical)\s*\(", "logging.*()"),
        (r"logger\.(debug|info|warning|error|critical)\s*\(", "logger.*()"),
        (r"\bprint\s*\(", "print()"),
    ],
    "file_write": [
        (r'open\s*\([^)]*["\'][wa][+]?["\']', "open() write mode"),
        (r"\.write\s*\(", ".write()"),
        (r"\.write_text\s*\(", ".write_text()"),
        (r"\.write_bytes\s*\(", ".write_bytes()"),
        (r"shutil\.copy", "shutil.copy"),
        (r"shutil\.move", "shutil.move"),
    ],
    "template": [
        (r"render_template\s*\(", "render_template()"),
        (r"render_template_string\s*\(", "render_template_string()"),
        (r"Template\s*\(", "Template()"),
        (r"Environment\s*\(", "Jinja2 Environment()"),
        (r"Markup\s*\(", "Markup()"),
    ],
    "redirect": [
        (r"redirect\s*\(", "redirect()"),
        (r"HttpResponseRedirect\s*\(", "HttpResponseRedirect()"),
        (r"RedirectResponse\s*\(", "RedirectResponse()"),
    ],
}


# =============================================================================
# TypeScript Source Patterns (Express, Node, Koa)
# =============================================================================

TYPESCRIPT_SOURCE_PATTERNS: dict[str, list[tuple[str, str]]] = {
    "http_body": [
        (r"req\.body\b", "req.body"),
        (r"request\.body\b", "request.body"),
        (r"ctx\.request\.body\b", "ctx.request.body"),
        (r"ctx\.body\b", "ctx.body"),
        (r"event\.body\b", "event.body"),
        (r"\.json\s*\(\s*\)", ".json()"),
    ],
    "http_query": [
        (r"req\.query\b", "req.query"),
        (r"req\.params\b", "req.params"),
        (r"request\.query\b", "request.query"),
        (r"ctx\.query\b", "ctx.query"),
        (r"ctx\.params\b", "ctx.params"),
        (r"searchParams\.get\s*\(", "searchParams.get"),
        (r"URLSearchParams\s*\(", "URLSearchParams"),
        (r"url\.searchParams\b", "url.searchParams"),
    ],
    "http_header": [
        (r"req\.headers\b", "req.headers"),
        (r"request\.headers\b", "request.headers"),
        (r"ctx\.headers\b", "ctx.headers"),
        (r"ctx\.request\.headers\b", "ctx.request.headers"),
        (r"headers\.get\s*\(", "headers.get"),
    ],
    "env_var": [
        (r"process\.env\b", "process.env"),
        (r"Deno\.env\.get\s*\(", "Deno.env.get"),
        (r"Bun\.env\b", "Bun.env"),
    ],
    "file_read": [
        (r"fs\.readFile\s*\(", "fs.readFile"),
        (r"fs\.readFileSync\s*\(", "fs.readFileSync"),
        (r"readFile\s*\(", "readFile"),
        (r"readFileSync\s*\(", "readFileSync"),
        (r"Deno\.readTextFile\s*\(", "Deno.readTextFile"),
        (r"Bun\.file\s*\(", "Bun.file"),
    ],
    "database": [
        (r"\.query\s*\(", ".query()"),
        (r"\.findOne\s*\(", ".findOne()"),
        (r"\.find\s*\(", ".find()"),
        (r"\.findFirst\s*\(", ".findFirst()"),
        (r"\.findUnique\s*\(", ".findUnique()"),
        (r"\.findMany\s*\(", ".findMany()"),
        (r"\.aggregate\s*\(", ".aggregate()"),
        (r"\.select\s*\(", ".select()"),
        (r"prisma\.\w+\.", "prisma query"),
        (r"db\.\w+\.", "db query"),
    ],
    "external_api": [
        (r"\bfetch\s*\(", "fetch()"),
        (r"axios\.(get|post|put|delete|patch)\s*\(", "axios request"),
        (r"axios\s*\(", "axios()"),
        (r"got\s*\(", "got()"),
        (r"got\.(get|post|put|delete|patch)\s*\(", "got request"),
        (r"http\.request\s*\(", "http.request"),
        (r"https\.request\s*\(", "https.request"),
    ],
}


# =============================================================================
# TypeScript Sink Patterns
# =============================================================================

TYPESCRIPT_SINK_PATTERNS: dict[str, list[tuple[str, str]]] = {
    "database": [
        (r"\.query\s*\(\s*`", ".query() with template literal"),
        (r"\.query\s*\(\s*[\"'][^\"']*\$\{", ".query() with interpolation"),
        (r"\.exec\s*\(", ".exec()"),
        (r"\.raw\s*\(", ".raw()"),
        (r"\$queryRaw\s*`", "$queryRaw template"),
        (r"\$executeRaw\s*`", "$executeRaw template"),
        (r"\.rawQuery\s*\(", ".rawQuery()"),
    ],
    "command_exec": [
        (r"\bexec\s*\(", "exec()"),
        (r"\bexecSync\s*\(", "execSync()"),
        (r"\bspawn\s*\(", "spawn()"),
        (r"\bspawnSync\s*\(", "spawnSync()"),
        (r"\beval\s*\(", "eval()"),
        (r"new\s+Function\s*\(", "new Function()"),
        (r"vm\.runInContext\s*\(", "vm.runInContext"),
        (r"vm\.runInNewContext\s*\(", "vm.runInNewContext"),
        (r"vm\.Script\s*\(", "vm.Script"),
        (r"child_process\.", "child_process"),
    ],
    "http_response": [
        (r"res\.send\s*\(", "res.send()"),
        (r"res\.json\s*\(", "res.json()"),
        (r"res\.write\s*\(", "res.write()"),
        (r"res\.end\s*\(", "res.end()"),
        (r"ctx\.body\s*=", "ctx.body ="),
        (r"response\.send\s*\(", "response.send()"),
        (r"response\.json\s*\(", "response.json()"),
        (r"return\s+Response\s*\(", "return Response()"),
        (r"new\s+Response\s*\(", "new Response()"),
    ],
    "logging": [
        (r"console\.(log|warn|error|info|debug)\s*\(", "console.*()"),
        (r"logger\.(log|warn|error|info|debug)\s*\(", "logger.*()"),
        (r"winston\.(log|warn|error|info|debug)\s*\(", "winston.*()"),
        (r"pino\.(log|warn|error|info|debug)\s*\(", "pino.*()"),
    ],
    "file_write": [
        (r"fs\.writeFile\s*\(", "fs.writeFile"),
        (r"fs\.writeFileSync\s*\(", "fs.writeFileSync"),
        (r"writeFile\s*\(", "writeFile"),
        (r"writeFileSync\s*\(", "writeFileSync"),
        (r"fs\.appendFile\s*\(", "fs.appendFile"),
        (r"Deno\.writeTextFile\s*\(", "Deno.writeTextFile"),
        (r"Bun\.write\s*\(", "Bun.write"),
    ],
    "template": [
        (r"\.render\s*\(", ".render()"),
        (r"dangerouslySetInnerHTML\s*=", "dangerouslySetInnerHTML"),
        (r"\.innerHTML\s*=", ".innerHTML ="),
        (r"\.outerHTML\s*=", ".outerHTML ="),
        (r"document\.write\s*\(", "document.write"),
        (r"insertAdjacentHTML\s*\(", "insertAdjacentHTML"),
    ],
    "redirect": [
        (r"res\.redirect\s*\(", "res.redirect()"),
        (r"response\.redirect\s*\(", "response.redirect()"),
        (r"ctx\.redirect\s*\(", "ctx.redirect()"),
        (r"window\.location\s*=", "window.location ="),
        (r"location\.href\s*=", "location.href ="),
        (r"location\.assign\s*\(", "location.assign()"),
        (r"location\.replace\s*\(", "location.replace()"),
    ],
}


# =============================================================================
# TypeScript Null Patterns
# =============================================================================

TYPESCRIPT_NULL_PATTERNS: list[tuple[str, str, str]] = [
    # (pattern, description, reason)
    (r"\.findOne\s*\(", ".findOne()", "database_query"),
    (r"\.findFirst\s*\(", ".findFirst()", "database_query"),
    (r"\.findUnique\s*\(", ".findUnique()", "database_query"),
    (r"\.get\s*\([^)]*\)\s*(?!\!)", ".get() without assertion", "map_lookup"),
    (r"Map\.prototype\.get\s*\(", "Map.get()", "map_lookup"),
    (r"\.get\s*<", "Map.get<>()", "map_lookup"),
    (r"JSON\.parse\s*\(", "JSON.parse()", "json_parse"),
    (r"JSON\.parseAsync\s*\(", "JSON.parseAsync()", "json_parse"),
    (r"\?\.", "optional chaining", "optional_chain"),
    (r"\?\?\s*", "nullish coalescing", "nullish_coalesce"),
    (r"\bas\s+\w+(?:\s*\|\s*null)?", "type assertion", "type_assertion"),
    (r"<\w+>", "type cast", "type_cast"),
    (r"!\s*[;,\)]", "non-null assertion", "non_null_assertion"),
    (r"await\s+\w+\.catch\s*\(", "caught promise", "promise_catch"),
]


# =============================================================================
# Python Sanitizer Patterns
# =============================================================================

PYTHON_SANITIZERS: dict[str, list[str]] = {
    "database": [
        r"%s",  # Parameterized query placeholder
        r"\?",  # SQLite placeholder
        r":\w+",  # Named parameter
        r"\.filter\s*\(",  # ORM filter (safe)
        r"\.values\s*\(",  # ORM values (safe)
    ],
    "command_exec": [
        r"shlex\.quote\s*\(",
        r"shlex\.split\s*\(",
        r"pipes\.quote\s*\(",
        r"shell=False",
    ],
    "template": [
        r"escape\s*\(",
        r"Markup\.escape\s*\(",
        r"bleach\.",
        r"html\.escape\s*\(",
    ],
    "redirect": [
        r"url_for\s*\(",
        r"reverse\s*\(",
        r"is_safe_url\s*\(",
    ],
}


# =============================================================================
# TypeScript Sanitizer Patterns
# =============================================================================

TYPESCRIPT_SANITIZERS: dict[str, list[str]] = {
    "database": [
        r"\$\d+",  # PostgreSQL placeholder
        r"\?",  # MySQL placeholder
        r"Prisma\.\w+",  # Prisma (uses parameterized queries)
        r"\.prepare\s*\(",  # Prepared statement
        r"sql`",  # Tagged template literal (often safe)
    ],
    "command_exec": [
        r"shellEscape\s*\(",
        r"escapeShellArg\s*\(",
        r"shellescape\s*\(",
    ],
    "template": [
        r"escape\s*\(",
        r"escapeHtml\s*\(",
        r"sanitize\s*\(",
        r"DOMPurify\.",
        r"xss\s*\(",
    ],
    "redirect": [
        r"encodeURIComponent\s*\(",
        r"encodeURI\s*\(",
        r"isValidUrl\s*\(",
        r"isSafeUrl\s*\(",
    ],
}


# =============================================================================
# Risk Calculation
# =============================================================================

# Source type to risk weight
SOURCE_RISK_WEIGHT: dict[str, int] = {
    "http_body": 10,
    "http_query": 10,
    "http_header": 8,
    "http_path": 9,
    "external_api": 7,
    "file_read": 6,
    "database": 5,
    "env_var": 3,
}

# Sink type to risk weight
SINK_RISK_WEIGHT: dict[str, int] = {
    "command_exec": 10,
    "database": 9,
    "template": 8,
    "redirect": 7,
    "file_write": 6,
    "http_response": 4,
    "logging": 2,
}


def hash_flow(source: Source, sink: Sink) -> str:
    """Generate a unique hash for a source-sink flow."""
    data = (
        f"{source.file}:{source.line}:{source.source_type}:"
        f"{sink.file}:{sink.line}:{sink.sink_type}"
    )
    return hashlib.sha256(data.encode()).hexdigest()[:16]


def calculate_risk(source: Source, sink: Sink, sanitized: bool = False) -> str:
    """Calculate risk level matching Go implementation."""
    if sanitized:
        return "info"

    # Critical: User input to exec or database
    http_inputs = {"http_body", "http_query", "http_header", "http_path", "user_input"}
    if source.source_type in http_inputs:
        if sink.sink_type in {"command_exec", "database"}:
            return "critical"
        if sink.sink_type in {"http_response", "template", "redirect"}:
            return "high"

    # Medium: Env vars or file to dangerous sinks
    if source.source_type == "env_var" and sink.sink_type in {"database", "command_exec"}:
        return "medium"
    if source.source_type == "file_read" and sink.sink_type in {"database", "command_exec"}:
        return "medium"
    if sink.sink_type == "file_write":
        return "medium"

    # Low: Logging
    if sink.sink_type == "logging":
        return "low"

    return "info"


def extract_variable(line: str, match_start: int) -> str:
    """Extract variable name from line at match position."""
    # Look backwards for assignment
    prefix = line[:match_start].rstrip()

    # Pattern: variable = matched_pattern
    assign_match = re.search(r"(\w+)\s*=\s*$", prefix)
    if assign_match:
        return assign_match.group(1)

    # Pattern: const/let/var variable = matched_pattern
    decl_match = re.search(r"(?:const|let|var)\s+(\w+)\s*=\s*$", prefix)
    if decl_match:
        return decl_match.group(1)

    # Pattern: variable := matched_pattern (Go-style, but check anyway)
    walrus_match = re.search(r"(\w+)\s*:=\s*$", prefix)
    if walrus_match:
        return walrus_match.group(1)

    # Pattern: variable: Type = matched_pattern (TypeScript)
    typed_match = re.search(r"(\w+)\s*:\s*\w+\s*=\s*$", prefix)
    if typed_match:
        return typed_match.group(1)

    # If no assignment found, extract from the matched expression
    # Look for property access: obj.property
    suffix = line[match_start:]
    prop_match = re.match(r"\w+\.(\w+)", suffix)
    if prop_match:
        return prop_match.group(1)

    # Fallback: first word in the match
    word_match = re.match(r"(\w+)", suffix)
    if word_match:
        return word_match.group(1)

    return "<unknown>"


def detect_sources(
    files: list[tuple[str, str]], language: str
) -> list[Source]:
    """Detect all data sources in the given files."""
    sources: list[Source] = []

    if language == "python":
        patterns = PYTHON_SOURCE_PATTERNS
    elif language == "typescript":
        patterns = TYPESCRIPT_SOURCE_PATTERNS
    else:
        return sources

    for file_path, content in files:
        lines = content.splitlines()
        for line_num, line in enumerate(lines, start=1):
            for source_type, pattern_list in patterns.items():
                for pattern, desc in pattern_list:
                    for match in re.finditer(pattern, line):
                        variable = extract_variable(line, match.start())
                        sources.append(
                            Source(
                                source_type=source_type,
                                variable=variable,
                                file=file_path,
                                line=line_num,
                                column=match.start() + 1,
                                pattern=desc,
                            )
                        )

    return sources


def detect_sinks(files: list[tuple[str, str]], language: str) -> list[Sink]:
    """Detect all data sinks in the given files."""
    sinks: list[Sink] = []

    if language == "python":
        patterns = PYTHON_SINK_PATTERNS
    elif language == "typescript":
        patterns = TYPESCRIPT_SINK_PATTERNS
    else:
        return sinks

    for file_path, content in files:
        lines = content.splitlines()
        for line_num, line in enumerate(lines, start=1):
            for sink_type, pattern_list in patterns.items():
                for pattern, desc in pattern_list:
                    for match in re.finditer(pattern, line):
                        variable = extract_variable(line, match.start())
                        sinks.append(
                            Sink(
                                sink_type=sink_type,
                                variable=variable,
                                file=file_path,
                                line=line_num,
                                column=match.start() + 1,
                                pattern=desc,
                                context=line,
                            )
                        )

    return sinks


def detect_null_sources(
    files: list[tuple[str, str]], language: str
) -> list[NilSource]:
    """Detect potentially null/nil variable sources."""
    nil_sources: list[NilSource] = []

    if language != "typescript":
        return nil_sources  # Only TypeScript for now

    for file_path, content in files:
        lines = content.splitlines()
        for line_num, line in enumerate(lines, start=1):
            for pattern, desc, reason in TYPESCRIPT_NULL_PATTERNS:
                for match in re.finditer(pattern, line):
                    variable = extract_variable(line, match.start())
                    nil_sources.append(
                        NilSource(
                            variable=variable,
                            file=file_path,
                            line=line_num,
                            column=match.start() + 1,
                            pattern=desc,
                            reason=reason,
                        )
                    )

    return nil_sources


def check_sanitization(source: Source, sink: Sink, language: str) -> tuple[bool, Optional[str]]:
    """Check if data flow is sanitized between source and sink."""
    sanitizers = {
        "python": {
            "database": [r"execute\s*\([^,]+,\s*[\[\(]", r"\?\s*,", r"%s"],  # Parameterized queries
            "command_exec": [r"shlex\.quote", r"subprocess\.\w+\([^)]*shell\s*=\s*False"],
            "http_response": [r"escape", r"html\.escape", r"markupsafe", r"bleach"],
            "template": [r"autoescape\s*=\s*True", r"\|safe\b"],
        },
        "typescript": {
            "database": [r"\$\d+", r"\?", r"prepare\("],  # Parameterized queries
            "command_exec": [r"shell:\s*false"],
            "http_response": [r"escapeHtml", r"sanitize", r"DOMPurify", r"xss"],
            "template": [r"textContent", r"createTextNode"],
        }
    }

    lang_sanitizers = sanitizers.get(language, {})
    sink_sanitizers = lang_sanitizers.get(sink.sink_type, [])

    # Check sink context for sanitization patterns
    for pattern in sink_sanitizers:
        if re.search(pattern, sink.context, re.IGNORECASE):
            return True, pattern

    return False, None


def track_flows(
    sources: list[Source], sinks: list[Sink], language: str
) -> list[Flow]:
    """Track data flows from sources to sinks."""
    flows: list[Flow] = []
    seen_flows: set[str] = set()

    for source in sources:
        for sink in sinks:
            # Only connect sources and sinks in the same file
            # or within reasonable proximity (same module)
            if source.file != sink.file:
                continue

            # Source must come before sink (simplistic)
            if source.line > sink.line:
                continue

            # Generate flow ID
            flow_id = hash_flow(source, sink)
            if flow_id in seen_flows:
                continue
            seen_flows.add(flow_id)

            # Check for sanitization first (needed for risk calculation)
            sanitized, sanitizer = check_sanitization(source, sink, language)

            # Calculate risk (uses sanitized flag)
            risk = calculate_risk(source, sink, sanitized)

            flows.append(
                Flow(
                    id=flow_id,
                    source=source,
                    sink=sink,
                    risk=risk,
                    sanitized=sanitized,
                    sanitizer=sanitizer,
                )
            )

    return flows


def analyze(files: list[str], language: str) -> AnalysisOutput:
    """
    Perform full data flow analysis on the given files.

    Args:
        files: List of file paths to analyze
        language: Language to use (python or typescript)

    Returns:
        AnalysisOutput with sources, sinks, flows, and nil_sources
    """
    output = AnalysisOutput()

    if not files:
        return output

    # Filter out non-existent files and read content
    file_contents: list[tuple[str, str]] = []
    for file_path in files:
        if not os.path.isfile(file_path):
            continue

        try:
            file_size = os.path.getsize(file_path)
            if file_size > MAX_FILE_SIZE:
                continue  # Skip files that are too large
        except OSError:
            continue

        try:
            with open(file_path, encoding="utf-8") as f:
                content = f.read()
            file_contents.append((file_path, content))
        except (OSError, UnicodeDecodeError):
            continue

    if not file_contents:
        output.error = "No valid files to analyze"
        return output

    # Detect sources
    output.sources = detect_sources(file_contents, language)

    # Detect sinks
    output.sinks = detect_sinks(file_contents, language)

    # Track flows
    output.flows = track_flows(output.sources, output.sinks, language)

    # Detect null sources (TypeScript only)
    output.nil_sources = detect_null_sources(file_contents, language)

    return output


def to_dict(obj) -> dict | list | str | int | bool | None:
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
                result[field_name] = to_dict(value)
        return result
    return obj


def output_to_json(output: AnalysisOutput) -> str:
    """Convert AnalysisOutput to JSON string."""
    data = {
        "sources": [to_dict(s) for s in output.sources],
        "sinks": [to_dict(s) for s in output.sinks],
        "flows": [to_dict(f) for f in output.flows],
        "nil_sources": [to_dict(n) for n in output.nil_sources],
    }
    if output.error:
        data["error"] = output.error
    return json.dumps(data, indent=2)


def main() -> None:
    """Main CLI entry point."""
    args = sys.argv[1:]

    if len(args) < 2:
        print(
            "Usage: data_flow.py <language> <file1> [file2] ...",
            file=sys.stderr,
        )
        print("", file=sys.stderr)
        print(
            "Analyzes files for data flow vulnerabilities.",
            file=sys.stderr,
        )
        print("", file=sys.stderr)
        print("Languages:", file=sys.stderr)
        print("  python     Python (Flask, Django, FastAPI)", file=sys.stderr)
        print("  typescript TypeScript/JavaScript (Express, Node)", file=sys.stderr)
        print("", file=sys.stderr)
        print("Output:", file=sys.stderr)
        print("  JSON to stdout with sources, sinks, flows, nil_sources", file=sys.stderr)
        sys.exit(1)

    language = args[0].lower()
    if language not in ("python", "typescript"):
        print(
            f"Error: Unknown language '{language}'. Use 'python' or 'typescript'.",
            file=sys.stderr,
        )
        sys.exit(1)

    files = args[1:]
    if not files:
        print("Error: No files specified", file=sys.stderr)
        sys.exit(1)

    # Resolve file paths
    resolved_files = [os.path.abspath(f) for f in files]

    try:
        result = analyze(resolved_files, language)
        print(output_to_json(result))
    except Exception as e:
        output = AnalysisOutput(error=f"{type(e).__name__}: {e}")
        print(output_to_json(output))
        sys.exit(1)


if __name__ == "__main__":
    main()
