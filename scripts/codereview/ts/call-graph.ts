import * as fs from "fs"
import * as path from "path"
import * as ts from "typescript"

const IGNORED_CALLEE_TARGETS = new Set<string>([
  // Common Node/TS/JS built-ins / globals that usually add noise to call graphs.
  "require",
  "console.log",
  "console.error",
  "console.warn",
  "console.info",
  "console.debug",
  "JSON.stringify",
  "JSON.parse",
  "Object.keys",
  "Object.values",
  "Object.entries",
])

/**
 * Represents a call site within a function.
 */
interface CallSite {
  target: string
  line: number
  column: number
  is_method: boolean
}

/**
 * Represents information about a function caller.
 */
interface CallerInfo {
  function: string
  file: string
  line: number
}

/**
 * Represents comprehensive function information including call relationships.
 */
interface FunctionInfo {
  name: string
  file: string
  line: number
  end_line: number
  call_sites: CallSite[]
  called_by: CallerInfo[]
}

/**
 * Output format for the call graph analysis.
 */
interface CallGraphOutput {
  functions: FunctionInfo[]
  error?: string
}

/**
 * Internal representation of a parsed function during analysis.
 */
interface ParsedFunction {
  name: string
  file: string
  startLine: number
  endLine: number
  callSites: CallSite[]
  node: ts.Node
}

/**
 * Analyzes TypeScript files and builds function-level call relationships.
 *
 * @param filePaths - Array of file paths to analyze
 * @param targetFunctions - Optional array of function names to focus on
 * @returns Call graph analysis output
 */
function analyzeCallGraph(filePaths: string[], targetFunctions?: string[]): CallGraphOutput {
  const result: CallGraphOutput = {
    functions: [],
  }

  if (filePaths.length === 0) {
    return result
  }

  // Filter out non-existent files
  const existingFiles = filePaths.filter((fp) => {
    try {
      return fs.existsSync(fp)
    } catch {
      return false
    }
  })

  if (existingFiles.length === 0) {
    result.error = "No valid files to analyze"
    return result
  }

  const useTypeChecker = !(
    process.env.CALL_GRAPH_LIGHT_MODE === "1" || process.env.CALL_GRAPH_LIGHT_MODE === "true"
  )
  let typeChecker: ts.TypeChecker | null
  const fileSourceMap = new Map<string, ts.SourceFile>()
  const allFunctions: ParsedFunction[] = []

  if (useTypeChecker) {
    const program = ts.createProgram(existingFiles, {
      target: ts.ScriptTarget.Latest,
      module: ts.ModuleKind.CommonJS,
      allowJs: true,
      checkJs: true,
    })
    const programTypeChecker = program.getTypeChecker()
    typeChecker = programTypeChecker

    for (const filePath of existingFiles) {
      const sourceFile = program.getSourceFile(filePath)
      if (!sourceFile) {
        continue
      }
      fileSourceMap.set(filePath, sourceFile)
    }
  } else {
    typeChecker = null
    for (const filePath of existingFiles) {
      try {
        const content = fs.readFileSync(filePath, "utf-8")
        const sourceFile = ts.createSourceFile(filePath, content, ts.ScriptTarget.Latest, true)
        fileSourceMap.set(filePath, sourceFile)
      } catch {}
    }
  }

  if (fileSourceMap.size === 0) {
    result.error = "No readable files to analyze"
    return result
  }

  if (!useTypeChecker) {
    for (const [filePath, sourceFile] of fileSourceMap.entries()) {
      const functions = extractFunctionsLightMode(sourceFile, filePath)
      allFunctions.push(...functions)
    }
  } else {
    if (!typeChecker) {
      result.error = "Type checker unavailable"
      return result
    }

    const activeTypeChecker = typeChecker
    for (const [filePath, sourceFile] of fileSourceMap.entries()) {
      const functions = extractFunctions(sourceFile, filePath, activeTypeChecker)
      allFunctions.push(...functions)
    }
  }

  // Build caller map: function name -> callers
  const callerMap = buildCallerMap(allFunctions, existingFiles)

  // Filter functions if target list provided
  let functionsToReport = allFunctions
  if (targetFunctions && targetFunctions.length > 0) {
    const targetSet = new Set(targetFunctions)
    functionsToReport = allFunctions.filter(
      (fn) => targetSet.has(fn.name) || targetSet.has(getBaseName(fn.name)),
    )
  }

  // Build output
  for (const fn of functionsToReport) {
    const callers = callerMap.get(fn.name) || []

    result.functions.push({
      name: fn.name,
      file: fn.file,
      line: fn.startLine,
      end_line: fn.endLine,
      call_sites: fn.callSites,
      called_by: callers,
    })
  }

  return result
}

