export interface JobEmailData {
  userEmail: string
  userName: string
  repoFullName: string
  branchName: string
  agentName: string
  prompt: string
  durationSeconds: number
  estimatedCostUsd?: number
  prUrl?: string
  jobLogsUrl: string
  errorMessage?: string
  finishedAt: Date
}

// Slugs used to register and reference templates in the email API.
export const TEMPLATE_SLUGS = {
  JOB_COMPLETED: 'branchly-job-completed',
  JOB_FAILED: 'branchly-job-failed',
  PR_OPENED: 'branchly-pr-opened',
} as const

// ---- Go template HTML strings (registered once in the email API) ----
// Variables use Go html/template syntax: {{.var_name}}
// Conditional blocks: {{if .var_name}}...{{end}}

const BASE_HEADER = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
</head>
<body style="margin:0;padding:0;background:#f4f4f5;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;">
<table width="100%" cellpadding="0" cellspacing="0" style="background:#f4f4f5;padding:32px 16px;">
<tr><td align="center">
<table width="600" cellpadding="0" cellspacing="0" style="background:#ffffff;border-radius:8px;border:1px solid #e4e4e7;overflow:hidden;max-width:600px;width:100%;">
<tr><td style="padding:24px 32px;background:#18181b;border-bottom:1px solid #27272a;">
  <span style="font-size:20px;font-weight:700;color:#ffffff;letter-spacing:-0.5px;">Branchly</span>
</td></tr>
<tr><td style="padding:32px;">`

const BASE_FOOTER = `</td></tr>
<tr><td style="padding:16px 32px;background:#f4f4f5;border-top:1px solid #e4e4e7;font-size:12px;color:#71717a;">
  You received this email because you have email notifications enabled in Branchly settings.
</td></tr>
</table>
</td></tr>
</table>
</body>
</html>`

function metaTable(rows: string): string {
  return `<table cellpadding="0" cellspacing="0" width="100%" style="border:1px solid #e4e4e7;border-radius:6px;padding:16px;margin-bottom:24px;">${rows}</table>`
}

function metaRow(label: string, value: string): string {
  return `<tr>
    <td style="padding:6px 0;font-size:13px;color:#71717a;width:140px;vertical-align:top;">${label}</td>
    <td style="padding:6px 0;font-size:13px;color:#18181b;word-break:break-word;">${value}</td>
  </tr>`
}

export const JOB_COMPLETED_HTML_TEMPLATE =
  BASE_HEADER +
  `<p style="margin:0 0 8px;font-size:22px;font-weight:700;color:#18181b;">Job completed</p>
<p style="margin:0 0 24px;font-size:14px;color:#71717a;">Your job finished successfully on <strong style="color:#18181b;">{{.repo_full_name}}</strong>.</p>` +
  metaTable(
    metaRow('Repository', '{{.repo_full_name}}') +
    metaRow('Branch', '{{.branch_name}}') +
    metaRow('Agent', '{{.agent_name}}') +
    metaRow('Duration', '{{.duration}}') +
    `{{if .estimated_cost}}` + metaRow('Estimated cost', '{{.estimated_cost}}') + `{{end}}` +
    metaRow('Finished', '{{.finished_at}}')
  ) +
  `<p style="margin:0 0 8px;font-size:13px;font-weight:600;color:#18181b;">Prompt</p>
<p style="margin:0 0 24px;padding:12px 16px;background:#f4f4f5;border-radius:6px;font-size:13px;color:#3f3f46;line-height:1.5;">{{.prompt}}</p>
<p style="margin:0;"><a href="{{.job_logs_url}}" style="font-size:13px;color:#6366f1;text-decoration:none;">View job logs →</a></p>
{{if .pr_url}}
<p style="margin:24px 0 0;"><a href="{{.pr_url}}" style="display:inline-block;padding:10px 20px;background:#18181b;color:#ffffff;text-decoration:none;border-radius:6px;font-size:14px;font-weight:500;">View Pull Request →</a></p>
{{end}}` +
  BASE_FOOTER

export const JOB_FAILED_HTML_TEMPLATE =
  BASE_HEADER +
  `<p style="margin:0 0 8px;font-size:22px;font-weight:700;color:#18181b;">Job failed</p>
