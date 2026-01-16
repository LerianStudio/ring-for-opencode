import * as ts from 'typescript';
import * as fs from 'fs';
import * as path from 'path';

interface Param {
  name: string;
  type: string;
}

interface FuncSig {
  params: Param[];
  returns: string[];
  is_async: boolean;
  is_exported: boolean;
  start_line: number;
  end_line: number;
}

interface FunctionDiff {
  name: string;
  change_type: 'added' | 'removed' | 'modified' | 'renamed';
  before?: FuncSig;
  after?: FuncSig;
  body_diff?: string;
}

interface FieldDiff {
  name: string;
  change_type: 'added' | 'removed' | 'modified';
  old_type?: string;
  new_type?: string;
}

interface TypeDiff {
  name: string;
  kind: string;
  change_type: 'added' | 'removed' | 'modified' | 'renamed';
  fields?: FieldDiff[];
  start_line: number;
  end_line: number;
}

interface ImportDiff {
  path: string;
  alias?: string;
  change_type: 'added' | 'removed';
}

interface ChangeSummary {
  functions_added: number;
  functions_removed: number;
  functions_modified: number;
  types_added: number;
  types_removed: number;
  types_modified: number;
  imports_added: number;
  imports_removed: number;
}

interface SemanticDiff {
  language: string;
  file_path: string;
  functions: FunctionDiff[];
  types: TypeDiff[];
  imports: ImportDiff[];
  summary: ChangeSummary;
  error?: string;
}

interface ParsedFunc {
  name: string;
  params: Param[];
  returns: string[];
  isAsync: boolean;
  isExported: boolean;
  startLine: number;
  endLine: number;
  bodyText: string;
}

interface ParsedType {
  name: string;
  kind: string;
  fields: Map<string, string>;
  isExported: boolean;
  startLine: number;
  endLine: number;
}

interface ParsedFile {
  functions: Map<string, ParsedFunc>;
  types: Map<string, ParsedType>;
  imports: Map<string, string>;
}

