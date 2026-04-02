package hugging_face_trackio

import "Lumaestro/internal/agents/skills"

func init() {
	skills.Register(skills.Skill{
		Name: "hugging-face-trackio",
		Category: "infrastructure",
		Content: `---
source: "https://github.com/huggingface/skills/tree/main/skills/huggingface-trackio"
name: hugging-face-trackio
description: Track ML experiments with Trackio using Python logging, alerts, and CLI metric retrieval.
risk: unknown
---

# Trackio - Experiment Tracking for ML Training

Trackio is an experiment tracking library for logging and visualizing ML training metrics. It syncs to Hugging Face Spaces for real-time monitoring dashboards.

## Three Interfaces

| Task | Interface | Reference |
|------|-----------|-----------|
| **Logging metrics** during training | Python API | [references/logging_metrics.md](references/logging_metrics.md) |
| **Firing alerts** for training diagnostics | Python API | [references/alerts.md](references/alerts.md) |
| **Retrieving metrics & alerts** after/during training | CLI | [references/retrieving_metrics.md](references/retrieving_metrics.md) |

## When to Use Each

### Python API → Logging

Use ` + "`" + `import trackio` + "`" + ` in your training scripts to log metrics:

- Initialize tracking with ` + "`" + `trackio.init()` + "`" + `
- Log metrics with ` + "`" + `trackio.log()` + "`" + ` or use TRL's ` + "`" + `report_to="trackio"` + "`" + `
- Finalize with ` + "`" + `trackio.finish()` + "`" + `

**Key concept**: For remote/cloud training, pass ` + "`" + `space_id` + "`" + ` — metrics sync to a Space dashboard so they persist after the instance terminates.

→ See [references/logging_metrics.md](references/logging_metrics.md) for setup, TRL integration, and configuration options.

### Python API → Alerts

Insert ` + "`" + `trackio.alert()` + "`" + ` calls in training code to flag important events — like inserting print statements for debugging, but structured and queryable:

- ` + "`" + `trackio.alert(title="...", level=trackio.AlertLevel.WARN)` + "`" + ` — fire an alert
- Three severity levels: ` + "`" + `INFO` + "`" + `, ` + "`" + `WARN` + "`" + `, ` + "`" + `ERROR` + "`" + `
- Alerts are printed to terminal, stored in the database, shown in the dashboard, and optionally sent to webhooks (Slack/Discord)

**Key concept for LLM agents**: Alerts are the primary mechanism for autonomous experiment iteration. An agent should insert alerts into training code for diagnostic conditions (loss spikes, NaN gradients, low accuracy, training stalls). Since alerts are printed to the terminal, an agent that is watching the training script's output will see them automatically. For background or detached runs, the agent can poll via CLI instead.

→ See [references/alerts.md](references/alerts.md) for the full alerts API, webhook setup, and autonomous agent workflows.

### CLI → Retrieving

Use the ` + "`" + `trackio` + "`" + ` command to query logged metrics and alerts:

- ` + "`" + `trackio list projects/runs/metrics` + "`" + ` — discover what's available
- ` + "`" + `trackio get project/run/metric` + "`" + ` — retrieve summaries and values
- ` + "`" + `trackio list alerts --project <name> --json` + "`" + ` — retrieve alerts
- ` + "`" + `trackio show` + "`" + ` — launch the dashboard
- ` + "`" + `trackio sync` + "`" + ` — sync to HF Space

**Key concept**: Add ` + "`" + `--json` + "`" + ` for programmatic output suitable for automation and LLM agents.

→ See [references/retrieving_metrics.md](references/retrieving_metrics.md) for all commands, workflows, and JSON output formats.

## Minimal Logging Setup

` + "`" + `` + "`" + `` + "`" + `python
import trackio

trackio.init(project="my-project", space_id="username/trackio")
trackio.log({"loss": 0.1, "accuracy": 0.9})
trackio.log({"loss": 0.09, "accuracy": 0.91})
trackio.finish()
` + "`" + `` + "`" + `` + "`" + `

### Minimal Retrieval

` + "`" + `` + "`" + `` + "`" + `bash
trackio list projects --json
trackio get metric --project my-project --run my-run --metric loss --json
` + "`" + `` + "`" + `` + "`" + `

## Autonomous ML Experiment Workflow

When running experiments autonomously as an LLM agent, the recommended workflow is:

1. **Set up training with alerts** — insert ` + "`" + `trackio.alert()` + "`" + ` calls for diagnostic conditions
2. **Launch training** — run the script in the background
3. **Poll for alerts** — use ` + "`" + `trackio list alerts --project <name> --json --since <timestamp>` + "`" + ` to check for new alerts
4. **Read metrics** — use ` + "`" + `trackio get metric ...` + "`" + ` to inspect specific values
5. **Iterate** — based on alerts and metrics, stop the run, adjust hyperparameters, and launch a new run

` + "`" + `` + "`" + `` + "`" + `python
import trackio

trackio.init(project="my-project", config={"lr": 1e-4})

for step in range(num_steps):
    loss = train_step()
    trackio.log({"loss": loss, "step": step})

    if step > 100 and loss > 5.0:
        trackio.alert(
            title="Loss divergence",
            text=f"Loss {loss:.4f} still high after {step} steps",
            level=trackio.AlertLevel.ERROR,
        )
    if step > 0 and abs(loss) < 1e-8:
        trackio.alert(
            title="Vanishing loss",
            text="Loss near zero — possible gradient collapse",
            level=trackio.AlertLevel.WARN,
        )

trackio.finish()
` + "`" + `` + "`" + `` + "`" + `

Then poll from a separate terminal/process:

` + "`" + `` + "`" + `` + "`" + `bash
trackio list alerts --project my-project --json --since "2025-01-01T00:00:00"
` + "`" + `` + "`" + `` + "`" + `
`,
	})
}