<p style="margin:0 0 24px;font-size:14px;color:#71717a;">Your job on <strong style="color:#18181b;">{{.repo_full_name}}</strong> did not complete successfully.</p>` +
  metaTable(
    metaRow('Repository', '{{.repo_full_name}}') +
    metaRow('Branch', '{{.branch_name}}') +
    metaRow('Agent', '{{.agent_name}}') +
    metaRow('Duration', '{{.duration}}') +
    metaRow('Failed at', '{{.finished_at}}')
  ) +
  `<p style="margin:0 0 8px;font-size:13px;font-weight:600;color:#18181b;">Prompt</p>
<p style="margin:0 0 24px;padding:12px 16px;background:#f4f4f5;border-radius:6px;font-size:13px;color:#3f3f46;line-height:1.5;">{{.prompt}}</p>
{{if .error_message}}
<p style="margin:0 0 8px;font-size:13px;font-weight:600;color:#18181b;">Error</p>
<p style="margin:0 0 24px;padding:12px 16px;background:#fef2f2;border:1px solid #fecaca;border-radius:6px;font-size:13px;color:#b91c1c;font-family:monospace;">{{.error_message}}</p>
{{end}}
<p style="margin:0;"><a href="{{.job_logs_url}}" style="font-size:13px;color:#6366f1;text-decoration:none;">View job logs →</a></p>` +
  BASE_FOOTER

export const PR_OPENED_HTML_TEMPLATE =
  BASE_HEADER +
  `<p style="margin:0 0 8px;font-size:22px;font-weight:700;color:#18181b;">Pull request opened</p>
<p style="margin:0 0 24px;font-size:14px;color:#71717a;">A pull request was opened on <strong style="color:#18181b;">{{.repo_full_name}}</strong>.</p>` +
  metaTable(
    metaRow('Repository', '{{.repo_full_name}}') +
    metaRow('Branch', '{{.branch_name}}') +
    metaRow('Agent', '{{.agent_name}}') +
    metaRow('Duration', '{{.duration}}') +
    metaRow('Opened at', '{{.finished_at}}')
  ) +
  `<p style="margin:0 0 8px;font-size:13px;font-weight:600;color:#18181b;">Prompt</p>
<p style="margin:0 0 24px;padding:12px 16px;background:#f4f4f5;border-radius:6px;font-size:13px;color:#3f3f46;line-height:1.5;">{{.prompt}}</p>
<p style="margin:0;"><a href="{{.job_logs_url}}" style="font-size:13px;color:#6366f1;text-decoration:none;">View job logs →</a></p>
{{if .pr_url}}
<p style="margin:24px 0 0;"><a href="{{.pr_url}}" style="display:inline-block;padding:10px 20px;background:#18181b;color:#ffffff;text-decoration:none;border-radius:6px;font-size:14px;font-weight:500;">View Pull Request →</a></p>
{{end}}` +
  BASE_FOOTER

// ---- Variable extractors ----

function formatDuration(seconds: number): string {
  if (seconds < 60) return `${seconds}s`
  const m = Math.floor(seconds / 60)
  const s = seconds % 60
  return s > 0 ? `${m}m ${s}s` : `${m}m`
}

function baseVars(data: JobEmailData): Record<string, string> {
  return {
    user_name: data.userName,
    repo_full_name: data.repoFullName,
    branch_name: data.branchName,
    agent_name: data.agentName,
    prompt: data.prompt,
    duration: formatDuration(data.durationSeconds),
    estimated_cost: data.estimatedCostUsd != null ? `$${data.estimatedCostUsd.toFixed(4)}` : '',
    pr_url: data.prUrl ?? '',
    job_logs_url: data.jobLogsUrl,
    finished_at: data.finishedAt.toUTCString(),
  }
}

export function jobCompletedVars(data: JobEmailData): Record<string, string> {
  return baseVars(data)
}

export function jobFailedVars(data: JobEmailData): Record<string, string> {
  return { ...baseVars(data), error_message: data.errorMessage ?? '' }
}

export function prOpenedVars(data: JobEmailData): Record<string, string> {
  return baseVars(data)
}
