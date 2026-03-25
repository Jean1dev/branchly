import Link from "next/link";

const GITHUB_URL = "https://github.com/Jean1dev/branchly";

const links = [
  { href: GITHUB_URL, label: "GitHub", external: true },
  { href: "#", label: "Docs", external: false },
  { href: "#", label: "Status", external: false },
];

export function LandingFooter() {
  return (
    <footer className="border-t border-gray-200 px-4 py-10 dark:border-gray-800 sm:px-6">
      <div className="mx-auto flex max-w-[1200px] flex-col items-start justify-between gap-6 sm:flex-row sm:items-center">
        <div>
          <p className="font-mono text-sm font-semibold text-gray-500 dark:text-gray-400">
            branchly
          </p>
          <p className="mt-1 text-xs text-gray-400 dark:text-gray-600">
            Free · Open source · Built in public
          </p>
        </div>
        <nav className="flex flex-wrap gap-6" aria-label="Footer">
          {links.map((l) => (
            <Link
              key={l.label}
              href={l.href}
              {...(l.external
                ? { target: "_blank", rel: "noopener noreferrer" }
                : {})}
              className="text-sm text-gray-500 transition-colors duration-150 hover:text-foreground dark:text-gray-400"
            >
              {l.label}
            </Link>
          ))}
        </nav>
      </div>
    </footer>
  );
}
