"use client";

import { cn } from "@/lib/utils";
import type { JobLog } from "@/types";
import { useEffect, useRef, useState } from "react";

const levelClass: Record<JobLog["level"], string> = {
  info: "text-gray-400",
  success: "text-gray-200",
  warning: "text-gray-300",
  error: "text-gray-500",
};

type LogTerminalProps = {
  lines: JobLog[];
  stream?: boolean;
  intervalMs?: number;
};

export function LogTerminal({
  lines,
  stream = false,
  intervalMs = 400,
}: LogTerminalProps) {
  const [streamCount, setStreamCount] = useState(0);
  const bottomRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    if (!stream) {
      return;
    }
    let intervalId: number | undefined;
    const timeoutId = window.setTimeout(() => {
      setStreamCount(0);
      let tick = 0;
      intervalId = window.setInterval(() => {
        tick += 1;
        setStreamCount(Math.min(tick, lines.length));
        if (tick >= lines.length && intervalId !== undefined) {
          window.clearInterval(intervalId);
        }
      }, intervalMs);
    }, 0);
    return () => {
      window.clearTimeout(timeoutId);
      if (intervalId !== undefined) {
        window.clearInterval(intervalId);
      }
    };
  }, [stream, intervalMs, lines.length]);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [stream, streamCount, lines.length]);

  const shown = stream ? lines.slice(0, streamCount) : lines;

  return (
    <div
      className="rounded-lg border border-gray-800 bg-[#0a0a0a] p-4 font-mono text-xs text-gray-200"
      role="log"
      aria-live="polite"
    >
      <div className="max-h-[min(420px,55vh)] space-y-1 overflow-y-auto">
        {shown.map((line, idx) => (
          <div key={`${line.timestamp}-${idx}`} className="flex gap-3">
            <span className="shrink-0 text-gray-500">{line.timestamp}</span>
            <span className={cn("shrink-0 uppercase", levelClass[line.level])}>
              {line.level}
            </span>
            <span className="min-w-0 break-words text-gray-100">
              {line.message}
            </span>
          </div>
        ))}
        <div ref={bottomRef} />
      </div>
    </div>
  );
}
