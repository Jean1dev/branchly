import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import type { JobEmailData } from './templates'

// We test the notification functions by mocking fetch (which they use to
// get user prefs from branchly-api) and the email provider (via module mock).

const mockSend = vi.fn().mockResolvedValue(undefined)

vi.mock('./index', () => ({
  getEmailProvider: () => ({ send: mockSend }),
}))

// Import after mocks are in place.
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
  it('sends email when enabled and on_job_completed=true', async () => {
    mockFetchWithPrefs({ enabled: true, on_job_completed: true, on_job_failed: true, on_pr_opened: true })
    await notifyJobCompleted('user-1', baseEmailData)
    expect(mockSend).toHaveBeenCalledOnce()
    const call = mockSend.mock.calls[0][0] as { to: string; subject: string; html: string }
    expect(call.to).toBe('user@example.com')
    expect(call.subject).toContain('owner/repo')
    expect(call.html).toContain('<!DOCTYPE html>')
  })

  it('does not send email when master enabled=false', async () => {
    mockFetchWithPrefs({ enabled: false, on_job_completed: true, on_job_failed: true, on_pr_opened: true })
    await notifyJobCompleted('user-1', baseEmailData)
    expect(mockSend).not.toHaveBeenCalled()
  })

  it('does not send email when on_job_completed=false', async () => {
    mockFetchWithPrefs({ enabled: true, on_job_completed: false, on_job_failed: true, on_pr_opened: true })
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
  it('sends email when enabled and on_job_failed=true', async () => {
    mockFetchWithPrefs({ enabled: true, on_job_completed: true, on_job_failed: true, on_pr_opened: true })
    await notifyJobFailed('user-1', { ...baseEmailData, errorMessage: 'timeout' })
    expect(mockSend).toHaveBeenCalledOnce()
    const call = mockSend.mock.calls[0][0] as { html: string }
    expect(call.html).toContain('timeout')
  })

  it('does not send email when on_job_failed=false', async () => {
    mockFetchWithPrefs({ enabled: true, on_job_completed: true, on_job_failed: false, on_pr_opened: true })
    await notifyJobFailed('user-1', baseEmailData)
    expect(mockSend).not.toHaveBeenCalled()
  })
})

// ---- notifyPROpened ----

describe('notifyPROpened', () => {
  it('sends email when enabled and on_pr_opened=true', async () => {
    mockFetchWithPrefs({ enabled: true, on_job_completed: true, on_job_failed: true, on_pr_opened: true })
    await notifyPROpened('user-1', { ...baseEmailData, prUrl: 'https://github.com/owner/repo/pull/5' })
    expect(mockSend).toHaveBeenCalledOnce()
    const call = mockSend.mock.calls[0][0] as { html: string }
    expect(call.html).toContain('Pull request opened')
  })

  it('does not send email when on_pr_opened=false', async () => {
    mockFetchWithPrefs({ enabled: true, on_job_completed: true, on_job_failed: true, on_pr_opened: false })
    await notifyPROpened('user-1', baseEmailData)
    expect(mockSend).not.toHaveBeenCalled()
  })
})