function extractFunctionsLightMode(sourceFile: ts.SourceFile, filePath: string): ParsedFunction[] {
  const functions: ParsedFunction[] = []

  function getLineNumber(pos: number): number {
    return sourceFile.getLineAndCharacterOfPosition(pos).line + 1
  }

  function extractCallSites(body: ts.Node): CallSite[] {
    const calls: CallSite[] = []

    function visitCalls(node: ts.Node) {
      if (ts.isCallExpression(node)) {
        calls.push({
          target: node.expression.getText(sourceFile),
          line: getLineNumber(node.getStart(sourceFile)),
          column: 1,
          is_method:
            ts.isPropertyAccessExpression(node.expression) ||
            ts.isElementAccessExpression(node.expression),
        })
      }

      if (ts.isNewExpression(node)) {
        calls.push({
          target: node.expression.getText(sourceFile),
          line: getLineNumber(node.getStart(sourceFile)),
          column: 1,
          is_method: false,
        })
      }

      ts.forEachChild(node, visitCalls)
    }

    visitCalls(body)
    return calls
  }

  function visit(node: ts.Node) {
    if (ts.isFunctionDeclaration(node) && node.name) {
      const callSites = node.body ? extractCallSites(node.body) : []
      functions.push({
        name: node.name.text,
        file: filePath,
        startLine: getLineNumber(node.getStart(sourceFile)),
        endLine: getLineNumber(node.getEnd()),
        callSites,
        node,
      })
    }

    if (ts.isVariableStatement(node)) {
      node.declarationList.declarations.forEach((decl) => {
        if (ts.isIdentifier(decl.name) && decl.initializer) {
          if (ts.isArrowFunction(decl.initializer) || ts.isFunctionExpression(decl.initializer)) {
            const funcExpr = decl.initializer
            const callSites = funcExpr.body ? extractCallSites(funcExpr.body) : []
            functions.push({
              name: decl.name.text,
              file: filePath,
              startLine: getLineNumber(node.getStart(sourceFile)),
              endLine: getLineNumber(node.getEnd()),
              callSites,
              node,
            })
          }
        }
      })
    }

    if (ts.isClassDeclaration(node) && node.name) {
      const className = node.name.text
      node.members.forEach((member) => {
        if (ts.isMethodDeclaration(member) && member.name) {
          const methodName = member.name.getText(sourceFile)
          const fullName = `${className}.${methodName}`
          const callSites = member.body ? extractCallSites(member.body) : []
          functions.push({
            name: fullName,
            file: filePath,
            startLine: getLineNumber(member.getStart(sourceFile)),
            endLine: getLineNumber(member.getEnd()),
            callSites,
            node: member,
          })
        }
      })
    }

    ts.forEachChild(node, visit)
  }

  visit(sourceFile)
  return functions
}

/**
 * Extracts all functions from a source file with their call sites.
 */
