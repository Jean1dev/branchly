import type { EmailProvider, EmailSendOptions } from './provider'
import {
  TEMPLATE_SLUGS,
  JOB_COMPLETED_HTML_TEMPLATE,
  JOB_FAILED_HTML_TEMPLATE,
  PR_OPENED_HTML_TEMPLATE,
} from './templates'

interface TemplateDefinition {
  slug: string
  name: string
  htmlTemplate: string
}

const BRANCHLY_TEMPLATES: TemplateDefinition[] = [
  {
    slug: TEMPLATE_SLUGS.JOB_COMPLETED,
    name: 'Branchly — Job Completed',
    htmlTemplate: JOB_COMPLETED_HTML_TEMPLATE,
  },
  {
    slug: TEMPLATE_SLUGS.JOB_FAILED,
    name: 'Branchly — Job Failed',
    htmlTemplate: JOB_FAILED_HTML_TEMPLATE,
  },
  {
    slug: TEMPLATE_SLUGS.PR_OPENED,
    name: 'Branchly — Pull Request Opened',
    htmlTemplate: PR_OPENED_HTML_TEMPLATE,
  },
]

export class EmailApiV2Provider implements EmailProvider {
  private readonly baseUrl: string
  private readonly fromEmail: string
  private initialized = false
  private initPromise: Promise<void> | null = null

  constructor(baseUrl: string, fromEmail = 'noreply@branchly.com') {
    this.baseUrl = baseUrl.replace(/\/$/, '')
    this.fromEmail = fromEmail
  }

  async send(options: EmailSendOptions): Promise<void> {
    await this.ensureTemplates()

    const res = await fetch(`${this.baseUrl}/v2/email/send`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        templateSlug: options.templateSlug,
        to: options.to,
        subject: options.subject,
        variables: options.variables,
      }),
    })

    if (!res.ok) {
      const body = await res.text().catch(() => '')
      throw new Error(`[EmailApiV2] send failed (${res.status}): ${body}`)
    }
  }

  // Ensures all Branchly templates are registered in the API.
  // Runs once per provider instance; subsequent calls are no-ops.
  private ensureTemplates(): Promise<void> {
    if (this.initialized) return Promise.resolve()
    if (this.initPromise) return this.initPromise
    this.initPromise = this.registerTemplates().then(() => {
      this.initialized = true
    })
    return this.initPromise
  }

  private async registerTemplates(): Promise<void> {
    await Promise.all(BRANCHLY_TEMPLATES.map((t) => this.upsertTemplate(t)))
  }

  private async upsertTemplate(t: TemplateDefinition): Promise<void> {
    const body = JSON.stringify({
      name: t.name,
      fromEmail: this.fromEmail,
      htmlTemplate: t.htmlTemplate,
    })
    const headers = { 'Content-Type': 'application/json' }

    // Try to update (PUT) first — works if the template already exists.
    const putRes = await fetch(`${this.baseUrl}/v2/email/templates/${t.slug}`, {
      method: 'PUT',
      headers,
      body,
    })

    if (putRes.ok) return

    if (putRes.status === 404) {
      // Template doesn't exist yet — create it.
      const postRes = await fetch(`${this.baseUrl}/v2/email/templates`, {
        method: 'POST',
        headers,
        body: JSON.stringify({ slug: t.slug, ...JSON.parse(body) }),
      })
      if (!postRes.ok) {
        const err = await postRes.text().catch(() => '')
        console.error(`[EmailApiV2] failed to register template "${t.slug}" (${postRes.status}): ${err}`)
      }
      return
    }

    const err = await putRes.text().catch(() => '')
    console.error(`[EmailApiV2] failed to upsert template "${t.slug}" (${putRes.status}): ${err}`)
  }
}
