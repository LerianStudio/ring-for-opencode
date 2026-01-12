/**
 * Ring OpenCode Plugin
 *
 * This module exports ONLY the plugin function for OpenCode.
 *
 * IMPORTANT: OpenCode's plugin loader iterates over ALL exports and calls
 * each one as a function. Any non-function export will crash the loader with:
 * "TypeError: fn3 is not a function"
 *
 * Therefore, this file MUST only export:
 * - RingUnifiedPlugin (named export)
 * - default export
 *
 * For internal APIs (hooks, config, utils, etc.), import directly from
 * the submodules:
 * - "./hooks/index.js"
 * - "./config/index.js"
 * - "./utils/state.js"
 * - "./loaders/index.js"
 * - "./tools/index.js"
 * - "./lifecycle/index.js"
 */

// =============================================================================
// PLUGIN EXPORTS ONLY
// =============================================================================

export { RingUnifiedPlugin, RingUnifiedPlugin as default } from "./ring-unified.js"
