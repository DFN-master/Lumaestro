package create_pr

import "Lumaestro/internal/agents/skills"

func init() {
	skills.Register(skills.Skill{
		Name: "create-pr",
		Category: "workflow",
		Content: `---
name: create-pr
description: Alias for sentry-skills:pr-writer. Use when users explicitly ask for "create-pr" or reference the legacy skill name. Redirects to the canonical PR writing workflow.
risk: unknown
source: community
---

# Alias: create-pr

This skill name is kept for compatibility.

Use ` + "`" + `sentry-skills:pr-writer` + "`" + ` as the canonical skill for creating and editing pull requests.

If invoked via ` + "`" + `create-pr` + "`" + `, run the same workflow and conventions documented in ` + "`" + `sentry-skills:pr-writer` + "`" + `.
`,
	})
}
