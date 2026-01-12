import { beforeEach, describe, expect, it } from "bun:test"

import type { WorkflowDefinition } from "../../../../plugin/orchestrator/types.js"
import {
  executeWorkflowStep,
  getWorkflow,
  listWorkflows,
  registerWorkflow,
  resetWorkflows,
  runWorkflow,
  validateWorkflowInput,
  type WorkflowRunDependencies,
} from "../../../../plugin/orchestrator/workflow/engine.js"

beforeEach(() => {
  resetWorkflows()
})

describe("Workflow Registry", () => {
  it("registers and retrieves workflows", () => {
    const workflow: WorkflowDefinition = {
      id: "test-workflow",
      name: "Test Workflow",
      description: "A test workflow",
      steps: [{ id: "step1", title: "Step 1", workerId: "coder", prompt: "{task}" }],
    }

    registerWorkflow(workflow)
    const retrieved = getWorkflow("test-workflow")

    expect(retrieved).toBeDefined()
    expect(retrieved?.id).toBe("test-workflow")
  })

  it("returns undefined for unknown workflow", () => {
    expect(getWorkflow("nonexistent")).toBeUndefined()
  })

  it("lists registered workflows", () => {
    const workflow: WorkflowDefinition = {
      id: "list-workflow",
      name: "List Workflow",
      description: "List",
      steps: [{ id: "step1", title: "Step 1", workerId: "coder", prompt: "{task}" }],
    }

    registerWorkflow(workflow)

    const workflows = listWorkflows()
    expect(workflows.some((entry) => entry.id === "list-workflow")).toBe(true)
  })
})

