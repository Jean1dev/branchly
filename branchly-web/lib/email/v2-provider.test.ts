import { describe, it, expect, vi, beforeEach } from 'vitest'
import { EmailApiV2Provider } from './v2-provider'
import { TEMPLATE_SLUGS } from './templates'

const BASE_URL = 'https://email.example.com'

function makeFetch(responses: Record<string, { status: number; body?: object }>) {
  return vi.fn((url: string, init?: RequestInit) => {
    const method = (init?.method ?? 'GET').toUpperCase()
    const key = `${method} ${url}`
    const match = responses[key]
    if (!match) {
      return Promise.resolve({ ok: false, status: 404, text: async () => 'not found' })
    }
    return Promise.resolve({
      ok: match.status >= 200 && match.status < 300,
      status: match.status,
      text: async () => JSON.stringify(match.body ?? {}),
      json: async () => match.body ?? {},
    })
  })
}

beforeEach(() => {
  vi.restoreAllMocks()
})

describe('EmailApiV2Provider — template registration', () => {
  it('registers templates with PUT on first send when they already exist', async () => {
    const fetchMock = makeFetch({
      [`PUT ${BASE_URL}/v2/email/templates/${TEMPLATE_SLUGS.JOB_COMPLETED}`]: { status: 200 },
      [`PUT ${BASE_URL}/v2/email/templates/${TEMPLATE_SLUGS.JOB_FAILED}`]: { status: 200 },
      [`PUT ${BASE_URL}/v2/email/templates/${TEMPLATE_SLUGS.PR_OPENED}`]: { status: 200 },
      [`POST ${BASE_URL}/v2/email/send`]: { status: 200, body: { status: 'pending' } },
    })
    global.fetch = fetchMock as unknown as typeof fetch

    const provider = new EmailApiV2Provider(BASE_URL)
    await provider.send({
      to: 'a@b.com',
      subject: 'Test',
      templateSlug: TEMPLATE_SLUGS.JOB_COMPLETED,
      variables: {},
    })

    const putCalls = fetchMock.mock.calls.filter(
      ([, init]) => (init as RequestInit)?.method === 'PUT'
    )
    expect(putCalls).toHaveLength(3)
  })

  it('falls back to POST when PUT returns 404 (template does not exist yet)', async () => {
    const fetchMock = makeFetch({
      [`PUT ${BASE_URL}/v2/email/templates/${TEMPLATE_SLUGS.JOB_COMPLETED}`]: { status: 404 },
      [`PUT ${BASE_URL}/v2/email/templates/${TEMPLATE_SLUGS.JOB_FAILED}`]: { status: 404 },
      [`PUT ${BASE_URL}/v2/email/templates/${TEMPLATE_SLUGS.PR_OPENED}`]: { status: 404 },
      [`POST ${BASE_URL}/v2/email/templates`]: { status: 201, body: { status: 'created' } },
      [`POST ${BASE_URL}/v2/email/send`]: { status: 200, body: { status: 'pending' } },
    })
    global.fetch = fetchMock as unknown as typeof fetch

    const provider = new EmailApiV2Provider(BASE_URL)
    await provider.send({
      to: 'a@b.com',
      subject: 'Test',
      templateSlug: TEMPLATE_SLUGS.JOB_COMPLETED,
      variables: {},
    })

    const postTemplateCalls = fetchMock.mock.calls.filter(
      ([url, init]) =>
        (url as string).endsWith('/v2/email/templates') &&
        (init as RequestInit)?.method === 'POST'
    )
    expect(postTemplateCalls).toHaveLength(3)
  })

  it('only registers templates once across multiple sends', async () => {
    const fetchMock = makeFetch({
      [`PUT ${BASE_URL}/v2/email/templates/${TEMPLATE_SLUGS.JOB_COMPLETED}`]: { status: 200 },
      [`PUT ${BASE_URL}/v2/email/templates/${TEMPLATE_SLUGS.JOB_FAILED}`]: { status: 200 },
      [`PUT ${BASE_URL}/v2/email/templates/${TEMPLATE_SLUGS.PR_OPENED}`]: { status: 200 },
      [`POST ${BASE_URL}/v2/email/send`]: { status: 200, body: { status: 'pending' } },
    })
    global.fetch = fetchMock as unknown as typeof fetch

    const provider = new EmailApiV2Provider(BASE_URL)
    await provider.send({ to: 'a@b.com', subject: 'S1', templateSlug: TEMPLATE_SLUGS.JOB_COMPLETED, variables: {} })
    await provider.send({ to: 'b@b.com', subject: 'S2', templateSlug: TEMPLATE_SLUGS.JOB_FAILED, variables: {} })

    const putCalls = fetchMock.mock.calls.filter(
      ([, init]) => (init as RequestInit)?.method === 'PUT'
    )
    // Templates are registered once only, not once per send.
    expect(putCalls).toHaveLength(3)
  })
})

