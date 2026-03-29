import { LandingAgents } from "@/components/features/landing-agents";
import { LandingFeatures } from "@/components/features/landing-features";
import { LandingFooter } from "@/components/features/landing-footer";
import { LandingHero } from "@/components/features/landing-hero";
import { LandingOss } from "@/components/features/landing-oss";
import { LandingSteps } from "@/components/features/landing-steps";

export default function HomePage() {
  return (
    <div className="min-h-screen bg-background">
      <LandingHero />
      <LandingSteps />
      <LandingFeatures />
      <LandingAgents />
      <LandingOss />
      <LandingFooter />
    </div>
  );
}
