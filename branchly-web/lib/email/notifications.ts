import { getEmailProvider } from './index'
import {
  jobCompletedTemplate,
  jobFailedTemplate,
  prOpenedTemplate,
  type JobEmailData,
} from './templates'

interface UserNotifData {
  email: string
  name: string
  notification_preferences: {
    email: {
      enabled: boolean
      on_job_completed: boolean
      on_job_failed: boolean
      on_pr_opened: boolean
    }
  }
}

async function fetchUserNotifData(userId: string): Promise<UserNotifData | null> {
  try {
    const res = await fetch(
      `${process.env.API_URL}/internal/users/${userId}/notification-preferences`,
      {
        headers: { 'X-Internal-Secret': process.env.INTERNAL_API_SECRET ?? '' },
        cache: 'no-store',
      }
    )
    if (!res.ok) return null
    const json = await res.json() as { data?: UserNotifData }
    return (json.data as UserNotifData) ?? null
  } catch (err) {
    console.error('[notifications] failed to fetch user notification data', err)
    return null
  }
}

async function sendSafe(to: string, subject: string, html: string): Promise<void> {
  try {
    await getEmailProvider().send({ to, subject, html })
  } catch (err) {
    console.error('[notifications] email send failed', err)
  }
}

export async function notifyJobCompleted(userId: string, data: JobEmailData): Promise<void> {
  const user = await fetchUserNotifData(userId)
  if (!user?.notification_preferences.email.enabled) return
  if (!user.notification_preferences.email.on_job_completed) return
  const html = jobCompletedTemplate(data)
  await sendSafe(data.userEmail, `Job completed on ${data.repoFullName}`, html)
}

export async function notifyJobFailed(userId: string, data: JobEmailData): Promise<void> {
  const user = await fetchUserNotifData(userId)
  if (!user?.notification_preferences.email.enabled) return
  if (!user.notification_preferences.email.on_job_failed) return
  const html = jobFailedTemplate(data)
  await sendSafe(data.userEmail, `Job failed on ${data.repoFullName}`, html)
}

export async function notifyPROpened(userId: string, data: JobEmailData): Promise<void> {
  const user = await fetchUserNotifData(userId)
  if (!user?.notification_preferences.email.enabled) return
  if (!user.notification_preferences.email.on_pr_opened) return
  const html = prOpenedTemplate(data)
  await sendSafe(data.userEmail, `Pull request opened on ${data.repoFullName}`, html)
}
