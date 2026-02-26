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
    badge: "Trending today",
    title: "Spicy Tuna Poke Bowl",
    description:
      "Sign in to browse curated menus, reserve lunch, and manage team credits.",
    availabilityLabel: "Available until 11:30 AM",
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
