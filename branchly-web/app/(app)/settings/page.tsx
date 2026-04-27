import { DisconnectAccountButton } from "@/components/features/disconnect-account-button";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { PageHeader } from "@/components/layout/page-header";
import { authOptions } from "@/lib/auth";
import { Skeleton } from "@/components/skeletons/skeleton";
import type { User } from "@/types";
import { getServerSession } from "next-auth";
import type { Session } from "next-auth";
import { Suspense } from "react";

function sessionToUser(session: Session): User {
  const email = session.user?.email ?? "";
  const githubUsername =
    session.githubLogin ??
    (email.includes("@") ? (email.split("@")[0]?.replace(/^\d+\+/, "") ?? "github") : "github");
  return {
    id: session.userId,
    name: session.user?.name ?? "User",
    email,
    avatar: session.user?.image ?? "",
    githubUsername,
  };
}

async function SettingsBody() {
  const session = await getServerSession(authOptions);
  if (!session?.userId) {
    return (
      <p className="text-sm text-gray-500 dark:text-gray-400">
        Could not load profile.
      </p>
    );
  }
  const user = sessionToUser(session);

  return (
    <div className="max-w-2xl space-y-10">
      <section className="space-y-4">
        <h2 className="text-lg font-semibold">Profile</h2>
        <Card className="space-y-4 p-6">
          <div>
            <p className="text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">
              Name
            </p>
            <p className="mt-1 text-sm">{user.name}</p>
          </div>
          <Separator />
          <div>
            <p className="text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">
              Email
            </p>
            <p className="mt-1 text-sm">{user.email}</p>
          </div>
          <Separator />
          <div>
            <p className="text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">
              GitHub username
            </p>
            <p className="mt-1 font-mono text-sm">{user.githubUsername}</p>
          </div>
        </Card>
      </section>
      <section className="space-y-4">
        <h2 className="text-lg font-semibold">Connected accounts</h2>
        <Card className="flex flex-wrap items-center justify-between gap-4 p-6">
          <div>
            <p className="font-medium">GitHub</p>
            <p className="text-sm text-gray-500 dark:text-gray-400">
              @{user.githubUsername}
            </p>
          </div>
          <Badge variant="success">Connected</Badge>
        </Card>
      </section>
      <section className="space-y-4">
        <h2 className="text-lg font-semibold">Git integrations</h2>
        <Card className="flex flex-wrap items-center justify-between gap-4 p-6">
          <div>
            <p className="font-medium">Manage integrations</p>
            <p className="text-sm text-gray-500 dark:text-gray-400">
              Connect GitHub and GitLab to access your repositories.
            </p>
          </div>
          <Button variant="secondary" size="sm" href="/settings/integrations">
            Manage →
          </Button>
        </Card>
      </section>
      <section className="space-y-4">
        <h2 className="text-lg font-semibold">Notifications</h2>
        <Card className="flex flex-wrap items-center justify-between gap-4 p-6">
          <div>
            <p className="font-medium">Email notifications</p>
            <p className="text-sm text-gray-500 dark:text-gray-400">
              Choose which job and pull request updates you receive by email.
            </p>
          </div>
          <Button variant="secondary" size="sm" href="/settings/notifications">
            Manage →
          </Button>
        </Card>
      </section>
      <section className="space-y-4">
        <h2 className="text-lg font-semibold">Danger zone</h2>
        <Card className="p-6">
          <DisconnectAccountButton />
        </Card>
      </section>
    </div>
  );
}

function SettingsSkeleton() {
  return (
    <div className="max-w-2xl space-y-10">
      <Skeleton className="h-40 w-full rounded-lg" />
      <Skeleton className="h-24 w-full rounded-lg" />
      <Skeleton className="h-16 w-full rounded-lg" />
    </div>
  );
}

export default function SettingsPage() {
  return (
    <>
      <PageHeader title="Settings" />
      <Suspense fallback={<SettingsSkeleton />}>
        <SettingsBody />
      </Suspense>
    </>
  );
}
