import { DisconnectAccountButton } from "@/components/features/disconnect-account-button";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { PageHeader } from "@/components/layout/page-header";
import { mockUser } from "@/lib/mock-data";
import { delay } from "@/lib/utils";
import { Suspense } from "react";
import { Skeleton } from "@/components/skeletons/skeleton";

async function SettingsBody() {
  await delay(320);
  return (
    <div className="max-w-2xl space-y-10">
      <section className="space-y-4">
        <h2 className="text-lg font-semibold">Profile</h2>
        <Card className="space-y-4 p-6">
          <div>
            <p className="text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">
              Name
            </p>
            <p className="mt-1 text-sm">{mockUser.name}</p>
          </div>
          <Separator />
          <div>
            <p className="text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">
              Email
            </p>
            <p className="mt-1 text-sm">{mockUser.email}</p>
          </div>
          <Separator />
          <div>
            <p className="text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">
              GitHub username
            </p>
            <p className="mt-1 font-mono text-sm">{mockUser.githubUsername}</p>
          </div>
        </Card>
      </section>
      <section className="space-y-4">
        <h2 className="text-lg font-semibold">Connected accounts</h2>
        <Card className="flex flex-wrap items-center justify-between gap-4 p-6">
          <div>
            <p className="font-medium">GitHub</p>
            <p className="text-sm text-gray-500 dark:text-gray-400">
              @{mockUser.githubUsername}
            </p>
          </div>
          <Badge variant="success">Connected</Badge>
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
