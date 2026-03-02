import type { Plugin } from "@opencode-ai/plugin"

/**
 * Agent Team plugin for OpenCode.
 * Delegates all hook logic to the `agent-team` Go binary via Bun shell.
 *
 * Install: add "opencode-agent-team" to opencode.json plugin array.
 * Requires: `agent-team` binary in PATH.
 */
const plugin: Plugin = async ({ $, worktree }) => {
  const cwd = worktree || process.cwd()

  return {
    "session.created": async () => {
      try {
        const input = JSON.stringify({ cwd })
        await $`echo ${input} | agent-team hook session-start --provider opencode`
      } catch {
        // non-fatal: agent-team binary may not be installed
      }
    },

    "session.idle": async () => {
      try {
        const input = JSON.stringify({ cwd })
        await $`echo ${input} | agent-team hook teammate-idle --provider opencode`
      } catch {
        // non-fatal
      }
    },

    "tool.execute.before": async (toolInput) => {
      try {
        const input = JSON.stringify({
          cwd,
          tool_name: toolInput.tool,
          tool_input: toolInput.args,
        })
        await $`echo ${input} | agent-team hook pre-tool-use --provider opencode`
      } catch {
        // non-fatal: don't block tool execution
      }
    },

    "tool.execute.after": async (toolInput) => {
      try {
        const input = JSON.stringify({
          cwd,
          tool_name: toolInput.tool,
        })
        await $`echo ${input} | agent-team hook post-tool-use --provider opencode`
      } catch {
        // non-fatal
      }
    },
  }
}

export default plugin