function parseFile(filePath: string): ParsedFile {
  const result: ParsedFile = {
    functions: new Map(),
    types: new Map(),
    imports: new Map(),
  };

  if (!filePath || !fs.existsSync(filePath)) {
    return result;
  }

  const content = fs.readFileSync(filePath, 'utf-8');
  const sourceFile = ts.createSourceFile(
    filePath,
    content,
    ts.ScriptTarget.Latest,
    true
  );

  function getLineNumber(pos: number): number {
    return sourceFile.getLineAndCharacterOfPosition(pos).line + 1;
  }

  function typeToString(type: ts.TypeNode | undefined): string {
    if (!type) return 'any';
    return content.substring(type.pos, type.end).trim();
  }

  function isExported(node: ts.Node): boolean {
    return (
      ts.canHaveModifiers(node) &&
      ts.getModifiers(node)?.some(
        (m) => m.kind === ts.SyntaxKind.ExportKeyword
      ) || false
    );
  }

  function visit(node: ts.Node) {
    // Extract imports
    if (ts.isImportDeclaration(node)) {
      const moduleSpecifier = node.moduleSpecifier;
      if (ts.isStringLiteral(moduleSpecifier)) {
        const importPath = moduleSpecifier.text;
        let alias = '';
        if (node.importClause?.name) {
          alias = node.importClause.name.text;
        }
        result.imports.set(importPath, alias);
      }
    }

    // Extract functions
    if (ts.isFunctionDeclaration(node) && node.name) {
      const func: ParsedFunc = {
        name: node.name.text,
        params: [],
        returns: [],
        isAsync: node.modifiers?.some(
          (m) => m.kind === ts.SyntaxKind.AsyncKeyword
        ) || false,
        isExported: isExported(node),
        startLine: getLineNumber(node.pos),
        endLine: getLineNumber(node.end),
        bodyText: node.body ? content.substring(node.body.pos, node.body.end) : '',
      };

      node.parameters.forEach((param) => {
        func.params.push({
          name: param.name.getText(sourceFile),
          type: typeToString(param.type),
        });
      });

      if (node.type) {
        func.returns.push(typeToString(node.type));
      }

      result.functions.set(func.name, func);
    }

    // Extract arrow functions assigned to const
    if (ts.isVariableStatement(node)) {
      const exported = isExported(node);
      node.declarationList.declarations.forEach((decl) => {
        if (
          ts.isIdentifier(decl.name) &&
          decl.initializer &&
          ts.isArrowFunction(decl.initializer)
        ) {
          const arrow = decl.initializer;
          const func: ParsedFunc = {
            name: decl.name.text,
            params: [],
            returns: [],
            isAsync: arrow.modifiers?.some(
              (m) => m.kind === ts.SyntaxKind.AsyncKeyword
            ) || false,
            isExported: exported,
            startLine: getLineNumber(node.pos),
            endLine: getLineNumber(node.end),
            bodyText: content.substring(arrow.body.pos, arrow.body.end),
          };

          arrow.parameters.forEach((param) => {
            func.params.push({
              name: param.name.getText(sourceFile),
              type: typeToString(param.type),
            });
          });

          if (arrow.type) {
            func.returns.push(typeToString(arrow.type));
          }

          result.functions.set(func.name, func);
        }
      });
    }

    // Extract interfaces
    if (ts.isInterfaceDeclaration(node)) {
      const parsedType: ParsedType = {
        name: node.name.text,
        kind: 'interface',
        fields: new Map(),
        isExported: isExported(node),
        startLine: getLineNumber(node.pos),
        endLine: getLineNumber(node.end),
      };

      node.members.forEach((member) => {
        if (ts.isPropertySignature(member) && member.name) {
          const name = member.name.getText(sourceFile);
          const type = typeToString(member.type);
          parsedType.fields.set(name, type);
        }
      });

      result.types.set(parsedType.name, parsedType);
    }

    // Extract type aliases
    if (ts.isTypeAliasDeclaration(node)) {
      const parsedType: ParsedType = {
        name: node.name.text,
        kind: 'type',
        fields: new Map(),
        isExported: isExported(node),
        startLine: getLineNumber(node.pos),
        endLine: getLineNumber(node.end),
      };

      if (ts.isTypeLiteralNode(node.type)) {
        node.type.members.forEach((member) => {
          if (ts.isPropertySignature(member) && member.name) {
            const name = member.name.getText(sourceFile);
            const type = typeToString(member.type);
            parsedType.fields.set(name, type);
          }
        });
      }

      result.types.set(parsedType.name, parsedType);
    }

    // Extract classes
    if (ts.isClassDeclaration(node) && node.name) {
      const parsedType: ParsedType = {
        name: node.name.text,
        kind: 'class',
        fields: new Map(),
        isExported: isExported(node),
        startLine: getLineNumber(node.pos),
        endLine: getLineNumber(node.end),
      };

      node.members.forEach((member) => {
        if (ts.isPropertyDeclaration(member) && member.name) {
          const name = member.name.getText(sourceFile);
          const type = typeToString(member.type);
          parsedType.fields.set(name, type);
        }
        // Extract class methods as functions
        if (ts.isMethodDeclaration(member) && member.name) {
          const methodName = `${node.name!.text}.${member.name.getText(sourceFile)}`;
          const func: ParsedFunc = {
            name: methodName,
            params: [],
            returns: [],
            isAsync: member.modifiers?.some(
              (m) => m.kind === ts.SyntaxKind.AsyncKeyword
            ) || false,
            isExported: isExported(node),
            startLine: getLineNumber(member.pos),
            endLine: getLineNumber(member.end),
            bodyText: member.body ? content.substring(member.body.pos, member.body.end) : '',
          };

          member.parameters.forEach((param) => {
            func.params.push({
              name: param.name.getText(sourceFile),
              type: typeToString(param.type),
            });
          });

          if (member.type) {
            func.returns.push(typeToString(member.type));
          }

          result.functions.set(func.name, func);
        }
      });

      result.types.set(parsedType.name, parsedType);
    }

    ts.forEachChild(node, visit);
  }

  visit(sourceFile);
  return result;
}