describe("executeWorkflowStep", () => {
  const successDeps: WorkflowRunDependencies = {
    resolveWorker: async (id) => id,
    sendToWorker: async () => ({ success: true, response: "## Summary\nDone" }),
  }

  const failDeps: WorkflowRunDependencies = {
    resolveWorker: async (id) => id,
    sendToWorker: async () => ({ success: false, error: "boom" }),
  }

  const testWorkflow: WorkflowDefinition = {
    id: "test-wf",
    name: "Test",
    description: "Test",
    steps: [{ id: "s1", title: "Step 1", workerId: "coder", prompt: "{task}", carry: true }],
  }

  it("throws for invalid stepIndex", async () => {
    await expect(
      executeWorkflowStep(
        {
          runId: "run-1",
          workflow: testWorkflow,
          stepIndex: 99,
          task: "Test task",
          carry: "",
          autoSpawn: true,
          limits: { maxSteps: 4, maxTaskChars: 1000, maxCarryChars: 1000, perStepTimeoutMs: 5000 },
        },
        successDeps,
      ),
    ).rejects.toThrow(/Invalid stepIndex/)
  })

  it("throws for negative stepIndex", async () => {
    await expect(
      executeWorkflowStep(
        {
          runId: "run-1",
          workflow: testWorkflow,
          stepIndex: -1,
          task: "Test task",
          carry: "",
          autoSpawn: true,
          limits: { maxSteps: 4, maxTaskChars: 1000, maxCarryChars: 1000, perStepTimeoutMs: 5000 },
        },
        successDeps,
      ),
    ).rejects.toThrow(/Invalid stepIndex/)
  })

  it("rejects invalid attachment paths", async () => {
    await expect(
      executeWorkflowStep(
        {
          runId: "run-1",
          workflow: testWorkflow,
          stepIndex: 0,
          task: "Test task",
          carry: "",
          autoSpawn: true,
          attachments: [{ type: "file", path: "../secret.txt" }],
          limits: { maxSteps: 4, maxTaskChars: 1000, maxCarryChars: 1000, perStepTimeoutMs: 5000 },
        },
        successDeps,
      ),
    ).rejects.toThrow(/Invalid attachment path/)
  })

  it("executes valid step successfully", async () => {
    const result = await executeWorkflowStep(
      {
        runId: "run-1",
        workflow: testWorkflow,
        stepIndex: 0,
        task: "Test task",
        carry: "",
        autoSpawn: true,
        limits: { maxSteps: 4, maxTaskChars: 1000, maxCarryChars: 1000, perStepTimeoutMs: 5000 },
      },
      successDeps,
    )

    expect(result.step.status).toBe("success")
    expect(result.step.workerId).toBe("coder")
  })

  it("returns error status when dispatch fails", async () => {
    const result = await executeWorkflowStep(
      {
        runId: "run-1",
        workflow: testWorkflow,
        stepIndex: 0,
        task: "Test task",
        carry: "",
        autoSpawn: true,
        limits: { maxSteps: 4, maxTaskChars: 1000, maxCarryChars: 1000, perStepTimeoutMs: 5000 },
      },
      failDeps,
    )

    expect(result.step.status).toBe("error")
    expect(result.step.error).toBe("boom")
  })

  it("compacts carry content within limits", async () => {
    const longResponse = `## Summary\n${"a".repeat(2000)}`
    const longDeps: WorkflowRunDependencies = {
      resolveWorker: async (id) => id,
      sendToWorker: async () => ({ success: true, response: longResponse }),
    }

    const result = await executeWorkflowStep(
      {
        runId: "run-1",
        workflow: testWorkflow,
        stepIndex: 0,
        task: "Test task",
        carry: "",
        autoSpawn: true,
        limits: { maxSteps: 4, maxTaskChars: 1000, maxCarryChars: 200, perStepTimeoutMs: 5000 },
      },
      longDeps,
    )

    expect(result.carry.length).toBeLessThanOrEqual(200)
    expect(result.carry).toContain("Summary")
  })

  it("drops old carry blocks when over limit", async () => {
    const existingCarry = `### Old\n#### Summary\n${"a".repeat(300)}`
    const deps: WorkflowRunDependencies = {
      resolveWorker: async (id) => id,
      sendToWorker: async () => ({ success: true, response: "## Summary\nNew" }),
    }

    const result = await executeWorkflowStep(
      {
        runId: "run-1",
        workflow: testWorkflow,
        stepIndex: 0,
        task: "Test task",
        carry: existingCarry,
        autoSpawn: true,
        limits: { maxSteps: 4, maxTaskChars: 1000, maxCarryChars: 150, perStepTimeoutMs: 5000 },
      },
      deps,
    )

    expect(result.carry.length).toBeLessThanOrEqual(150)
    expect(result.carry).not.toContain("Old")
  })

  it("clamps step timeout to limits", async () => {
    let seenTimeout = 0
    const timeoutDeps: WorkflowRunDependencies = {
      resolveWorker: async (id) => id,
      sendToWorker: async (_workerId, _message, options) => {
        seenTimeout = options.timeoutMs
        return { success: true, response: "ok" }
      },
    }

    const timedWorkflow: WorkflowDefinition = {
      id: "timed",
      name: "Timed",
      description: "Timed",
      steps: [
        {
          id: "s1",
          title: "S1",
          workerId: "coder",
          prompt: "{task}",
          timeoutMs: 999999,
          carry: true,
        },
      ],
    }

    await executeWorkflowStep(
      {
        runId: "run-1",
        workflow: timedWorkflow,
        stepIndex: 0,
        task: "Test task",
        carry: "",
        autoSpawn: true,
        limits: { maxSteps: 4, maxTaskChars: 1000, maxCarryChars: 1000, perStepTimeoutMs: 5000 },
      },
      timeoutDeps,
    )

    expect(seenTimeout).toBe(5000)
  })
})

