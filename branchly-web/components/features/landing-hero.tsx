import { Button } from "@/components/ui/button";
import Link from "next/link";

const GITHUB_URL = "https://github.com/Jean1dev/branchly";

const badges = [
  { label: "Free & open source", color: "green" },
  { label: "Multi-agent", color: "blue" },
  { label: "Agent runtime", color: "default" },
  { label: "Beta", color: "default" },
] as const;

const badgeClass: Record<string, string> = {
  green:
    "border-green-800/60 bg-green-950/40 text-green-400 dark:border-green-800/40 dark:bg-green-950/30 dark:text-green-400",
  blue: "border-blue-800/60 bg-blue-950/40 text-blue-400 dark:border-blue-800/40 dark:bg-blue-950/30 dark:text-blue-400",
  default:
    "border-gray-700/60 bg-transparent text-gray-500 dark:border-gray-700 dark:text-gray-500",
};

const terminalLines = [
  { color: "text-gray-600", text: "Agent runtime · Claude Code 2.0.0" },
  { color: "text-gray-600", text: "Sonnet 4.6 · API Usage Billing" },
  { color: "", text: "" },
  { color: "text-blue-400", text: "▸ Cloning lucasmendes/ecommerce-api..." },
  { color: "text-blue-400", text: "▸ Indexing codebase (47 files)" },
  { color: "text-purple-400", text: "▸ Agent analyzing task..." },
  { color: "text-purple-400", text: "▸ Reading handlers/products.go" },
  { color: "text-purple-400", text: "▸ Writing models/pagination.go" },
  { color: "text-purple-400", text: "▸ Editing handlers/products.go" },
  { color: "text-green-400", text: "✓ Changes committed" },
  { color: "text-green-400", text: "✓ PR opened → github.com/…/pull/47" },
];

function GitHubIcon() {
  return (
    <svg
      className="h-4 w-4"
      viewBox="0 0 24 24"
      fill="currentColor"
      aria-hidden
    >
      <path d="M12 2C6.477 2 2 6.477 2 12c0 4.42 2.865 8.166 6.839 9.489.5.092.682-.217.682-.482 0-.237-.008-.866-.013-1.7-2.782.603-3.369-1.342-3.369-1.342-.454-1.155-1.11-1.462-1.11-1.462-.908-.62.069-.608.069-.608 1.003.07 1.531 1.03 1.531 1.03.892 1.529 2.341 1.087 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.11-4.555-4.943 0-1.091.39-1.984 1.029-2.683-.103-.253-.446-1.27.098-2.647 0 0 .84-.269 2.75 1.025A9.578 9.578 0 0112 6.836c.85.004 1.705.114 2.504.336 1.909-1.294 2.747-1.025 2.747-1.025.546 1.377.202 2.394.1 2.647.64.699 1.028 1.592 1.028 2.683 0 3.842-2.339 4.687-4.566 4.935.359.309.678.919.678 1.852 0 1.336-.012 2.415-.012 2.741 0 .267.18.578.688.48C19.138 20.163 22 16.418 22 12c0-5.523-4.477-10-10-10z" />
    </svg>
  );
}

function Terminal() {
  return (
    <div className="overflow-hidden rounded-xl border border-gray-800 bg-[#0d0d0d] font-mono text-xs shadow-2xl">
      {/* title bar */}
      <div className="flex items-center gap-1.5 border-b border-gray-800 bg-[#111] px-3.5 py-2.5">
        <span className="h-2.5 w-2.5 rounded-full bg-[#ff5f57]" />
        <span className="h-2.5 w-2.5 rounded-full bg-[#febc2e]" />
        <span className="h-2.5 w-2.5 rounded-full bg-[#28c840]" />
        <span className="ml-2 text-gray-600">~/projects/branchly</span>
      </div>

      {/* body */}
      <div className="space-y-0.5 px-4 py-4 leading-relaxed">
        {terminalLines.map((line, i) =>
          line.text === "" ? (
            <div key={i} className="h-3" />
          ) : (
            <div key={i} className={line.color}>
              {line.text}
            </div>
          )
        )}
        <div className="mt-1 text-gray-600">
          {">"}{" "}
          <span className="inline-block h-3 w-1.5 animate-pulse bg-gray-400 align-middle" />
        </div>
      </div>

      {/* footer */}
      <div className="flex justify-between border-t border-gray-800 bg-[#111] px-3.5 py-1.5 text-[10px] text-gray-700">
        <span>? for shortcuts</span>
        <span>medium · /effort</span>
      </div>
    </div>
  );
}

export function LandingHero() {
  return (
    <section className="flex min-h-[min(90vh,920px)] flex-col justify-center px-4 pb-20 pt-24 sm:px-6 sm:pb-24 sm:pt-28">
      <div className="mx-auto w-full max-w-6xl">
        <div className="grid items-center gap-12 lg:grid-cols-[1fr_min(400px,44%)] lg:gap-16">
          {/* left */}
          <div className="flex flex-col items-center text-center lg:items-start lg:text-left">
            {/* badges */}
            <div className="mb-6 flex flex-wrap justify-center gap-2 lg:justify-start">
              {badges.map((b) => (
                <span
                  key={b.label}
                  className={`rounded-full border px-3 py-1 font-mono text-[11px] ${badgeClass[b.color]}`}
                >
                  {b.label}
                </span>
              ))}
            </div>

            <h1 className="max-w-2xl text-balance font-bold leading-[1.05] tracking-tight text-foreground [font-size:clamp(2.6rem,5.5vw,4.2rem)]">
              Ship features.
              <br />
              Just describe
              <br />
              them.
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
                See how it works →
              </Button>
            </div>

            <p className="mt-5 text-xs text-gray-500 dark:text-gray-400">
              No credit card required · GitHub OAuth only · Self-hostable
            </p>

            {/* github link */}
            <Link
              href={GITHUB_URL}
              target="_blank"
              rel="noopener noreferrer"
              className="mt-6 flex items-center gap-2 text-xs text-gray-500 transition-colors hover:text-gray-900 dark:hover:text-gray-200"
            >
              <GitHubIcon />
              github.com/Jean1dev/branchly
            </Link>
          </div>

          {/* right — terminal */}
          <div className="mx-auto w-full max-w-md lg:mx-0 lg:max-w-none">
            <p className="mb-3 hidden text-left font-mono text-[11px] text-gray-500 lg:block">
              Same toolchain developers already use
            </p>
            <Terminal />
            <p className="mt-3 text-center text-xs text-gray-500 dark:text-gray-400 lg:text-left">
              Powered by Claude Code · Now in beta
            </p>
          </div>
        </div>
      </div>
    </section>
  );
}
