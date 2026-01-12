/**
 * Orchestrator Module
 *
 * Unified exports for the orchestrator subsystem.
 */

export {
  createOrchestratorContext,
  getWorkflowLimits,
  isInsideWorker,
  type LoadedOrchestratorConfig,
  loadOrchestratorConfig,
  type OrchestratorConfigFile,
} from "./config.js"

export { JobRegistry, jobRegistry } from "./jobs.js"

export {
  builtInProfiles,
  getProfile,
  listProfileIds,
  mergeProfile,
  resolveModelTag,
} from "./profiles.js"

export { createTaskTools } from "./tools/task-tools.js"

export type {
  DeviceRegistryEntry,
  DeviceRegistryFile,
  DeviceRegistrySessionEntry,
  DeviceRegistryWorkerEntry,
  Job,
  JobReport,
  JobStatus,
  OrchestratorContext,
  TaskDispatchInput,
  TaskDispatchResult,
  WorkerExecution,
  WorkerInstance,
  WorkerKind,
  WorkerPoolCallback,
  WorkerPoolEvent,
  WorkerProfile,
  WorkerSpawnOptions,
  WorkerStatus,
  WorkflowDefinition,
  WorkflowRunInput,
  WorkflowRunLimits,
  WorkflowRunResult,
  WorkflowStepDefinition,
  WorkflowStepResult,
} from "./types.js"

export { type SpawnOptions, WorkerPool, workerPool } from "./worker-pool.js"

export {
  executeWorkflowStep,
  getWorkflow,
  listWorkflows,
  registerWorkflow,
  runWorkflow,
  validateWorkflowInput,
  type WorkflowRunDependencies,
} from "./workflow/engine.js"
