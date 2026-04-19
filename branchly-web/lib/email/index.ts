import { type EmailProvider, StubEmailProvider } from './provider'

let provider: EmailProvider

export function getEmailProvider(): EmailProvider {
  if (!provider) {
    // Future: switch on EMAIL_PROVIDER env var to select real implementations
    provider = new StubEmailProvider()
  }
  return provider
}
