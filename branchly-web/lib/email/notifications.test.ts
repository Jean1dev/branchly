import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { TEMPLATE_SLUGS, type JobEmailData } from './templates'

const mockSend = vi.fn().mockResolvedValue(undefined)

vi.mock('./index', () => ({
  getEmailProvider: () => ({ send: mockSend }),
}))

const { notifyJobCompleted, notifyJobFailed, notifyPROpened } = await import('./notifications')

const baseEmailData: JobEmailData = {
  userEmail: 'user@example.com',
  userName: 'Test User',
  repoFullName: 'owner/repo',
  branchName: 'branchly/add-feature',
  agentName: 'Claude Code',
  prompt: 'add tests',
  durationSeconds: 30,
  jobLogsUrl: 'https://app.branchly.com/jobs/job-1',
  finishedAt: new Date(),
}

function mockFetchWithPrefs(prefs: {
  enabled: boolean
  on_job_completed: boolean
  on_job_failed: boolean
  on_pr_opened: boolean
}) {
  global.fetch = vi.fn().mockResolvedValue({
    ok: true,
    json: async () => ({
      data: {
        email: 'user@example.com',
        name: 'Test User',
        notification_preferences: { email: prefs },
      },
    }),
  })
}

function mockFetchFailure() {
  global.fetch = vi.fn().mockRejectedValue(new Error('network error'))
}

function mockFetchNotOk() {
  global.fetch = vi.fn().mockResolvedValue({ ok: false })
}

const allEnabled = { enabled: true, on_job_completed: true, on_job_failed: true, on_pr_opened: true }

beforeEach(() => {
  mockSend.mockClear()
  process.env.API_URL = 'http://api:8080'
  process.env.INTERNAL_API_SECRET = 'test-secret'
})

afterEach(() => {
  vi.restoreAllMocks()
})

// ---- notifyJobCompleted ----

describe('notifyJobCompleted', () => {
  it('sends with correct template slug and recipient when enabled', async () => {
    mockFetchWithPrefs(allEnabled)
    await notifyJobCompleted('user-1', baseEmailData)

    expect(mockSend).toHaveBeenCalledOnce()
    const opts = mockSend.mock.calls[0][0] as { to: string; templateSlug: string; variables: Record<string, string> }
    expect(opts.to).toBe('user@example.com')
    expect(opts.templateSlug).toBe(TEMPLATE_SLUGS.JOB_COMPLETED)
  })

  it('passes repo_full_name and agent_name as variables', async () => {
    mockFetchWithPrefs(allEnabled)
    await notifyJobCompleted('user-1', baseEmailData)

    const opts = mockSend.mock.calls[0][0] as { variables: Record<string, string> }
    expect(opts.variables.repo_full_name).toBe('owner/repo')
    expect(opts.variables.agent_name).toBe('Claude Code')
  })

  it('does not send when master enabled=false', async () => {
    mockFetchWithPrefs({ ...allEnabled, enabled: false })
    await notifyJobCompleted('user-1', baseEmailData)
    expect(mockSend).not.toHaveBeenCalled()
  })

  it('does not send when on_job_completed=false', async () => {
    mockFetchWithPrefs({ ...allEnabled, on_job_completed: false })
    await notifyJobCompleted('user-1', baseEmailData)
    expect(mockSend).not.toHaveBeenCalled()
  })

  it('does not throw when fetch fails', async () => {
    mockFetchFailure()
    await expect(notifyJobCompleted('user-1', baseEmailData)).resolves.toBeUndefined()
    expect(mockSend).not.toHaveBeenCalled()
  })

  it('does not send when fetch returns non-ok', async () => {
    mockFetchNotOk()
    await notifyJobCompleted('user-1', baseEmailData)
    expect(mockSend).not.toHaveBeenCalled()
  })
})

// ---- notifyJobFailed ----

describe('notifyJobFailed', () => {
  it('sends with job_failed template slug', async () => {
    mockFetchWithPrefs(allEnabled)
    await notifyJobFailed('user-1', { ...baseEmailData, errorMessage: 'timeout' })

    expect(mockSend).toHaveBeenCalledOnce()
    const opts = mockSend.mock.calls[0][0] as { templateSlug: string; variables: Record<string, string> }
    expect(opts.templateSlug).toBe(TEMPLATE_SLUGS.JOB_FAILED)
    expect(opts.variables.error_message).toBe('timeout')
  })

  it('does not send when on_job_failed=false', async () => {
    mockFetchWithPrefs({ ...allEnabled, on_job_failed: false })
    await notifyJobFailed('user-1', baseEmailData)
    expect(mockSend).not.toHaveBeenCalled()
  })

  it('does not send when master disabled', async () => {
    mockFetchWithPrefs({ ...allEnabled, enabled: false })
    await notifyJobFailed('user-1', baseEmailData)
    expect(mockSend).not.toHaveBeenCalled()
  })
})

// ---- notifyPROpened ----

describe('notifyPROpened', () => {
  it('sends with pr_opened template slug', async () => {
    mockFetchWithPrefs(allEnabled)
    await notifyPROpened('user-1', { ...baseEmailData, prUrl: 'https://github.com/owner/repo/pull/5' })

    expect(mockSend).toHaveBeenCalledOnce()
    const opts = mockSend.mock.calls[0][0] as { templateSlug: string; variables: Record<string, string> }
    expect(opts.templateSlug).toBe(TEMPLATE_SLUGS.PR_OPENED)
    expect(opts.variables.pr_url).toBe('https://github.com/owner/repo/pull/5')
  })

  it('does not send when on_pr_opened=false', async () => {
    mockFetchWithPrefs({ ...allEnabled, on_pr_opened: false })
    await notifyPROpened('user-1', baseEmailData)
    expect(mockSend).not.toHaveBeenCalled()
  })
})
