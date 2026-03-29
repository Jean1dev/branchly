const features = [
  {
    icon: (
      <svg
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth={2}
        className="h-4 w-4 text-green-400"
      >
        <path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5" />
      </svg>
    ),
    title: "100% free",
    body: "No pricing tiers. No limits hidden behind a paywall. Free forever.",
  },
  {
    icon: (
      <svg
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth={2}
        className="h-4 w-4 text-purple-400"
      >
        <path d="M9 19c-5 1.5-5-2.5-7-3m14 6v-3.87a3.37 3.37 0 00-.94-2.61c3.14-.35 6.44-1.54 6.44-7A5.44 5.44 0 0020 4.77 5.07 5.07 0 0019.91 1S18.73.65 16 2.48a13.38 13.38 0 00-7 0C6.27.65 5.09 1 5.09 1A5.07 5.07 0 005 4.77a5.44 5.44 0 00-1.5 3.78c0 5.42 3.3 6.61 6.44 7A3.37 3.37 0 009 18.13V22" />
      </svg>
    ),
    title: "Open source",
    body: "Every line of code is public. Audit it, fork it, contribute to it.",
  },
  {
    icon: (
      <svg
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth={2}
        className="h-4 w-4 text-blue-400"
      >
        <rect x="2" y="3" width="20" height="14" rx="2" />
        <path d="M8 21h8M12 17v4" />
      </svg>
    ),
    title: "Self-hostable",
    body: "Run on your own infra with Docker Compose. Your code never leaves your servers.",
  },
  {
    icon: (
      <svg
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth={2}
        className="h-4 w-4 text-orange-400"
      >
        <circle cx="12" cy="12" r="3" />
        <path d="M12 1v4M12 19v4M4.22 4.22l2.83 2.83M16.95 16.95l2.83 2.83M1 12h4M19 12h4M4.22 19.78l2.83-2.83M16.95 7.05l2.83-2.83" />
      </svg>
    ),
    title: "Multi-agent",
    body: "Plug in Claude, GPT, Gemini or your own agent. The interface is open.",
  },
  {
    icon: (
      <svg
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth={2}
        className="h-4 w-4 text-pink-400"
      >
        <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
      </svg>
    ),
    title: "Secure by design",
    body: "OAuth only. Credentials encrypted at rest. Ephemeral containers per job.",
  },
  {
    icon: (
      <svg
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth={2}
        className="h-4 w-4 text-emerald-400"
      >
        <polyline points="16 18 22 12 16 6" />
        <polyline points="8 6 2 12 8 18" />
      </svg>
    ),
    title: "Real-time logs",
    body: "Watch the agent work live. Every file read, every edit, streamed in real time.",
  },
];

export function LandingFeatures() {
  return (
    <section className="border-t border-gray-200 px-4 py-24 dark:border-gray-800 sm:px-6">
      <div className="mx-auto max-w-[1200px]">
        <p className="mb-3 font-mono text-[11px] uppercase tracking-widest text-gray-500">
          Why Branchly
        </p>
        <h2 className="text-2xl font-bold tracking-tight text-foreground sm:text-3xl">
          Built different
        </h2>
        <p className="mt-2 text-base text-gray-500 dark:text-gray-400">
          Designed for developers who care about ownership.
        </p>

        <div className="mt-12 grid gap-px rounded-xl border border-gray-200 bg-gray-200 overflow-hidden sm:grid-cols-2 lg:grid-cols-3 dark:border-gray-800 dark:bg-gray-800">
          {features.map((f) => (
            <div key={f.title} className="bg-background px-7 py-8">
              <div className="flex h-8 w-8 items-center justify-center rounded-lg border border-gray-200 dark:border-gray-800">
                {f.icon}
              </div>
              <h3 className="mt-4 text-sm font-semibold text-foreground">
                {f.title}
              </h3>
              <p className="mt-1.5 text-sm leading-relaxed text-gray-500 dark:text-gray-400">
                {f.body}
              </p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
