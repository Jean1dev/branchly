import { ClaudeCodeTerminalBadge } from "@/components/features/claude-code-terminal-badge";
import { Button } from "@/components/ui/button";

export function LandingHero() {
  return (
    <section className="flex min-h-[min(90vh,920px)] flex-col justify-center px-4 pb-20 pt-24 sm:px-6 sm:pb-24 sm:pt-28">
      <div className="mx-auto w-full max-w-6xl">
        <div className="grid items-center gap-12 lg:grid-cols-[1fr_min(380px,42%)] lg:gap-14 xl:gap-20">
          <div className="flex flex-col items-center text-center lg:items-start lg:text-left">
            <p className="mb-4 max-w-xl font-mono text-[11px] leading-relaxed text-gray-500 sm:text-xs">
              <span className="text-gray-400">Agent runtime</span>
              <span className="mx-2 text-gray-700 dark:text-gray-600" aria-hidden>
                ·
              </span>
              Claude Code
              <span className="mx-2 text-gray-700 dark:text-gray-600" aria-hidden>
                ·
              </span>
              <span className="rounded border border-gray-700 px-1.5 py-0.5 text-[10px] uppercase tracking-wide text-gray-400 dark:border-gray-600">
                Beta
              </span>
            </p>
            <h1 className="max-w-2xl text-balance font-semibold leading-[1.05] tracking-tight text-foreground [font-size:clamp(2.5rem,5.5vw,4rem)] lg:max-w-3xl">
              Ship features.
              <br />
              Just describe them.
            </h1>
            <p className="mt-5 max-w-md text-base text-gray-500 dark:text-gray-400 sm:text-lg">
              Branchly reads your repository, writes the code and opens the pull
              request. You review.
            </p>
            <div className="mt-8 flex w-full max-w-md flex-col gap-3 sm:max-w-none sm:flex-row sm:justify-center lg:justify-start">
              <Button size="lg" href="/login" className="w-full sm:w-auto">
                Start for free
              </Button>
              <Button
                variant="ghost"
                size="lg"
                href="#how"
                className="w-full sm:w-auto"
              >
                See how it works
              </Button>
            </div>
            <p className="mt-5 text-xs text-gray-500 dark:text-gray-400">
              No credit card required · GitHub OAuth only
            </p>
          </div>
          <div className="mx-auto w-full max-w-md lg:mx-0 lg:max-w-none">
            <div className="mb-3 hidden text-left font-mono text-[11px] text-gray-500 lg:block">
              Same toolchain developers already use
            </div>
            <div className="rounded-lg border border-gray-200 bg-gray-50/80 p-3 dark:border-gray-800 dark:bg-gray-950/50 sm:p-4">
              <ClaudeCodeTerminalBadge showPoweredByLine={false} />
            </div>
            <p className="mt-3 text-center text-xs text-gray-500 dark:text-gray-400 lg:text-left">
              Powered by Claude Code · Now in beta
            </p>
          </div>
        </div>
      </div>
    </section>
  );
}
