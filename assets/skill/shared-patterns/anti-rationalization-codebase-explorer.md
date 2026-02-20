# Anti-Rationalization: ring:codebase-explorer Dispatch (Step 3)

**MANDATORY: Step 3 MUST use `Task(subagent_type="ring:codebase-explorer")`.**

| Rationalization | Why It's WRONG | Required Action |
|-----------------|----------------|-----------------|
| "I'll use Bash find/ls to quickly explore" | Bash cannot analyze patterns, just lists files. ring:codebase-explorer provides architectural analysis. | **Use Task with subagent_type="ring:codebase-explorer"** |
| "The Explore agent is faster" | "Explore" subagent_type is not "ring:codebase-explorer". Different agents. | **Use exact string: "ring:codebase-explorer"** |
| "I already know the structure from find output" | Knowing file paths is not understanding architecture. Agent provides analysis. | **Use Task with subagent_type="ring:codebase-explorer"** |
| "This is a small codebase, Bash is enough" | Size is irrelevant. The agent provides standardized output format required by Step 4. | **Use Task with subagent_type="ring:codebase-explorer"** |
| "I'll explore manually then dispatch agents" | Manual exploration skips the codebase-report.md artifact required for Step 4 gate. | **Use Task with subagent_type="ring:codebase-explorer"** |
