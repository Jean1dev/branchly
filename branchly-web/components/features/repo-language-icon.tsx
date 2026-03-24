"use client";

import { cn } from "@/lib/utils";
import { Code2 } from "lucide-react";
import Image from "next/image";
import { useState } from "react";

const SLUGS: Record<string, string> = {
  typescript: "typescript",
  javascript: "javascript",
  ts: "typescript",
  js: "javascript",
  react: "react",
  reactjs: "react",
  vue: "vuedotjs",
  svelte: "svelte",
  angular: "angular",
  astro: "astro",
  nextjs: "nextdotjs",
  nuxt: "nuxtdotjs",
  solidjs: "solid",
  go: "go",
  golang: "go",
  python: "python",
  rust: "rust",
  ruby: "ruby",
  rails: "rubyonrails",
  php: "php",
  java: "openjdk",
  kotlin: "kotlin",
  scala: "scala",
  swift: "swift",
  csharp: "csharp",
  "c#": "csharp",
  fsharp: "fsharp",
  "f#": "fsharp",
  cplusplus: "cplusplus",
  "c++": "cplusplus",
  cpp: "cplusplus",
  c: "c",
  objectivec: "apple",
  "objective-c": "apple",
  dart: "dart",
  flutter: "flutter",
  elixir: "elixir",
  erlang: "erlang",
  haskell: "haskell",
  clojure: "clojure",
  lua: "lua",
  perl: "perl",
  r: "r",
  matlab: "mathworks",
  julia: "julia",
  zig: "zig",
  nim: "nim",
  v: "v",
  solidity: "solidity",
  shell: "gnubash",
  bash: "gnubash",
  zsh: "gnubash",
  powershell: "powershell",
  dockerfile: "docker",
  docker: "docker",
  html: "html5",
  css: "css3",
  scss: "sass",
  sass: "sass",
  less: "less",
  stylus: "stylus",
  markdown: "markdown",
  mdx: "mdx",
  yaml: "yaml",
  json: "json",
  toml: "toml",
  xml: "xml",
  graphql: "graphql",
  sql: "postgresql",
  mysql: "mysql",
  postgresql: "postgresql",
  postgres: "postgresql",
  mongodb: "mongodb",
  redis: "redis",
  sqlite: "sqlite",
  terraform: "terraform",
  hcl: "terraform",
  ansible: "ansible",
  kubernetes: "kubernetes",
  helm: "helm",
  nginx: "nginx",
  jupyter: "jupyter",
  notebook: "jupyter",
  arduino: "arduino",
  tex: "latex",
  latex: "latex",
  vim: "vim",
  neovim: "neovim",
  emacs: "gnuemacs",
  makefile: "gnu",
  cmake: "cmake",
  gradle: "gradle",
  maven: "apachemaven",
  dotnet: "dotnet",
  deno: "deno",
  bun: "bun",
  node: "nodedotjs",
  swiftui: "swift",
  wasm: "webassembly",
  webassembly: "webassembly",
  cobol: "gnucobol",
  fortran: "fortran",
  groovy: "apachegroovy",
  crystal: "crystal",
  ocaml: "ocaml",
  reason: "reason",
  rescript: "rescript",
  purescript: "purescript",
  elm: "elm",
  clojurescript: "clojure",
  wasmcloud: "wasmcloud",
  gleam: "gleam",
  typst: "typst",
};

function slugForLanguage(lang: string | undefined | null): string | null {
  const raw = (lang ?? "").trim().toLowerCase();
  if (!raw) return null;
  if (SLUGS[raw]) return SLUGS[raw];
  const compact = raw.replace(/[\s._-]+/g, "");
  if (SLUGS[compact]) return SLUGS[compact];
  const first = raw.split(/[\s/]+/)[0];
  if (first && SLUGS[first]) return SLUGS[first];
  return null;
}

const SOFTEN_ON_DARK_SLUGS = new Set(["javascript"]);

export type RepoLanguageIconProps = {
  language?: string | null;
  className?: string;
  size?: "sm" | "md";
};

export function RepoLanguageIcon({
  language,
  className,
  size = "md",
}: RepoLanguageIconProps) {
  const slug = slugForLanguage(language);
  const [iconFailed, setIconFailed] = useState(false);

  const box =
    size === "sm"
      ? "h-7 w-7 rounded-md p-0.5"
      : "h-9 w-9 rounded-md p-1";
  const img = size === "sm" ? "h-5 w-5" : "h-6 w-6";

  if (!slug || iconFailed) {
    return (
      <span
        className={cn(
          "flex shrink-0 items-center justify-center border border-gray-200 bg-gray-50 text-gray-600 dark:border-gray-700 dark:bg-gray-900 dark:text-gray-300",
          box,
          className
        )}
        title={language?.trim() || "Language"}
        aria-hidden
      >
        <Code2 className={size === "sm" ? "h-3.5 w-3.5" : "h-4 w-4"} />
      </span>
    );
  }

  const src = `https://cdn.simpleicons.org/${slug}`;
  const softenDark = SOFTEN_ON_DARK_SLUGS.has(slug);

  return (
    <span
      className={cn(
        "flex shrink-0 items-center justify-center overflow-hidden border border-gray-200 bg-white dark:border-neutral-500 dark:bg-neutral-200",
        softenDark && "dark:border-zinc-700 dark:bg-zinc-900",
        box,
        className
      )}
      title={language?.trim() || undefined}
      aria-hidden
    >
      <Image
        src={src}
        alt=""
        width={24}
        height={24}
        className={cn(
          img,
          "object-contain",
          softenDark &&
            "dark:brightness-[0.72] dark:saturate-[0.72] dark:contrast-[1.05]"
        )}
        unoptimized
        onError={() => setIconFailed(true)}
      />
    </span>
  );
}
