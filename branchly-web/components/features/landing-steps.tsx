const steps = [
  {
    n: "01",
    title: "Connect your repo",
    body: "Link your GitHub repository in one click via OAuth. Branchly never stores your code.",
  },
  {
    n: "02",
    title: "Describe the task",
    body: "Write what you want in plain English. The agent reads your codebase and understands the context.",
  },
  {
    n: "03",
    title: "Review the PR",
    body: "Branchly writes the code, commits to a new branch and opens a pull request. You just review.",
  },
];

export function LandingSteps() {
  return (
    <section
      id="how"
      className="border-t border-gray-200 px-4 py-24 dark:border-gray-800 sm:px-6"
    >
      <div className="mx-auto max-w-[1200px]">
        <p className="mb-3 font-mono text-[11px] uppercase tracking-widest text-gray-500">
          How it works
        </p>
        <h2 className="text-2xl font-bold tracking-tight text-foreground sm:text-3xl">
          From description to pull request
        </h2>
        <p className="mt-2 text-base text-gray-500 dark:text-gray-400">
          Three steps. No setup. No context switching.
        </p>

        <div className="mt-12 grid gap-px rounded-xl border border-gray-200 bg-gray-200 overflow-hidden md:grid-cols-3 dark:border-gray-800 dark:bg-gray-800">
          {steps.map((s, i) => (
            <div
              key={s.n}
              className="relative bg-background px-7 py-8"
            >
              <span className="font-mono text-xs text-gray-400 dark:text-gray-600">
                {s.n}
              </span>
              <h3 className="mt-3 text-base font-semibold text-foreground">
                {s.title}
              </h3>
              <p className="mt-2 text-sm leading-relaxed text-gray-500 dark:text-gray-400">
                {s.body}
              </p>
              {i < steps.length - 1 && (
                <span className="absolute -right-2.5 top-1/2 hidden -translate-y-1/2 text-gray-400 dark:text-gray-600 md:block">
                  ›
                </span>
              )}
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