function extractFunctions(
  sourceFile: ts.SourceFile,
  filePath: string,
  typeChecker: ts.TypeChecker,
): ParsedFunction[] {
  const functions: ParsedFunction[] = []

  function extractCallSites(body: ts.Node): CallSite[] {
    const calls: CallSite[] = []

    function visitCalls(node: ts.Node) {
      if (ts.isCallExpression(node)) {
        const call = extractCallInfo(node, sourceFile, typeChecker)
        if (call) {
          calls.push(call)
        }
      }

      if (ts.isNewExpression(node)) {
        const call = extractNewExpressionInfo(node, sourceFile, typeChecker)
        if (call) {
          calls.push(call)
        }
      }

      ts.forEachChild(node, visitCalls)
    }

    visitCalls(body)
    return calls
  }

  function getLineNumber(pos: number): number {
    return sourceFile.getLineAndCharacterOfPosition(pos).line + 1
  }

  function visit(node: ts.Node) {
    // Function declarations
    if (ts.isFunctionDeclaration(node) && node.name) {
      const callSites = node.body ? extractCallSites(node.body) : []
      functions.push({
        name: node.name.text,
        file: filePath,
        startLine: getLineNumber(node.getStart(sourceFile)),
        endLine: getLineNumber(node.getEnd()),
        callSites,
        node,
      })
    }

    // Arrow functions assigned to const/let
    if (ts.isVariableStatement(node)) {
      node.declarationList.declarations.forEach((decl) => {
        if (ts.isIdentifier(decl.name) && decl.initializer) {
          if (ts.isArrowFunction(decl.initializer) || ts.isFunctionExpression(decl.initializer)) {
            const funcExpr = decl.initializer
            const callSites = funcExpr.body ? extractCallSites(funcExpr.body) : []
            functions.push({
              name: decl.name.text,
              file: filePath,
              startLine: getLineNumber(node.getStart(sourceFile)),
              endLine: getLineNumber(node.getEnd()),
              callSites,
              node,
            })
          }
        }
      })
    }

    // Class methods
    if (ts.isClassDeclaration(node) && node.name) {
      const className = node.name.text

      node.members.forEach((member) => {
        if (ts.isMethodDeclaration(member) && member.name) {
          const methodName = member.name.getText(sourceFile)
          const fullName = `${className}.${methodName}`
          const callSites = member.body ? extractCallSites(member.body) : []

          functions.push({
            name: fullName,
            file: filePath,
            startLine: getLineNumber(member.getStart(sourceFile)),
            endLine: getLineNumber(member.getEnd()),
            callSites,
            node: member,
          })
        }

        // Constructors
        if (ts.isConstructorDeclaration(member)) {
          const fullName = `${className}.constructor`
          const callSites = member.body ? extractCallSites(member.body) : []

          functions.push({
            name: fullName,
            file: filePath,
            startLine: getLineNumber(member.getStart(sourceFile)),
            endLine: getLineNumber(member.getEnd()),
            callSites,
            node: member,
          })
        }
      })
    }

    // Object literal methods (e.g., in module.exports)
    if (ts.isPropertyAssignment(node)) {
      if (
        node.initializer &&
        (ts.isFunctionExpression(node.initializer) || ts.isArrowFunction(node.initializer))
      ) {
        const funcExpr = node.initializer
        const name = node.name.getText(sourceFile)
        const callSites = funcExpr.body ? extractCallSites(funcExpr.body) : []

        functions.push({
          name,
          file: filePath,
          startLine: getLineNumber(node.getStart(sourceFile)),
          endLine: getLineNumber(node.getEnd()),
          callSites,
          node,
        })
      }
    }

    ts.forEachChild(node, visit)
  }

  visit(sourceFile)
  return functions
}

/**
 * Extracts call information from a CallExpression node.
 */
