const agents = [
  {
    initials: "CC",
    name: "Claude Code",
    desc: "Anthropic's coding agent. State-of-the-art performance on real engineering tasks.",
    tag: "Active · Default",
    tagStyle:
      "border-purple-900/60 bg-purple-950/40 text-purple-400 dark:border-purple-900/50",
    avatarStyle: "bg-purple-950/60 text-purple-400",
    active: true,
  },
  {
    initials: "GP",
    name: "GPT-4o / Codex",
    desc: "OpenAI models via API. Bring your own key.",
    tag: "Active",
    tagStyle:
      "border-green-900/60 bg-green-950/30 text-green-400 dark:border-green-900/50",
    avatarStyle: "bg-green-950/40 text-green-500",
    active: true,
  },
  {
    initials: "GE",
    name: "Gemini",
    desc: "Google's models for coding tasks via Vertex AI.",
    tag: "Coming soon",
    tagStyle:
      "border-gray-700/60 bg-transparent text-gray-600 dark:border-gray-700",
    avatarStyle: "bg-yellow-950/30 text-yellow-500",
    active: false,
  },
  {
    initials: "?",
    name: "Your own agent",
    desc: "Implement the Agent interface and plug in any model or custom logic.",
    tag: "Open source",
    tagStyle:
      "border-green-900/60 bg-green-950/30 text-green-500 dark:border-green-900/50",
    avatarStyle:
      "border border-dashed border-blue-800/60 bg-blue-950/20 text-blue-400",
    active: false,
  },
];

export function LandingAgents() {
  return (
    <section className="border-t border-gray-200 px-4 py-24 dark:border-gray-800 sm:px-6">
      <div className="mx-auto max-w-[1200px]">
        <p className="mb-3 font-mono text-[11px] uppercase tracking-widest text-gray-500">
          Multi-agent
        </p>
        <h2 className="text-2xl font-bold tracking-tight text-foreground sm:text-3xl">
          Choose your agent
        </h2>
        <p className="mt-2 text-base text-gray-500 dark:text-gray-400">
          Built to support multiple AI agents — including your own.
        </p>

        <div className="mt-12 grid gap-3 sm:grid-cols-2">
          {agents.map((a) => (
            <div
              key={a.name}
              className={`flex items-start gap-4 rounded-xl border p-5 ${
                a.active
                  ? "border-purple-900/50 bg-purple-950/10 dark:border-purple-900/40 dark:bg-purple-950/10"
                  : "border-gray-200 bg-background dark:border-gray-800"
              }`}
            >
              <div
                className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-lg font-mono text-sm font-bold ${a.avatarStyle}`}
              >
                {a.initials}
              </div>
              <div>
                <p className="text-sm font-semibold text-foreground">
                  {a.name}
                </p>
                <p className="mt-1 text-xs leading-relaxed text-gray-500 dark:text-gray-400">
                  {a.desc}
                </p>
                <span
                  className={`mt-2 inline-block rounded border px-2 py-0.5 font-mono text-[10px] ${a.tagStyle}`}
                >
                  {a.tag}
                </span>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