function compareFunctions(
  before: Map<string, ParsedFunc>,
  after: Map<string, ParsedFunc>
): FunctionDiff[] {
  const diffs: FunctionDiff[] = [];

  // Find removed and modified
  before.forEach((beforeFunc, name) => {
    const afterFunc = after.get(name);
    if (!afterFunc) {
      diffs.push({
        name,
        change_type: 'removed',
        before: {
          params: beforeFunc.params,
          returns: beforeFunc.returns,
          is_async: beforeFunc.isAsync,
          is_exported: beforeFunc.isExported,
          start_line: beforeFunc.startLine,
          end_line: beforeFunc.endLine,
        },
      });
      return;
    }

    // Check for modifications
    const sigChanged =
      JSON.stringify(beforeFunc.params) !== JSON.stringify(afterFunc.params) ||
      JSON.stringify(beforeFunc.returns) !== JSON.stringify(afterFunc.returns) ||
      beforeFunc.isAsync !== afterFunc.isAsync;

    const bodyChanged = beforeFunc.bodyText !== afterFunc.bodyText;

    if (sigChanged || bodyChanged) {
      const changes: string[] = [];
      if (JSON.stringify(beforeFunc.params) !== JSON.stringify(afterFunc.params)) {
        changes.push('parameters changed');
      }
      if (JSON.stringify(beforeFunc.returns) !== JSON.stringify(afterFunc.returns)) {
        changes.push('return type changed');
      }
      if (beforeFunc.isAsync !== afterFunc.isAsync) {
        changes.push('async modifier changed');
      }
      if (bodyChanged) {
        changes.push('implementation changed');
      }

      diffs.push({
        name,
        change_type: 'modified',
        before: {
          params: beforeFunc.params,
          returns: beforeFunc.returns,
          is_async: beforeFunc.isAsync,
          is_exported: beforeFunc.isExported,
          start_line: beforeFunc.startLine,
          end_line: beforeFunc.endLine,
        },
        after: {
          params: afterFunc.params,
          returns: afterFunc.returns,
          is_async: afterFunc.isAsync,
          is_exported: afterFunc.isExported,
          start_line: afterFunc.startLine,
          end_line: afterFunc.endLine,
        },
        body_diff: changes.join(', '),
      });
    }
  });

  // Find added
  after.forEach((afterFunc, name) => {
    if (!before.has(name)) {
      diffs.push({
        name,
        change_type: 'added',
        after: {
          params: afterFunc.params,
          returns: afterFunc.returns,
          is_async: afterFunc.isAsync,
          is_exported: afterFunc.isExported,
          start_line: afterFunc.startLine,
          end_line: afterFunc.endLine,
        },
      });
    }
  });

  return diffs;
}

function compareTypes(
  before: Map<string, ParsedType>,
  after: Map<string, ParsedType>
): TypeDiff[] {
  const diffs: TypeDiff[] = [];

  before.forEach((beforeType, name) => {
    const afterType = after.get(name);
    if (!afterType) {
      diffs.push({
        name,
        kind: beforeType.kind,
        change_type: 'removed',
        start_line: beforeType.startLine,
        end_line: beforeType.endLine,
      });
      return;
    }

    // Compare fields
    const fieldDiffs: FieldDiff[] = [];
    beforeType.fields.forEach((type, fieldName) => {
      const afterFieldType = afterType.fields.get(fieldName);
      if (!afterFieldType) {
        fieldDiffs.push({
          name: fieldName,
          change_type: 'removed',
          old_type: type,
        });
      } else if (afterFieldType !== type) {
        fieldDiffs.push({
          name: fieldName,
          change_type: 'modified',
          old_type: type,
          new_type: afterFieldType,
        });
      }
    });

    afterType.fields.forEach((type, fieldName) => {
      if (!beforeType.fields.has(fieldName)) {
        fieldDiffs.push({
          name: fieldName,
          change_type: 'added',
          new_type: type,
        });
      }
    });

    if (fieldDiffs.length > 0 || beforeType.kind !== afterType.kind) {
      diffs.push({
        name,
        kind: afterType.kind,
        change_type: 'modified',
        fields: fieldDiffs,
        start_line: afterType.startLine,
        end_line: afterType.endLine,
      });
    }
  });

  after.forEach((afterType, name) => {
    if (!before.has(name)) {
      diffs.push({
        name,
        kind: afterType.kind,
        change_type: 'added',
        start_line: afterType.startLine,
        end_line: afterType.endLine,
      });
    }
  });

  return diffs;
}

