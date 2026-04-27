import { type EmailProvider, StubEmailProvider } from './provider'
import { EmailApiV2Provider } from './v2-provider'

let provider: EmailProvider

export function getEmailProvider(): EmailProvider {
  if (!provider) {
    const apiUrl = process.env.EMAIL_API_URL
    if (process.env.EMAIL_PROVIDER === 'emailapi-v2' && apiUrl) {
      provider = new EmailApiV2Provider(apiUrl, process.env.EMAIL_FROM)
    } else {
      provider = new StubEmailProvider()
    }
  }
  return provider
}