describe("runWorkflow", () => {
  const successDeps: WorkflowRunDependencies = {
    resolveWorker: async (id) => id,
    sendToWorker: async () => ({ success: true, response: "## Summary\nDone" }),
  }

  const failDeps: WorkflowRunDependencies = {
    resolveWorker: async (id) => id,
    sendToWorker: async () => ({ success: false, error: "boom" }),
  }

  it("throws for unknown workflow", async () => {
    await expect(
      runWorkflow(
        {
          workflowId: "nonexistent",
          task: "Test",
          limits: { maxSteps: 4, maxTaskChars: 1000, maxCarryChars: 1000, perStepTimeoutMs: 5000 },
        },
        successDeps,
      ),
    ).rejects.toThrow(/Unknown workflow/)
  })

  it("handles empty workflow", async () => {
    registerWorkflow({
      id: "empty-wf",
      name: "Empty",
      description: "Empty",
      steps: [],
    })

    const result = await runWorkflow(
      {
        workflowId: "empty-wf",
        task: "Test",
        limits: { maxSteps: 4, maxTaskChars: 1000, maxCarryChars: 1000, perStepTimeoutMs: 5000 },
      },
      successDeps,
    )

    expect(result.status).toBe("success")
    expect(result.steps.length).toBe(0)
    expect(result.lastStepResult).toBeUndefined()
  })

  it("runs multi-step workflow successfully", async () => {
    registerWorkflow({
      id: "ok-wf",
      name: "OK",
      description: "OK",
      steps: [
        { id: "s1", title: "S1", workerId: "coder", prompt: "{task}", carry: true },
        { id: "s2", title: "S2", workerId: "coder", prompt: "{carry}", carry: true },
      ],
    })

    const messages: string[] = []
    const deps: WorkflowRunDependencies = {
      resolveWorker: async (id) => id,
      sendToWorker: async (_id, message) => {
        messages.push(message)
        return { success: true, response: "## Summary\nDone" }
      },
    }

    const result = await runWorkflow(
      {
        workflowId: "ok-wf",
        task: "Test",
        limits: { maxSteps: 4, maxTaskChars: 1000, maxCarryChars: 500, perStepTimeoutMs: 5000 },
      },
      deps,
    )

    expect(result.status).toBe("success")
    expect(result.steps.length).toBe(2)
    expect(result.lastStepResult?.status).toBe("success")
    expect(messages[1]).toContain("Summary")
  })

  it("stops on failed step", async () => {
    registerWorkflow({
      id: "error-wf",
      name: "Error",
      description: "Error",
      steps: [{ id: "s1", title: "S1", workerId: "coder", prompt: "{task}" }],
    })

    const result = await runWorkflow(
      {
        workflowId: "error-wf",
        task: "Test",
        limits: { maxSteps: 4, maxTaskChars: 1000, maxCarryChars: 1000, perStepTimeoutMs: 5000 },
      },
      failDeps,
    )

    expect(result.status).toBe("error")
    expect(result.steps[0]?.status).toBe("error")
  })
})

describe("validateWorkflowInput", () => {
  const workflow: WorkflowDefinition = {
    id: "test",
    name: "Test",
    description: "Test",
    steps: [{ id: "s1", title: "S1", workerId: "coder", prompt: "{task}" }],
  }

  it("throws when task exceeds maxTaskChars", () => {
    const longTask = "a".repeat(1001)
    expect(() =>
      validateWorkflowInput(
        {
          workflowId: "test",
          task: longTask,
          limits: { maxSteps: 4, maxTaskChars: 1000, maxCarryChars: 1000, perStepTimeoutMs: 5000 },
        },
        workflow,
      ),
    ).toThrow(/maxTaskChars/)
  })

  it("throws when workflow exceeds maxSteps", () => {
    const tooManySteps: WorkflowDefinition = {
      id: "too-many",
      name: "Too many",
      description: "Too many",
      steps: [
        { id: "s1", title: "S1", workerId: "coder", prompt: "{task}" },
        { id: "s2", title: "S2", workerId: "coder", prompt: "{task}" },
      ],
    }

    expect(() =>
      validateWorkflowInput(
        {
          workflowId: "too-many",
          task: "Test",
          limits: { maxSteps: 1, maxTaskChars: 1000, maxCarryChars: 1000, perStepTimeoutMs: 5000 },
        },
        tooManySteps,
      ),
    ).toThrow(/maxSteps/)
  })
})
