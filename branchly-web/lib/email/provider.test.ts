import { describe, it, expect, vi } from 'vitest'
import { StubEmailProvider } from './provider'

describe('StubEmailProvider', () => {
  it('send() resolves without throwing', async () => {
    const provider = new StubEmailProvider()
    await expect(
      provider.send({ to: 'a@b.com', subject: 'Test', html: '<p>hi</p>' })
    ).resolves.toBeUndefined()
  })

  it('logs to console without throwing on long HTML', async () => {
    const provider = new StubEmailProvider()
    const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => undefined)
    await provider.send({
      to: 'test@example.com',
      subject: 'Long email',
      html: '<p>' + 'x'.repeat(5000) + '</p>',
    })
    expect(consoleSpy).toHaveBeenCalled()
    consoleSpy.mockRestore()
  })
})
