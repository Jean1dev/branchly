import { GitHubMark } from "@/components/icons/github-mark";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import Link from "next/link";

export default function LoginPage() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center px-4 pb-16 pt-24">
      <Card className="w-full max-w-[360px] p-8">
        <Link
          href="/"
          className="block text-center font-mono text-base font-semibold"
        >
          branchly
        </Link>
        <h1 className="mt-8 text-center text-xl font-semibold">
          Welcome to Branchly
        </h1>
        <p className="mt-2 text-center text-sm text-gray-500 dark:text-gray-400">
          Sign in with your GitHub account to continue
        </p>
        <Button className="mt-8 w-full gap-2" size="lg" href="/dashboard">
          <GitHubMark className="h-5 w-5" />
          Continue with GitHub
        </Button>
        <p className="mt-6 text-center text-xs text-gray-500 dark:text-gray-400">
          By signing in you agree to our Terms of Service
        </p>
      </Card>
    </div>
  );
}
