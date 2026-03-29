import { Card } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import type { JobCost } from "@/types";

function formatUSD(v: number): string {
  if (v < 0.01) return "< $0.01";
  return `$${v.toFixed(4)}`;
}

function formatTokens(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`;
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}k`;
  return String(n);
}

function formatDuration(secs: number): string {
  if (secs < 60) return `${Math.round(secs)}s`;
  const m = Math.floor(secs / 60);
  const s = Math.round(secs % 60);
  return s > 0 ? `${m}m ${s}s` : `${m}m`;
}

interface Props {
  cost: JobCost;
}

export function JobCostCard({ cost }: Props) {
  return (
    <Card className="space-y-4 p-6">
      <div className="flex items-center justify-between">
        <p className="text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">
          Estimated cost
        </p>
        {cost.isEstimate && (
          <span
            className="rounded bg-amber-100 px-1.5 py-0.5 text-[10px] font-medium text-amber-700 dark:bg-amber-900/30 dark:text-amber-400"
            title="Tokens are estimated from execution duration — not exact values from the model API."
          >
            Estimate
          </span>
        )}
      </div>

      <div>
        <p className="text-2xl font-semibold tabular-nums">
          {formatUSD(cost.estimatedUSD)}
        </p>
        <p className="mt-0.5 text-xs text-gray-500 dark:text-gray-400">
          {cost.modelUsed}
        </p>
      </div>

      <Separator />

      <div className="grid grid-cols-2 gap-3 text-sm">
        <div>
          <p className="text-xs text-gray-500 dark:text-gray-400">Duration</p>
          <p className="mt-0.5 font-medium tabular-nums">
            {formatDuration(cost.durationSecs)}
          </p>
        </div>
        <div>
          <p className="text-xs text-gray-500 dark:text-gray-400">Total tokens</p>
          <p className="mt-0.5 font-medium tabular-nums">
            {formatTokens(cost.totalTokens)}
          </p>
        </div>
        <div>
          <p className="text-xs text-gray-500 dark:text-gray-400">Input</p>
          <p className="mt-0.5 font-medium tabular-nums">
            {formatTokens(cost.inputTokens)}
          </p>
        </div>
        <div>
          <p className="text-xs text-gray-500 dark:text-gray-400">Output</p>
          <p className="mt-0.5 font-medium tabular-nums">
            {formatTokens(cost.outputTokens)}
          </p>
        </div>
      </div>
    </Card>
  );
}
