import { describe, it, expect, vi } from 'vitest'
import { StubEmailProvider } from './provider'

describe('StubEmailProvider', () => {
  it('send() resolves without throwing', async () => {
    const provider = new StubEmailProvider()
    await expect(
      provider.send({
        to: 'a@b.com',
        subject: 'Test',
        templateSlug: 'branchly-job-completed',
        variables: { repo_full_name: 'owner/repo' },
      })
    ).resolves.toBeUndefined()
  })

  it('logs to console without throwing', async () => {
    const provider = new StubEmailProvider()
    const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => undefined)
    await provider.send({
      to: 'test@example.com',
      subject: 'Subject',
      templateSlug: 'branchly-job-failed',
      variables: { error_message: 'timeout' },
    })
    expect(consoleSpy).toHaveBeenCalled()
    consoleSpy.mockRestore()
  })
})
