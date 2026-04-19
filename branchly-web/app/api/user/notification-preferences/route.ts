import { apiFetch } from '@/lib/api-client'

export async function GET() {
  try {
    const res = await apiFetch('/settings/notification-preferences')
    const data = await res.json()
    return Response.json(data, { status: res.status })
  } catch {
    return Response.json({ error: 'Unauthorized' }, { status: 401 })
  }
}

export async function PATCH(request: Request) {
  try {
    const body: unknown = await request.json()
    const res = await apiFetch('/settings/notification-preferences', {
      method: 'PATCH',
      body: JSON.stringify(body),
    })
    const data = await res.json()
    return Response.json(data, { status: res.status })
  } catch {
    return Response.json({ error: 'Unauthorized' }, { status: 401 })
  }
}
