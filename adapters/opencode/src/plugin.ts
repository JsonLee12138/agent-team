import type { Plugin } from "@opencode-ai/plugin"

/**
 * Agent Team plugin for OpenCode.
 * Delegates all hook logic to the `agent-team` Go binary via Bun shell.
 *
 * Install: add "opencode-agent-team" to opencode.json plugin array.
 * Requires: `agent-team` binary in PATH.
 */
const plugin: Plugin = async ({ $ }) => {
  const cwd = process.cwd()

  // Verify agent-team binary is available on load
  try {
    await $`agent-team version`.quiet()
    console.error("[agent-team] plugin loaded")
  } catch {
    console.error("[agent-team] WARNING: agent-team binary not found in PATH")
  }

  return {
    "tool.execute.before": async (input, output) => {
      try {
        const payload = JSON.stringify({
          cwd,
          tool_name: input.tool,
          tool_input: output.args,
        })
        await $`echo ${payload} | agent-team hook pre-tool-use --provider opencode`
      } catch {
        // non-fatal
      }
    },

    "tool.execute.after": async (input) => {
      try {
        const payload = JSON.stringify({
          cwd,
          tool_name: input.tool,
        })
        await $`echo ${payload} | agent-team hook post-tool-use --provider opencode`
      } catch {
        // non-fatal
      }
    },
  }
}

export default plugin
