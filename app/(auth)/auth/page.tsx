import {
  AuthSections,
  type AuthSectionsData,
} from "@/components/(auth)/auth-sections";

const authSectionsData: AuthSectionsData = {
  brand: {
    name: "SoftLunch",
    tagline: "Fuel for your best work",
    versionLabel: "Internal Platform v2.4",
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
  },
};

export default function AuthPage() {
  return <AuthSections data={authSectionsData} />;
}