function extractCallInfo(
  node: ts.CallExpression,
  sourceFile: ts.SourceFile,
  typeChecker: ts.TypeChecker,
): CallSite | null {
  const { line, character } = sourceFile.getLineAndCharacterOfPosition(node.getStart(sourceFile))

  function shouldIgnoreCall(): boolean {
    // `require(...)`
    if (ts.isIdentifier(node.expression) && node.expression.text === "require") {
      return true
    }

    // `console.*(...)`, `JSON.*(...)`, `Object.*(...)`
    const isIgnoredPropertyCall = (objName: string, memberName: string): boolean =>
      IGNORED_CALLEE_TARGETS.has(`${objName}.${memberName}`)

    if (ts.isPropertyAccessExpression(node.expression)) {
      const objExpr = node.expression.expression
      const memberName = node.expression.name.text
      if (ts.isIdentifier(objExpr) && isIgnoredPropertyCall(objExpr.text, memberName)) {
        return true
      }
    }

    // `console['log'](...)` and similar
    if (ts.isElementAccessExpression(node.expression)) {
      const objExpr = node.expression.expression
      const arg = node.expression.argumentExpression
      if (
        ts.isIdentifier(objExpr) &&
        arg &&
        ts.isStringLiteralLike(arg) &&
        isIgnoredPropertyCall(objExpr.text, arg.text)
      ) {
        return true
      }
    }

    return false
  }

  if (shouldIgnoreCall()) {
    return null
  }

  const isMethod =
    ts.isPropertyAccessExpression(node.expression) || ts.isElementAccessExpression(node.expression)

  function resolveCalledSymbol(): ts.Symbol | undefined {
    // Prefer resolving the symbol at the callee expression (usually yields the most human-meaningful symbol,
    // e.g. the variable/function name instead of an internal call-signature symbol like `__call`).
    let symbolFromLocation: ts.Symbol | undefined
    if (ts.isPropertyAccessExpression(node.expression)) {
      symbolFromLocation =
        typeChecker.getSymbolAtLocation(node.expression.name) ??
        typeChecker.getSymbolAtLocation(node.expression)
    } else if (ts.isElementAccessExpression(node.expression)) {
      const arg = node.expression.argumentExpression
      if (arg && ts.isStringLiteralLike(arg)) {
        symbolFromLocation =
          typeChecker.getSymbolAtLocation(arg) ?? typeChecker.getSymbolAtLocation(node.expression)
      } else {
        symbolFromLocation = typeChecker.getSymbolAtLocation(node.expression)
      }
    } else {
      symbolFromLocation = typeChecker.getSymbolAtLocation(node.expression)
    }

    // Also try resolved signature (useful for some call-signature cases where location resolution is missing).
    const signature = typeChecker.getResolvedSignature(node)
    if (signature) {
      const sigAny = signature as unknown as {
        getDeclaration?: () => ts.SignatureDeclaration | undefined
        declaration?: ts.SignatureDeclaration
      }
      const sigDecl = sigAny.getDeclaration ? sigAny.getDeclaration() : sigAny.declaration
      const sigSymbol = (sigDecl as unknown as { symbol?: ts.Symbol } | undefined)?.symbol
      if (!symbolFromLocation && sigSymbol) {
        return sigSymbol
      }
    }

    return symbolFromLocation
  }

  function getCanonicalTargetFromSymbol(sym: ts.Symbol): string | null {
    let symbol = sym
    if (symbol.flags & ts.SymbolFlags.Alias) {
      symbol = typeChecker.getAliasedSymbol(symbol)
    }

    const symbolName = symbol.getName()
    const declarations = symbol.getDeclarations() ?? []
    if (declarations.length === 0 && !symbolName) {
      return null
    }

    // Prefer a stable container-qualified name where possible (e.g., `ClassName.method`).
    for (const decl of declarations) {
      if (ts.isConstructorDeclaration(decl)) {
        const parent = decl.parent
        if ((ts.isClassDeclaration(parent) || ts.isClassExpression(parent)) && parent.name) {
          return `${parent.name.text}.constructor`
        }
      }

      if (
        ts.isMethodDeclaration(decl) ||
        ts.isMethodSignature(decl) ||
        ts.isPropertyDeclaration(decl) ||
        ts.isPropertySignature(decl) ||
        ts.isGetAccessorDeclaration(decl) ||
        ts.isSetAccessorDeclaration(decl)
      ) {
        const parent = decl.parent
        if (
          (ts.isClassDeclaration(parent) ||
            ts.isClassExpression(parent) ||
            ts.isInterfaceDeclaration(parent)) &&
          parent.name
        ) {
          return `${parent.name.text}.${symbolName}`
        }
      }
    }

    // Otherwise, include a file prefix as metadata while keeping the function name suffix stable:
    // `someFile.fnName` so `getBaseName(...)` still yields `fnName`.
    if (declarations.length > 0) {
      const declFile = declarations[0].getSourceFile().fileName
      const fileBase = path.basename(declFile, path.extname(declFile))
      if (fileBase && symbolName) {
        return `${fileBase}.${symbolName}`
      }
    }

    return symbolName || null
  }

  const resolvedSymbol = resolveCalledSymbol()
  let target = resolvedSymbol ? getCanonicalTargetFromSymbol(resolvedSymbol) : null

  // Fallback: only use text-based extraction if the type checker couldn't resolve a symbol.
  if (!target) {
    if (ts.isPropertyAccessExpression(node.expression)) {
      target = node.expression.getText(sourceFile)
    } else if (ts.isIdentifier(node.expression)) {
      target = node.expression.text
    } else if (ts.isElementAccessExpression(node.expression)) {
      target = "<dynamic>"
    } else {
      target = "<anonymous>"
    }

    if (IGNORED_CALLEE_TARGETS.has(target)) {
      return null
    }
  }

  return {
    target,
    line: line + 1,
    column: character + 1,
    is_method: isMethod,
  }
}

