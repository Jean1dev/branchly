"use client";

import { useEffect, useState } from "react";

const REPO = "Jean1dev/branchly";

export function GitHubStars() {
  const [stars, setStars] = useState<string | null>(null);

  useEffect(() => {
    fetch(`https://api.github.com/repos/${REPO}`)
      .then((r) => r.json())
      .then((d) => {
        const n: number = d.stargazers_count;
        if (typeof n === "number") {
          setStars(n >= 1000 ? (n / 1000).toFixed(1) + "k" : String(n));
        }
      })
      .catch(() => setStars("0"));
  }, []);

  return (
    <span className="flex items-center gap-1.5 rounded-lg border border-gray-200 bg-gray-50 px-3 py-1.5 font-mono text-xs text-gray-500 dark:border-gray-800 dark:bg-gray-950 dark:text-gray-400">
      <span className="text-yellow-400" aria-hidden>
        ★
      </span>
      <span>{stars ?? "—"}</span>
    </span>
  );
}
