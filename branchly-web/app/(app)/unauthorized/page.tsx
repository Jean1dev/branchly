import { UnauthorizedActions } from "@/components/features/unauthorized-actions";
import { Card } from "@/components/ui/card";
import { authOptions } from "@/lib/auth";
import { getServerSession } from "next-auth";

export const metadata = {
  title: "Session issue · Branchly",
};

export default async function UnauthorizedPage() {
  const session = await getServerSession(authOptions);

  return (
    <div className="mx-auto flex min-h-[50vh] max-w-lg flex-col justify-center">
      <Card className="space-y-6 p-8">
        <div className="space-y-2">
          <h1 className="text-xl font-semibold tracking-tight">
            We could not verify your session with the API
          </h1>
          <p className="text-sm text-gray-600 dark:text-gray-400">
            You may be signed in to GitHub, but Branchly could not load the
            token used to talk to the backend. This usually means the API is
            unreachable, environment variables are wrong, or you need a fresh
            sign-in after a configuration change.
          </p>
        </div>
        <UnauthorizedActions isSignedIn={!!session} />
      </Card>
    </div>
  );
}
