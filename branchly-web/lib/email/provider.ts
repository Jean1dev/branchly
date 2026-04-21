export interface EmailSendOptions {
  to: string
  subject: string
  templateSlug: string
  variables: Record<string, string>
}

export interface EmailProvider {
  send(options: EmailSendOptions): Promise<void>
}

export class StubEmailProvider implements EmailProvider {
  async send(options: EmailSendOptions): Promise<void> {
    console.log('[EmailProvider:stub] Sending email:')
    console.log(`  To: ${options.to}`)
    console.log(`  Subject: ${options.subject}`)
    console.log(`  Template: ${options.templateSlug}`)
    console.log(`  Variables: ${JSON.stringify(options.variables)}`)
  }
}