/**
 * Extracts call information from a NewExpression node.
 */
function extractNewExpressionInfo(
  node: ts.NewExpression,
  sourceFile: ts.SourceFile,
  typeChecker: ts.TypeChecker,
): CallSite | null {
  const { line, character } = sourceFile.getLineAndCharacterOfPosition(node.getStart(sourceFile))

  let target: string

  if (ts.isIdentifier(node.expression)) {
    target = `new ${node.expression.text}`
  } else if (ts.isPropertyAccessExpression(node.expression)) {
    target = `new ${node.expression.getText(sourceFile)}`
  } else {
    target = "new <anonymous>"
  }

  return {
    target,
    line: line + 1,
    column: character + 1,
    is_method: false,
  }
}

/**
 * Builds a map of function names to their callers.
 */
function buildCallerMap(functions: ParsedFunction[], files: string[]): Map<string, CallerInfo[]> {
  const callerMap = new Map<string, CallerInfo[]>()

  // Initialize map for all functions
  for (const fn of functions) {
    callerMap.set(fn.name, [])
  }

  // Build caller relationships
  for (const callerFn of functions) {
    for (const callSite of callerFn.callSites) {
      // Find the target function
      const targetName = callSite.target
      const baseName = getBaseName(targetName)

      // Try to match with known functions
      for (const targetFn of functions) {
        if (
          targetFn.name === targetName ||
          targetFn.name === baseName ||
          getBaseName(targetFn.name) === baseName
        ) {
          const callers = callerMap.get(targetFn.name) || []
          callers.push({
            function: callerFn.name,
            file: callerFn.file,
            line: callSite.line,
          })
          callerMap.set(targetFn.name, callers)
          break
        }
      }
    }
  }

  return callerMap
}

/**
 * Gets the base name of a function (without class prefix).
 */
function getBaseName(name: string): string {
  const parts = name.split(".")
  return parts[parts.length - 1]
}

/**
 * Parses CLI arguments.
 */
function parseArgs(args: string[]): { files: string[]; functions: string[] } {
  const files: string[] = []
  const functions: string[] = []

  for (let i = 0; i < args.length; i++) {
    const arg = args[i]
    if (arg === "--functions" && i + 1 < args.length) {
      // Parse comma-separated function names
      const funcList = args[i + 1]
      functions.push(
        ...funcList
          .split(",")
          .map((f) => f.trim())
          .filter((f) => f),
      )
      i++
    } else if (!arg.startsWith("--")) {
      files.push(arg)
    }
  }

  return { files, functions }
}

/**
 * Main CLI entry point.
 */
function main() {
  const args = process.argv.slice(2)

  if (args.length === 0) {
    console.error("Usage: call-graph <files...> [--functions func1,func2,...]")
    console.error("")
    console.error("Analyzes TypeScript files and outputs function call relationships.")
    console.error("")
    console.error("Options:")
    console.error("  --functions    Comma-separated list of function names to focus on")
    process.exit(1)
  }

  const { files, functions } = parseArgs(args)

  if (files.length === 0) {
    console.error("Error: No files specified")
    process.exit(1)
  }

  // Resolve file paths
  const resolvedFiles = files.map((f) => path.resolve(f))

  try {
    const targetFunctions = functions.length > 0 ? functions : undefined
    const result = analyzeCallGraph(resolvedFiles, targetFunctions)
    console.log(JSON.stringify(result, null, 2))
  } catch (error) {
    const output: CallGraphOutput = {
      functions: [],
      error: error instanceof Error ? error.message : String(error),
    }
    console.log(JSON.stringify(output, null, 2))
    process.exit(1)
  }
}

main()
