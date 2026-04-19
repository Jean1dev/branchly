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

function baseLayout(title: string, body: string): string {
  return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>${title}</title>
</head>
<body style="margin:0;padding:0;background:#f4f4f5;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;">
  <table width="100%" cellpadding="0" cellspacing="0" style="background:#f4f4f5;padding:32px 16px;">
    <tr><td align="center">
      <table width="600" cellpadding="0" cellspacing="0" style="background:#ffffff;border-radius:8px;border:1px solid #e4e4e7;overflow:hidden;max-width:600px;width:100%;">
        <tr>
          <td style="padding:24px 32px;background:#18181b;border-bottom:1px solid #27272a;">
            <span style="font-size:20px;font-weight:700;color:#ffffff;letter-spacing:-0.5px;">Branchly</span>
          </td>
        </tr>
        <tr><td style="padding:32px;">${body}</td></tr>
        <tr>
          <td style="padding:16px 32px;background:#f4f4f5;border-top:1px solid #e4e4e7;font-size:12px;color:#71717a;">
            You received this email because you have email notifications enabled in Branchly settings.
          </td>
        </tr>
      </table>
    </td></tr>
  </table>
</body>
</html>`
}

function formatDuration(seconds: number): string {
  if (seconds < 60) return `${seconds}s`
  const m = Math.floor(seconds / 60)
  const s = seconds % 60
  return s > 0 ? `${m}m ${s}s` : `${m}m`
}

function metaRow(label: string, value: string): string {
  return `<tr>
    <td style="padding:6px 0;font-size:13px;color:#71717a;width:140px;vertical-align:top;">${label}</td>
    <td style="padding:6px 0;font-size:13px;color:#18181b;word-break:break-word;">${value}</td>
  </tr>`
}

export function jobCompletedTemplate(data: JobEmailData): string {
  const costLine = data.estimatedCostUsd != null
    ? metaRow('Estimated cost', `$${data.estimatedCostUsd.toFixed(4)}`)
    : ''
  const prLine = data.prUrl
    ? `<p style="margin:24px 0 0;"><a href="${data.prUrl}" style="display:inline-block;padding:10px 20px;background:#18181b;color:#ffffff;text-decoration:none;border-radius:6px;font-size:14px;font-weight:500;">View Pull Request →</a></p>`
    : ''

  const body = `
    <p style="margin:0 0 8px;font-size:22px;font-weight:700;color:#18181b;">Job completed</p>
    <p style="margin:0 0 24px;font-size:14px;color:#71717a;">Your job finished successfully on <strong style="color:#18181b;">${data.repoFullName}</strong>.</p>
    <table cellpadding="0" cellspacing="0" width="100%" style="border:1px solid #e4e4e7;border-radius:6px;padding:16px;margin-bottom:24px;">
      ${metaRow('Repository', data.repoFullName)}
      ${metaRow('Branch', data.branchName)}
      ${metaRow('Agent', data.agentName)}
      ${metaRow('Duration', formatDuration(data.durationSeconds))}
      ${costLine}
      ${metaRow('Finished', data.finishedAt.toUTCString())}
    </table>
    <p style="margin:0 0 8px;font-size:13px;font-weight:600;color:#18181b;">Prompt</p>
    <p style="margin:0 0 24px;padding:12px 16px;background:#f4f4f5;border-radius:6px;font-size:13px;color:#3f3f46;line-height:1.5;">${data.prompt}</p>
    <p style="margin:0;"><a href="${data.jobLogsUrl}" style="font-size:13px;color:#6366f1;text-decoration:none;">View job logs →</a></p>
    ${prLine}
  `
  return baseLayout('Job completed — Branchly', body)
}

export function jobFailedTemplate(data: JobEmailData): string {
  const errorBlock = data.errorMessage
    ? `<p style="margin:0 0 8px;font-size:13px;font-weight:600;color:#18181b;">Error</p>
       <p style="margin:0 0 24px;padding:12px 16px;background:#fef2f2;border:1px solid #fecaca;border-radius:6px;font-size:13px;color:#b91c1c;font-family:monospace;">${data.errorMessage}</p>`
    : ''

  const body = `
    <p style="margin:0 0 8px;font-size:22px;font-weight:700;color:#18181b;">Job failed</p>
    <p style="margin:0 0 24px;font-size:14px;color:#71717a;">Your job on <strong style="color:#18181b;">${data.repoFullName}</strong> did not complete successfully.</p>
    <table cellpadding="0" cellspacing="0" width="100%" style="border:1px solid #e4e4e7;border-radius:6px;padding:16px;margin-bottom:24px;">
      ${metaRow('Repository', data.repoFullName)}
      ${metaRow('Branch', data.branchName)}
      ${metaRow('Agent', data.agentName)}
      ${metaRow('Duration', formatDuration(data.durationSeconds))}
      ${metaRow('Failed at', data.finishedAt.toUTCString())}
    </table>
    <p style="margin:0 0 8px;font-size:13px;font-weight:600;color:#18181b;">Prompt</p>
    <p style="margin:0 0 24px;padding:12px 16px;background:#f4f4f5;border-radius:6px;font-size:13px;color:#3f3f46;line-height:1.5;">${data.prompt}</p>
    ${errorBlock}
    <p style="margin:0;"><a href="${data.jobLogsUrl}" style="font-size:13px;color:#6366f1;text-decoration:none;">View job logs →</a></p>
  `
  return baseLayout('Job failed — Branchly', body)
}

export function prOpenedTemplate(data: JobEmailData): string {
  const prButton = data.prUrl
    ? `<p style="margin:24px 0 0;"><a href="${data.prUrl}" style="display:inline-block;padding:10px 20px;background:#18181b;color:#ffffff;text-decoration:none;border-radius:6px;font-size:14px;font-weight:500;">View Pull Request →</a></p>`
    : ''

  const body = `
    <p style="margin:0 0 8px;font-size:22px;font-weight:700;color:#18181b;">Pull request opened</p>
    <p style="margin:0 0 24px;font-size:14px;color:#71717a;">A pull request was opened on <strong style="color:#18181b;">${data.repoFullName}</strong>.</p>
    <table cellpadding="0" cellspacing="0" width="100%" style="border:1px solid #e4e4e7;border-radius:6px;padding:16px;margin-bottom:24px;">
      ${metaRow('Repository', data.repoFullName)}
      ${metaRow('Branch', data.branchName)}
      ${metaRow('Agent', data.agentName)}
      ${metaRow('Duration', formatDuration(data.durationSeconds))}
      ${metaRow('Opened at', data.finishedAt.toUTCString())}
    </table>
    <p style="margin:0 0 8px;font-size:13px;font-weight:600;color:#18181b;">Prompt</p>
    <p style="margin:0 0 24px;padding:12px 16px;background:#f4f4f5;border-radius:6px;font-size:13px;color:#3f3f46;line-height:1.5;">${data.prompt}</p>
    <p style="margin:0;"><a href="${data.jobLogsUrl}" style="font-size:13px;color:#6366f1;text-decoration:none;">View job logs →</a></p>
    ${prButton}
  `
  return baseLayout('Pull request opened — Branchly', body)
}
