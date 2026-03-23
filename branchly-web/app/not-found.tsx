import { Button } from "@/components/ui/button";
import Link from "next/link";

export default function NotFound() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-6 px-4">
      <h1 className="text-2xl font-semibold">Page not found</h1>
      <p className="text-center text-sm text-gray-500 dark:text-gray-400">
        The page you are looking for does not exist.
      </p>
      <Button href="/dashboard" size="lg">
        Go to dashboard
      </Button>
      <Link href="/" className="text-sm text-gray-500 hover:underline dark:text-gray-400">
        Back to home
      </Link>
    </div>
  );
}
