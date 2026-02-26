import { Suspense } from "react";

import {
  AuthSections,
  type AuthSectionsData,
} from "@/components/(auth)/auth-sections";

const authSectionsData: AuthSectionsData = {
  brand: {
    name: "SoftLunch",
    tagline: "Fuel for your best work",
    versionLabel: "",
  },
  hero: {
    badge: "Food ordering coordination",
    title: "One place for office lunch operations",
    description:
      "Coordinate vendors, place orders, and manage member credits with a simple workflow for both members and managers.",
    availabilityLabel: "Fast, clear, and team-friendly",
  },
  form: {
    signInLabel: "Sign in",
    signUpLabel: "Sign up",
    allowedDomainsHint: "Sign-ups are restricted to:",
    domainBlockedMessage:
      "Your account email domain is not allowed for this environment.",
  },
};

export default function AuthPage() {
  return (
    <Suspense fallback={null}>
      <AuthSections data={authSectionsData} />
    </Suspense>
  );
}
