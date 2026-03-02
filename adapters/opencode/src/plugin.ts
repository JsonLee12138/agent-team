import { promises as fs } from "node:fs"
import path from "node:path"
import type { Plugin } from "@opencode-ai/plugin"

/**
 * Agent Team plugin for OpenCode.
 * Delegates all hook logic to the `agent-team` Go binary via Bun shell.
 * Bridges skills from .agents/teams/ to .opencode/skills/ on session start.
 *
 * Install: add "opencode-agent-team" to opencode.json plugin array.
 * Requires: `agent-team` binary in PATH.
 */

/**
 * Bridge skills from .agents/teams/ (and ~/.agents/roles/) into .opencode/skills/
 * so OpenCode can discover them via its fixed skill scan paths.
 */
async function bridgeSkills(cwd: string): Promise<void> {
  const targetDir = path.join(cwd, ".opencode", "skills")
  await fs.mkdir(targetDir, { recursive: true })

  // Collect skill source directories: project-level + global
  const sourceDirs: Array<{ dir: string; scope: string }> = []

  // Project-level: .agents/teams/
  const projectTeams = path.join(cwd, ".agents", "teams")
  sourceDirs.push({ dir: projectTeams, scope: "project" })

  // Global: ~/.agents/roles/
  const home = process.env.HOME || process.env.USERPROFILE || ""
  if (home) {
    sourceDirs.push({ dir: path.join(home, ".agents", "roles"), scope: "global" })
  }

  for (const { dir } of sourceDirs) {
    let entries: Array<import("node:fs").Dirent>
    try {
      entries = await fs.readdir(dir, { withFileTypes: true })
    } catch {
      continue
    }

    for (const entry of entries) {
      if (!entry.isDirectory()) continue

      const src = path.join(dir, entry.name)
      // Only bridge directories that contain SKILL.md (valid skills)
      try {
        await fs.access(path.join(src, "SKILL.md"))
      } catch {
        continue
      }

      const dest = path.join(targetDir, entry.name)
      // Skip if already exists (don't overwrite)
      try {
        await fs.lstat(dest)
        continue
      } catch {
        // dest doesn't exist, create symlink
      }

      try {
        await fs.symlink(src, dest, "dir")
      } catch {
        // non-fatal: may fail on some platforms or permission issues
      }
    }
  }
}

const plugin: Plugin = async ({ $, worktree }) => {
  const cwd = worktree || process.cwd()

  // Verify agent-team binary is available on load
  try {
    await $`agent-team version`.quiet()
    console.error("[agent-team] plugin loaded")
  } catch {
    console.error("[agent-team] WARNING: agent-team binary not found in PATH")
  }

  return {
    "session.created": async () => {
      // Bridge skills from .agents/teams/ to .opencode/skills/
      try {
        await bridgeSkills(cwd)
      } catch {
        // non-fatal
      }

      // Run session-start hook (role injection, worktree detection)
      try {
        const input = JSON.stringify({ cwd })
        await $`echo ${input} | agent-team hook session-start --provider opencode`
      } catch {
        // non-fatal
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
        // non-fatal
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
