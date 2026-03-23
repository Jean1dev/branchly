import { LandingFooter } from "@/components/features/landing-footer";
import { LandingHero } from "@/components/features/landing-hero";
import { LandingSteps } from "@/components/features/landing-steps";

export default function HomePage() {
  return (
    <div className="min-h-screen bg-background">
      <LandingHero />
      <LandingSteps />
      <LandingFooter />
    </div>
  );
}
