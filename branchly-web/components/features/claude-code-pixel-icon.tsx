import { cn } from "@/lib/utils";

const CORAL: [number, number][] = [
  [3, 0],
  [7, 0],
  [2, 1],
  [8, 1],
  [1, 2],
  [3, 2],
  [5, 2],
  [7, 2],
  [9, 2],
  [0, 3],
  [3, 3],
  [7, 3],
  [10, 3],
  [0, 4],
  [10, 4],
  [1, 5],
  [2, 5],
  [3, 5],
  [4, 5],
  [5, 5],
  [6, 5],
  [7, 5],
  [8, 5],
  [9, 5],
  [2, 6],
  [8, 6],
  [1, 7],
  [4, 7],
  [6, 7],
  [9, 7],
];

const EYES: [number, number][] = [
  [3, 3],
  [7, 3],
];

export function ClaudeCodePixelIcon({ className }: { className?: string }) {
  return (
    <svg
      className={cn("shrink-0", className)}
      viewBox="-0.5 -0.5 11 8"
      aria-hidden
    >
      {CORAL.map(([x, y], i) => (
        <rect
          key={i}
          x={x}
          y={y}
          width={1}
          height={1}
          fill="#d37a5a"
          shapeRendering="crispEdges"
        />
      ))}
      {EYES.map(([x, y], i) => (
        <rect
          key={`e-${i}`}
          x={x}
          y={y}
          width={1}
          height={1}
          fill="#200510"
          shapeRendering="crispEdges"
        />
      ))}
    </svg>
  );
}