function compareImports(
  before: Map<string, string>,
  after: Map<string, string>
): ImportDiff[] {
  const diffs: ImportDiff[] = [];

  before.forEach((alias, importPath) => {
    if (!after.has(importPath)) {
      diffs.push({
        path: importPath,
        alias: alias || undefined,
        change_type: 'removed',
      });
    }
  });

  after.forEach((alias, importPath) => {
    if (!before.has(importPath)) {
      diffs.push({
        path: importPath,
        alias: alias || undefined,
        change_type: 'added',
      });
    }
  });

  return diffs;
}

function extractDiff(beforePath: string, afterPath: string): SemanticDiff {
  const before = parseFile(beforePath);
  const after = parseFile(afterPath);

  const functions = compareFunctions(before.functions, after.functions);
  const types = compareTypes(before.types, after.types);
  const imports = compareImports(before.imports, after.imports);

  const summary: ChangeSummary = {
    functions_added: functions.filter((f) => f.change_type === 'added').length,
    functions_removed: functions.filter((f) => f.change_type === 'removed').length,
    functions_modified: functions.filter((f) => f.change_type === 'modified').length,
    types_added: types.filter((t) => t.change_type === 'added').length,
    types_removed: types.filter((t) => t.change_type === 'removed').length,
    types_modified: types.filter((t) => t.change_type === 'modified').length,
    imports_added: imports.filter((i) => i.change_type === 'added').length,
    imports_removed: imports.filter((i) => i.change_type === 'removed').length,
  };

  return {
    language: 'typescript',
    file_path: afterPath || beforePath,
    functions,
    types,
    imports,
    summary,
  };
}

// CLI argument parsing
function parseArgs(args: string[]): { before: string; after: string } {
  let before = '';
  let after = '';

  for (let i = 0; i < args.length; i++) {
    const arg = args[i];
    if (arg === '--before' && i + 1 < args.length) {
      before = args[i + 1];
      i++;
    } else if (arg === '--after' && i + 1 < args.length) {
      after = args[i + 1];
      i++;
    } else if (!arg.startsWith('--')) {
      // Positional arguments for backward compatibility
      // Go wrapper calls: node script.js <before> <after>
      if (!before) {
        before = arg;
      } else if (!after) {
        after = arg;
      }
    }
  }

  // Handle empty string markers (Go wrapper sends '""' for empty paths)
  if (before === '""' || before === "''") {
    before = '';
  }
  if (after === '""' || after === "''") {
    after = '';
  }

  return { before, after };
}

// CLI entry point
function main() {
  const args = process.argv.slice(2);

  if (args.length < 2) {
    console.error('Usage: ast-extractor --before <before-path> --after <after-path>');
    console.error('Use empty string "" for new/deleted files');
    process.exit(1);
  }

  const { before: beforePath, after: afterPath } = parseArgs(args);

  if (!beforePath && !afterPath) {
    console.error('Error: At least one of --before or --after must be specified');
    process.exit(1);
  }

  try {
    const diff = extractDiff(beforePath, afterPath);
    console.log(JSON.stringify(diff, null, 2));
  } catch (error) {
    const diff: SemanticDiff = {
      language: 'typescript',
      file_path: afterPath || beforePath,
      functions: [],
      types: [],
      imports: [],
      summary: {
        functions_added: 0,
        functions_removed: 0,
        functions_modified: 0,
        types_added: 0,
        types_removed: 0,
        types_modified: 0,
        imports_added: 0,
        imports_removed: 0,
      },
      error: error instanceof Error ? error.message : String(error),
    };
    console.log(JSON.stringify(diff, null, 2));
    process.exit(1);
  }
}

main();
