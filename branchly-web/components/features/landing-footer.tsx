import Link from "next/link";

const links = [
  { href: "https://github.com", label: "GitHub" },
  { href: "#", label: "Docs" },
  { href: "#", label: "Status" },
];

export function LandingFooter() {
  return (
    <footer className="border-t border-gray-200 px-4 py-10 dark:border-gray-800 sm:px-6">
      <div className="mx-auto flex max-w-[1200px] flex-col items-start justify-between gap-6 sm:flex-row sm:items-center">
        <p className="font-mono text-sm text-gray-500 dark:text-gray-400">
          branchly · © 2024 Branchly
        </p>
        <nav className="flex flex-wrap gap-6" aria-label="Footer">
          {links.map((l) => (
            <Link
              key={l.label}
              href={l.href}
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