describe('EmailApiV2Provider — send', () => {
  it('sends POST to /v2/email/send with correct payload', async () => {
    const fetchMock = makeFetch({
      [`PUT ${BASE_URL}/v2/email/templates/${TEMPLATE_SLUGS.JOB_COMPLETED}`]: { status: 200 },
      [`PUT ${BASE_URL}/v2/email/templates/${TEMPLATE_SLUGS.JOB_FAILED}`]: { status: 200 },
      [`PUT ${BASE_URL}/v2/email/templates/${TEMPLATE_SLUGS.PR_OPENED}`]: { status: 200 },
      [`POST ${BASE_URL}/v2/email/send`]: { status: 200, body: { status: 'pending' } },
    })
    global.fetch = fetchMock as unknown as typeof fetch

    const provider = new EmailApiV2Provider(BASE_URL)
    await provider.send({
      to: 'user@example.com',
      subject: 'Job done',
      templateSlug: TEMPLATE_SLUGS.JOB_COMPLETED,
      variables: { repo_full_name: 'owner/repo', agent_name: 'Claude Code' },
    })

    const sendCall = fetchMock.mock.calls.find(
      ([url, init]) =>
        (url as string).endsWith('/v2/email/send') &&
        (init as RequestInit)?.method === 'POST'
    )
    expect(sendCall).toBeDefined()
    const body = JSON.parse((sendCall![1] as RequestInit).body as string)
    expect(body.to).toBe('user@example.com')
    expect(body.subject).toBe('Job done')
    expect(body.templateSlug).toBe(TEMPLATE_SLUGS.JOB_COMPLETED)
    expect(body.variables.repo_full_name).toBe('owner/repo')
  })

  it('throws when send API returns non-ok status', async () => {
    const fetchMock = makeFetch({
      [`PUT ${BASE_URL}/v2/email/templates/${TEMPLATE_SLUGS.JOB_COMPLETED}`]: { status: 200 },
      [`PUT ${BASE_URL}/v2/email/templates/${TEMPLATE_SLUGS.JOB_FAILED}`]: { status: 200 },
      [`PUT ${BASE_URL}/v2/email/templates/${TEMPLATE_SLUGS.PR_OPENED}`]: { status: 200 },
      [`POST ${BASE_URL}/v2/email/send`]: { status: 500 },
    })
    global.fetch = fetchMock as unknown as typeof fetch

    const provider = new EmailApiV2Provider(BASE_URL)
    await expect(
      provider.send({ to: 'a@b.com', subject: 'S', templateSlug: TEMPLATE_SLUGS.JOB_COMPLETED, variables: {} })
    ).rejects.toThrow('[EmailApiV2] send failed')
  })

  it('uses custom fromEmail when provided', async () => {
    const fetchMock = makeFetch({
      [`PUT ${BASE_URL}/v2/email/templates/${TEMPLATE_SLUGS.JOB_COMPLETED}`]: { status: 200 },
      [`PUT ${BASE_URL}/v2/email/templates/${TEMPLATE_SLUGS.JOB_FAILED}`]: { status: 200 },
      [`PUT ${BASE_URL}/v2/email/templates/${TEMPLATE_SLUGS.PR_OPENED}`]: { status: 200 },
      [`POST ${BASE_URL}/v2/email/send`]: { status: 200, body: { status: 'pending' } },
    })
    global.fetch = fetchMock as unknown as typeof fetch

    const provider = new EmailApiV2Provider(BASE_URL, 'custom@branchly.com')
    await provider.send({ to: 'a@b.com', subject: 'S', templateSlug: TEMPLATE_SLUGS.JOB_COMPLETED, variables: {} })

    const putCall = fetchMock.mock.calls.find(
      ([, init]) => (init as RequestInit)?.method === 'PUT'
    )
    const body = JSON.parse((putCall![1] as RequestInit).body as string)
    expect(body.fromEmail).toBe('custom@branchly.com')
  })

  it('strips trailing slash from base URL', async () => {
    const fetchMock = makeFetch({
      [`PUT ${BASE_URL}/v2/email/templates/${TEMPLATE_SLUGS.JOB_COMPLETED}`]: { status: 200 },
      [`PUT ${BASE_URL}/v2/email/templates/${TEMPLATE_SLUGS.JOB_FAILED}`]: { status: 200 },
      [`PUT ${BASE_URL}/v2/email/templates/${TEMPLATE_SLUGS.PR_OPENED}`]: { status: 200 },
      [`POST ${BASE_URL}/v2/email/send`]: { status: 200, body: { status: 'pending' } },
    })
    global.fetch = fetchMock as unknown as typeof fetch

    const provider = new EmailApiV2Provider(`${BASE_URL}/`) // trailing slash
    await provider.send({ to: 'a@b.com', subject: 'S', templateSlug: TEMPLATE_SLUGS.JOB_COMPLETED, variables: {} })

    const sendCall = fetchMock.mock.calls.find(
      ([url, init]) =>
        (url as string).includes('/v2/email/send') &&
        (init as RequestInit)?.method === 'POST'
    )
    expect(sendCall![0]).toBe(`${BASE_URL}/v2/email/send`)
  })
})
