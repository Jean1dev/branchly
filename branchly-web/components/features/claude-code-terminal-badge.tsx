import { ClaudeCodePixelIcon } from "@/components/features/claude-code-pixel-icon";
import { cn } from "@/lib/utils";

type ClaudeCodeTerminalBadgeProps = {
  showPoweredByLine?: boolean;
  className?: string;
  innerClassName?: string;
};

export function ClaudeCodeTerminalBadge({
  showPoweredByLine = true,
  className,
  innerClassName,
}: ClaudeCodeTerminalBadgeProps) {
  return (
    <div className={cn("w-full", className)}>
      <div
        className={cn(
          "overflow-hidden rounded-md border border-white/15 bg-[#300a24] text-left font-mono text-[11px] leading-snug text-gray-300 shadow-none ring-1 ring-black/30 sm:text-xs",
          innerClassName
        )}
        role="img"
        aria-label="Claude Code no terminal — agente Branchly"
      >
        <div className="border-b border-white/10 px-2.5 py-1.5 sm:px-3 sm:py-2">
          <span className="text-[#d37a5a]">Tip:</span>{" "}
          <span className="text-gray-400">
            Press{" "}
            <kbd className="rounded border border-white/20 bg-black/25 px-1 py-px font-mono text-[10px] text-gray-300 sm:text-[11px]">
              ?
            </kbd>{" "}
            for shortcuts
          </span>
        </div>
        <div className="flex items-start gap-3 px-2.5 py-3 sm:gap-4 sm:px-3 sm:py-3.5">
          <ClaudeCodePixelIcon className="h-11 w-11 sm:h-[52px] sm:w-[52px]" />
          <div className="min-w-0 flex-1 space-y-1 pt-0.5">
            <div className="whitespace-nowrap">
              <span className="font-semibold text-white">Claude Code</span>
              <span className="text-gray-500"> 2.0.0</span>
            </div>
            <div className="truncate text-gray-400">
              Sonnet 4.6 · API Usage Billing
            </div>
            <div className="truncate text-gray-500">~/projects/branchly</div>
          </div>
        </div>
        <div className="flex items-center gap-2 border-t border-white/10 px-2.5 py-2 sm:px-3">
          <span className="select-none font-mono text-sm text-white">&gt;</span>
          <span
            className="inline-block h-3.5 w-2 bg-gray-400"
            aria-hidden
          />
        </div>
        <div className="flex items-center justify-between gap-2 border-t border-white/10 px-2.5 py-1.5 text-[10px] text-gray-600 sm:px-3">
          <span className="min-w-0 truncate">? for shortcuts</span>
          <span className="flex shrink-0 items-center gap-1 text-gray-500">
            <span aria-hidden>◐</span>
            <span className="whitespace-nowrap">medium · /effort</span>
          </span>
        </div>
      </div>
      {showPoweredByLine ? (
        <p className="mt-3 text-center text-xs text-gray-500 dark:text-gray-400">
          Powered by · Now in beta
        </p>
      ) : null}
    </div>
  );
}
