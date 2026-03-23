import { Card } from "@/components/ui/card";

const steps = [
  {
    n: "1",
    title: "Connect your repo",
    body: "Link your GitHub repository in one click",
  },
  {
    n: "2",
    title: "Describe the task",
    body: "Write what you want in plain english",
  },
  {
    n: "3",
    title: "Review the PR",
    body: "Branchly writes the code and opens the pull request",
  },
];

export function LandingSteps() {
  return (
    <section
      id="how"
      className="border-t border-gray-200 px-4 py-24 dark:border-gray-800 sm:px-6"
    >
      <div className="mx-auto max-w-[1200px]">
        <h2 className="text-center text-2xl font-semibold tracking-tight sm:text-3xl">
          From description to pull request in minutes
        </h2>
        <div className="mt-12 grid gap-6 md:grid-cols-3">
          {steps.map((s) => (
            <Card key={s.n} className="p-6">
              <span className="font-mono text-sm font-semibold text-gray-500 dark:text-gray-400">
                {s.n}
              </span>
              <h3 className="mt-3 text-lg font-semibold">{s.title}</h3>
              <p className="mt-2 text-sm text-gray-500 dark:text-gray-400">
                {s.body}
              </p>
            </Card>
          ))}
        </div>
      </div>
    </section>
  );
}
