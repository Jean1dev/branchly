import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import type { ReactNode } from "react";

type EmptyStateProps = {
  title: string;
  description: string;
  action?: { label: string; href: string };
  icon?: ReactNode;
};

export function EmptyState({
  title,
  description,
  action,
  icon,
}: EmptyStateProps) {
  return (
    <Card className="flex flex-col items-center justify-center py-16 text-center">
      {icon ? <div className="mb-4 text-gray-400">{icon}</div> : null}
      <h2 className="text-lg font-semibold">{title}</h2>
      <p className="mt-2 max-w-sm text-sm text-gray-500 dark:text-gray-400">
        {description}
      </p>
      {action ? (
        <Button className="mt-6" href={action.href}>
          {action.label}
        </Button>
      ) : null}
    </Card>
  );
}
