---
name: "ring:dev-report"
description: View the feedback report from the last development cycle
agent: build
---

View the feedback report from the last development cycle.

## Usage

```
/ring:dev-report [cycle-date]
```

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `cycle-date` | No | Date of the cycle (YYYY-MM-DD). Defaults to most recent. |

## Examples

```bash
# View most recent report
/ring:dev-report

# View specific date
/ring:dev-report 2024-01-15
```

## Report Contents

The feedback report includes:

### Summary
- Total tasks processed
- Success/partial/failed counts
- Average assertiveness score
- Total cycle duration

### Per-Task Metrics
- Assertiveness score
- Duration per gate
- Iteration counts
- Issues encountered

### Analysis
- Gates with most retrabalho
- Recurring failure patterns
- Improvement suggestions

### Recommendations
- Suggested skill improvements
- Suggested agent adjustments
- Process optimizations

## Example Output

```
📊 Development Cycle Report

Date: 2024-01-15
Duration: 2h 45m

┌──────────────────────────────────────────┐
│ Summary                                   │
├──────────────────────────────────────────┤
│ Tasks: 5 total                           │
│   ✅ SUCCESS: 4 (80%)                    │
│   ⚠️ PARTIAL: 1 (20%)                    │
│   ❌ FAILED: 0 (0%)                      │
│                                          │
│ Assertiveness: 87.4% (target: 85%)       │
└──────────────────────────────────────────┘

Top Issues:
1. Gate 5 (testing): 3 tasks needed extra iterations
2. Gate 6 (review): Security findings in 2 tasks

Recommendations:
→ Skill ring:dev-unit-testing: Add test planning phase
→ Agent backend-*: Reinforce input validation

Full report: docs/dev-team/feedback/cycle-2024-01-15.md
```

## Report Location

Reports are saved to: `docs/dev-team/feedback/cycle-YYYY-MM-DD.md`

## Available Reports

List all reports:
```bash
ls docs/dev-team/feedback/
```

## Related Commands

| Command | Description |
|---------|-------------|
| `/ring:dev-cycle` | Start new cycle |
| `/ring:dev-status` | Check current status |
| `/ring:dev-cancel` | Cancel running cycle |

---

Now loading the most recent feedback report...

Search for reports in: `docs/dev-team/feedback/cycle-*.md`

---

## MANDATORY: Load Full Skill

**This command MUST load the skill for complete workflow execution.**

```
Use Skill tool: ring:dev-feedback-loop
```

The skill contains the complete workflow with:
- Metrics collection from ring:dev-cycle
- Pattern analysis
- Improvement recommendations
- Report generation format
- Historical comparison
