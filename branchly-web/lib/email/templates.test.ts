import { describe, it, expect } from 'vitest'
import {
  jobCompletedTemplate,
  jobFailedTemplate,
  prOpenedTemplate,
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

describe('jobCompletedTemplate', () => {
  it('returns a string containing HTML doctype', () => {
    const html = jobCompletedTemplate(baseData)
    expect(html).toContain('<!DOCTYPE html>')
  })

  it('includes the repository name', () => {
    const html = jobCompletedTemplate(baseData)
    expect(html).toContain('owner/repo')
  })

  it('includes the branch name', () => {
    const html = jobCompletedTemplate(baseData)
    expect(html).toContain('branchly/add-feature')
  })

  it('includes the agent name', () => {
    const html = jobCompletedTemplate(baseData)
    expect(html).toContain('Claude Code')
  })

  it('includes the prompt text', () => {
    const html = jobCompletedTemplate(baseData)
    expect(html).toContain('Add a comprehensive test suite')
  })

  it('includes the job logs URL', () => {
    const html = jobCompletedTemplate(baseData)
    expect(html).toContain('https://app.branchly.com/jobs/job-123')
  })

  it('formats duration in seconds correctly', () => {
    const html = jobCompletedTemplate({ ...baseData, durationSeconds: 45 })
    expect(html).toContain('45s')
  })

  it('formats duration in minutes when >= 60s', () => {
    const html = jobCompletedTemplate({ ...baseData, durationSeconds: 90 })
    expect(html).toContain('1m 30s')
  })

  it('includes PR button when prUrl is provided', () => {
    const html = jobCompletedTemplate({
      ...baseData,
      prUrl: 'https://github.com/owner/repo/pull/42',
    })
    expect(html).toContain('https://github.com/owner/repo/pull/42')
    expect(html).toContain('View Pull Request')
  })

  it('omits PR button when prUrl is not provided', () => {
    const html = jobCompletedTemplate({ ...baseData, prUrl: undefined })
    expect(html).not.toContain('View Pull Request')
  })

  it('includes estimated cost when provided', () => {
    const html = jobCompletedTemplate({
      ...baseData,
      estimatedCostUsd: 0.0042,
    })
    expect(html).toContain('$0.0042')
  })

  it('omits cost line when estimatedCostUsd is not provided', () => {
    const html = jobCompletedTemplate({ ...baseData, estimatedCostUsd: undefined })
    expect(html).not.toContain('Estimated cost')
  })
})

describe('jobFailedTemplate', () => {
  it('returns a string containing HTML doctype', () => {
    const html = jobFailedTemplate(baseData)
    expect(html).toContain('<!DOCTYPE html>')
  })

  it('includes "Job failed" heading', () => {
    const html = jobFailedTemplate(baseData)
    expect(html).toContain('Job failed')
  })

  it('includes the repository name', () => {
    const html = jobFailedTemplate(baseData)
    expect(html).toContain('owner/repo')
  })

  it('includes the error message when provided', () => {
    const html = jobFailedTemplate({
      ...baseData,
      errorMessage: 'agent process exited with code 1',
    })
    expect(html).toContain('agent process exited with code 1')
  })

  it('omits error block when errorMessage is not provided', () => {
    const html = jobFailedTemplate({ ...baseData, errorMessage: undefined })
    expect(html).not.toContain('Error')
  })

  it('includes the job logs URL', () => {
    const html = jobFailedTemplate(baseData)
    expect(html).toContain(baseData.jobLogsUrl)
  })
})

describe('prOpenedTemplate', () => {
  it('returns a string containing HTML doctype', () => {
    const html = prOpenedTemplate(baseData)
    expect(html).toContain('<!DOCTYPE html>')
  })

  it('includes "Pull request opened" heading', () => {
    const html = prOpenedTemplate(baseData)
    expect(html).toContain('Pull request opened')
  })

  it('includes the PR URL when provided', () => {
    const html = prOpenedTemplate({
      ...baseData,
      prUrl: 'https://github.com/owner/repo/pull/7',
    })
    expect(html).toContain('https://github.com/owner/repo/pull/7')
  })

  it('includes the repository and branch names', () => {
    const html = prOpenedTemplate(baseData)
    expect(html).toContain('owner/repo')
    expect(html).toContain('branchly/add-feature')
  })
})
