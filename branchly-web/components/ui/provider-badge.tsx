import type { GitProvider } from "@/types";
import { ProviderLogo } from "./provider-logo";

interface ProviderBadgeProps {
  provider: GitProvider;
}

const providerLabel: Record<GitProvider, string> = {
  github: "GitHub",
  gitlab: "GitLab",
  "azure-devops": "Azure DevOps",
};

export function ProviderBadge({ provider }: ProviderBadgeProps) {
  return (
    <span className="inline-flex items-center gap-1">
      <ProviderLogo provider={provider} size={12} />
      <span style={{ fontSize: 12, lineHeight: "20px" }}>
        {providerLabel[provider]}
      </span>
    </span>
  );
}
