---
name: "ring:brainstorm"
description: Interactive design refinement using Socratic method
agent: build
subtask: false
---

Transform rough ideas into fully-formed designs through structured questioning and alternative exploration.

## Process Phases

### Phase 1: Autonomous Recon
- Inspect repository structure, documentation, and recent commits
- Form initial understanding of the codebase context
- Share findings before asking questions

### Phase 2: Understanding
- Share synthesized understanding for validation
- Ask targeted questions (max 3) to fill knowledge gaps
- Gather: purpose, constraints, success criteria

### Phase 3: Exploration
- Propose 2-3 different architectural approaches
- Present trade-offs for each option
- Recommend preferred approach with rationale
- Ask user to select approach

### Phase 4: Design Presentation
- Present design in 200-300 word sections
- Cover: architecture, components, data flow, error handling, testing
- Validate each section incrementally
- Require explicit approval ("Approved", "Looks good", "Proceed")

### Phase 5: Documentation
- Write validated design to `docs/plans/YYYY-MM-DD-<topic>-design.md`
- Commit the design document

### Phase 6: Next Steps (Optional)
- If implementing: Create detailed implementation plan
- Break design into bite-sized executable tasks

## Guidelines

- Maximum 3 questions per phase
- Explore codebase before asking questions
- Design must be explicitly approved before proceeding
- Responses like "interesting" do NOT count as approval

$ARGUMENTS
