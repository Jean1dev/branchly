import { describe, it, expect } from 'vitest'
import {
  TEMPLATE_SLUGS,
  JOB_COMPLETED_HTML_TEMPLATE,
  JOB_FAILED_HTML_TEMPLATE,
  PR_OPENED_HTML_TEMPLATE,
  jobCompletedVars,
  jobFailedVars,
  prOpenedVars,
  type JobEmailData,
} from './templates'

const baseData: JobEmailData = {
  userEmail: 'user@example.com',
  userName: 'Test User',
  repoFullName: 'owner/repo',
  branchName: 'branchly/add-feature',
  agentName: 'Claude Code',
  prompt: 'Add a comprehensive test suite',
  durationSeconds: 90,
  jobLogsUrl: 'https://app.branchly.com/jobs/job-123',
  finishedAt: new Date('2026-04-20T12:00:00Z'),
}

// ---- Template slug constants ----

describe('TEMPLATE_SLUGS', () => {
  it('defines unique slugs for each event type', () => {
    const slugs = Object.values(TEMPLATE_SLUGS)
    expect(new Set(slugs).size).toBe(slugs.length)
  })

  it('all slugs start with "branchly-"', () => {
    for (const slug of Object.values(TEMPLATE_SLUGS)) {
      expect(slug).toMatch(/^branchly-/)
    }
  })
})

// ---- Go template HTML strings ----

describe('JOB_COMPLETED_HTML_TEMPLATE', () => {
  it('contains HTML doctype', () => {
    expect(JOB_COMPLETED_HTML_TEMPLATE).toContain('<!DOCTYPE html>')
  })

  it('contains Go template variable placeholders', () => {
    expect(JOB_COMPLETED_HTML_TEMPLATE).toContain('{{.repo_full_name}}')
    expect(JOB_COMPLETED_HTML_TEMPLATE).toContain('{{.branch_name}}')
    expect(JOB_COMPLETED_HTML_TEMPLATE).toContain('{{.agent_name}}')
    expect(JOB_COMPLETED_HTML_TEMPLATE).toContain('{{.prompt}}')
    expect(JOB_COMPLETED_HTML_TEMPLATE).toContain('{{.duration}}')
    expect(JOB_COMPLETED_HTML_TEMPLATE).toContain('{{.job_logs_url}}')
  })

  it('uses conditional block for PR URL', () => {
    expect(JOB_COMPLETED_HTML_TEMPLATE).toContain('{{if .pr_url}}')
    expect(JOB_COMPLETED_HTML_TEMPLATE).toContain('{{.pr_url}}')
  })

  it('uses conditional block for estimated cost', () => {
    expect(JOB_COMPLETED_HTML_TEMPLATE).toContain('{{if .estimated_cost}}')
    expect(JOB_COMPLETED_HTML_TEMPLATE).toContain('{{.estimated_cost}}')
  })
})

describe('JOB_FAILED_HTML_TEMPLATE', () => {
  it('contains HTML doctype', () => {
    expect(JOB_FAILED_HTML_TEMPLATE).toContain('<!DOCTYPE html>')
  })

  it('contains Go template variable for error message', () => {
    expect(JOB_FAILED_HTML_TEMPLATE).toContain('{{if .error_message}}')
    expect(JOB_FAILED_HTML_TEMPLATE).toContain('{{.error_message}}')
  })

  it('contains required variable placeholders', () => {
    expect(JOB_FAILED_HTML_TEMPLATE).toContain('{{.repo_full_name}}')
    expect(JOB_FAILED_HTML_TEMPLATE).toContain('{{.prompt}}')
    expect(JOB_FAILED_HTML_TEMPLATE).toContain('{{.job_logs_url}}')
  })
})

describe('PR_OPENED_HTML_TEMPLATE', () => {
  it('contains HTML doctype', () => {
    expect(PR_OPENED_HTML_TEMPLATE).toContain('<!DOCTYPE html>')
  })

  it('contains PR URL conditional block', () => {
    expect(PR_OPENED_HTML_TEMPLATE).toContain('{{if .pr_url}}')
  })

  it('contains required variable placeholders', () => {
    expect(PR_OPENED_HTML_TEMPLATE).toContain('{{.repo_full_name}}')
    expect(PR_OPENED_HTML_TEMPLATE).toContain('{{.branch_name}}')
  })
})

// ---- Variable extractors ----

describe('jobCompletedVars', () => {
  it('extracts all required keys', () => {
    const vars = jobCompletedVars(baseData)
    expect(vars.repo_full_name).toBe('owner/repo')
    expect(vars.branch_name).toBe('branchly/add-feature')
    expect(vars.agent_name).toBe('Claude Code')
    expect(vars.prompt).toBe('Add a comprehensive test suite')
    expect(vars.job_logs_url).toBe('https://app.branchly.com/jobs/job-123')
  })

  it('formats duration in seconds', () => {
    const vars = jobCompletedVars({ ...baseData, durationSeconds: 45 })
    expect(vars.duration).toBe('45s')
  })

  it('formats duration in minutes and seconds', () => {
    const vars = jobCompletedVars({ ...baseData, durationSeconds: 90 })
    expect(vars.duration).toBe('1m 30s')
  })

  it('formats whole minutes correctly', () => {
    const vars = jobCompletedVars({ ...baseData, durationSeconds: 120 })
    expect(vars.duration).toBe('2m')
  })

  it('includes formatted cost when estimatedCostUsd is provided', () => {
    const vars = jobCompletedVars({ ...baseData, estimatedCostUsd: 0.0042 })
    expect(vars.estimated_cost).toBe('$0.0042')
  })

  it('returns empty string for estimated_cost when not provided', () => {
    const vars = jobCompletedVars({ ...baseData, estimatedCostUsd: undefined })
    expect(vars.estimated_cost).toBe('')
  })

  it('returns empty string for pr_url when not provided', () => {
    const vars = jobCompletedVars({ ...baseData, prUrl: undefined })
    expect(vars.pr_url).toBe('')
  })

  it('includes pr_url when provided', () => {
    const vars = jobCompletedVars({ ...baseData, prUrl: 'https://github.com/owner/repo/pull/1' })
    expect(vars.pr_url).toBe('https://github.com/owner/repo/pull/1')
  })
})

describe('jobFailedVars', () => {
  it('includes error_message when provided', () => {
    const vars = jobFailedVars({ ...baseData, errorMessage: 'agent timed out' })
    expect(vars.error_message).toBe('agent timed out')
  })

  it('returns empty string for error_message when not provided', () => {
    const vars = jobFailedVars({ ...baseData, errorMessage: undefined })
    expect(vars.error_message).toBe('')
  })

  it('includes base variables', () => {
    const vars = jobFailedVars(baseData)
    expect(vars.repo_full_name).toBe('owner/repo')
    expect(vars.agent_name).toBe('Claude Code')
  })
})

describe('prOpenedVars', () => {
  it('extracts all required keys', () => {
    const vars = prOpenedVars({ ...baseData, prUrl: 'https://github.com/owner/repo/pull/7' })
    expect(vars.pr_url).toBe('https://github.com/owner/repo/pull/7')
    expect(vars.repo_full_name).toBe('owner/repo')
    expect(vars.branch_name).toBe('branchly/add-feature')
  })
})
