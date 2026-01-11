# Doubt Resolver Tool (`ring_doubt`)

Use the `ring_doubt` tool when you have genuine uncertainty and need the human to choose an option.

## When to Use

Use this tool when:
- Multiple approaches are viable
- The choice affects correctness or scope
- The decision cannot be resolved from repo conventions or existing patterns

Do NOT use it for trivial clarifications.

## Tool Contract

The tool prints a numbered list and the user replies normally.

**Supports:**
- Single-select (default)
- Multi-select (`multi: true`)
- Free-text alternative (`allowCustom: true` by default)

## Example Call

Call the tool with:
- `question`: what decision is needed
- `options`: list of option strings
- `multi`: optional
- `allowCustom`: optional
- `context`: optional

Example:

```json
{
  "question": "Which approach should I take for session storage?",
  "options": [
    "Keep JSON files in .opencode/state",
    "Move to SQLite in .ring/cache",
    "Defer storage; keep in-memory only"
  ],
  "multi": false,
  "allowCustom": true,
  "context": "We need persistence across /clear and crash recovery."
}
```
