export interface EmailPayload {
  to: string
  subject: string
  html: string
}

export interface EmailProvider {
  send(payload: EmailPayload): Promise<void>
}

export class StubEmailProvider implements EmailProvider {
  async send(payload: EmailPayload): Promise<void> {
    console.log('[EmailProvider:stub] Sending email:')
    console.log(`  To: ${payload.to}`)
    console.log(`  Subject: ${payload.subject}`)
    console.log(`  HTML length: ${payload.html.length} chars`)
  }
}
