"use client";

import { SignIn, SignUp } from "@clerk/nextjs";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { createContext, useContext, type ReactNode } from "react";

import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { parseAllowedEmailDomains } from "@/lib/auth-policy";
import { cn } from "@/lib/utils";

export type AuthSectionsData = {
  brand: {
    name: string;
    tagline: string;
    versionLabel: string;
  };
  hero: {
    badge: string;
    title: string;
    description: string;
    availabilityLabel: string;
  };
  form: {
    signInLabel: string;
    signUpLabel: string;
    allowedDomainsHint: string;
    domainBlockedMessage: string;
  };
};

type AuthSectionsContextValue = {
  mode: "sign-in" | "sign-up";
  setMode: (mode: "sign-in" | "sign-up") => void;
  redirectUrl: string;
  allowedDomains: string[];
  authErrorMessage: string;
};

const AuthSectionsContext = createContext<AuthSectionsContextValue | null>(null);

function getRedirectUrl(rawRedirect: string | null) {
  if (!rawRedirect || !rawRedirect.startsWith("/")) {
    return "/";
  }
  return rawRedirect;
}

function getAuthErrorMessage(
  rawError: string | null,
  allowedDomains: string[],
  blockedDomainMessage: string,
) {
  if (rawError !== "EMAIL_DOMAIN_NOT_ALLOWED") {
    return "";
  }

  if (allowedDomains.length === 0) {
    return blockedDomainMessage;
  }

  const domainsText = allowedDomains.map((domain) => `@${domain}`).join(", ");
  return `${blockedDomainMessage} Allowed domains: ${domainsText}.`;
}

function useAuthSectionsContext() {
  const context = useContext(AuthSectionsContext);
  if (!context) {
    throw new Error("useAuthSectionsContext must be used inside AuthSections.");
  }
  return context;
}

function AuthSectionsProvider({
  children,
  data,
}: {
  children: ReactNode;
  data: AuthSectionsData;
}) {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();

  const mode = searchParams.get("mode") === "sign-up" ? "sign-up" : "sign-in";
  const redirectUrl = getRedirectUrl(searchParams.get("redirect_url"));
  const allowedDomains = parseAllowedEmailDomains(process.env.NEXT_PUBLIC_ALLOWED_EMAIL_DOMAINS);
  const authErrorMessage = getAuthErrorMessage(
    searchParams.get("error"),
    allowedDomains,
    data.form.domainBlockedMessage,
  );

  const setMode = (nextMode: "sign-in" | "sign-up") => {
    const params = new URLSearchParams(searchParams.toString());
    if (nextMode === "sign-in") {
      params.delete("mode");
    } else {
      params.set("mode", "sign-up");
    }
    const nextQuery = params.toString();
    router.replace(nextQuery ? `${pathname}?${nextQuery}` : pathname);
  };

  return (
    <AuthSectionsContext.Provider value={{ mode, setMode, redirectUrl, allowedDomains, authErrorMessage }}>
      {children}
    </AuthSectionsContext.Provider>
  );
}

function AuthHeroSection({ data }: { data: AuthSectionsData }) {
  return (
    <section className="relative hidden w-[52%] overflow-hidden bg-gradient-to-br from-[#d7ebe6] via-[#cce4de] to-[#bdd9d2] p-10 lg:flex lg:flex-col lg:justify-between">
      <div className="flex items-center gap-3">
        <div className="flex h-10 w-10 items-center justify-center rounded-full bg-[#1f6f64] text-white">
          #
        </div>
        <div>
          <p className="text-xl font-semibold text-[#123830]">{data.brand.name}</p>
          <p className="text-sm text-[#4d6e66]">{data.brand.tagline}</p>
        </div>
      </div>

      <Card className="mx-auto w-full max-w-[540px] rounded-[36px] border-0 bg-gradient-to-br from-[#1a5148] via-[#23665a] to-[#2b7367] p-8 text-white shadow-2xl">
        <p className="text-xs uppercase tracking-[0.18em] text-[#c8e6dd]">{data.hero.badge}</p>
        <h2 className="mt-4 text-4xl font-semibold leading-tight">{data.hero.title}</h2>
        <p className="mt-3 max-w-sm text-sm text-[#d9eee8]">{data.hero.description}</p>
        <div className="mt-8 inline-flex items-center rounded-full bg-white/15 px-3 py-1 text-xs">
          {data.hero.availabilityLabel}
        </div>
      </Card>

      <p className="text-xs text-[#5e7c75]">{data.brand.versionLabel}</p>
    </section>
  );
}

function AuthFormSection({ data }: { data: AuthSectionsData }) {
  const { mode, setMode, redirectUrl, allowedDomains, authErrorMessage } = useAuthSectionsContext();

  const clerkAppearance = {
    elements: {
      card: "shadow-none border-0 bg-transparent p-0",
      headerTitle: "text-2xl font-semibold text-[#123830]",
      headerSubtitle: "text-[#5e7871]",
      socialButtonsBlockButton:
        "rounded-full border border-[#dce9e5] bg-white text-[#123830] hover:bg-[#f4f8f6]",
      socialButtonsBlockButtonText: "font-medium",
      dividerLine: "bg-[#dce9e5]",
      dividerText: "text-[#7a908a]",
      formFieldInput:
        "rounded-full border border-[#dbe8e3] bg-[#f8fbfa] text-[#123830] focus:border-[#1f6f64] focus:ring-[#1f6f64]",
      formButtonPrimary:
        "rounded-full bg-[#1f6f64] hover:bg-[#17564d] text-white font-semibold",
      footerActionLink: "text-[#1f6f64] hover:text-[#17564d]",
    },
  } as const;

  return (
    <section className="w-full bg-white p-5 md:p-10 lg:w-[48%] lg:p-12">
      <div className="mx-auto w-full max-w-md">
        <div className="mb-6 flex items-center gap-3 lg:hidden">
          <div className="flex h-9 w-9 items-center justify-center rounded-full bg-[#1f6f64] text-white">
            #
          </div>
          <div>
            <p className="text-lg font-semibold text-[#123830]">{data.brand.name}</p>
            <p className="text-xs text-[#5d7871]">{data.brand.tagline}</p>
          </div>
        </div>

        <div className="mb-5 flex rounded-full bg-[#f2f7f5] p-1">
          <Button
            type="button"
            onClick={() => setMode("sign-in")}
            variant={mode === "sign-in" ? "default" : "ghost"}
            className={cn("flex-1", mode === "sign-in" ? "" : "text-[#5e7871]")}
          >
            {data.form.signInLabel}
          </Button>
          <Button
            type="button"
            onClick={() => setMode("sign-up")}
            variant={mode === "sign-up" ? "default" : "ghost"}
            className={cn("flex-1", mode === "sign-up" ? "" : "text-[#5e7871]")}
          >
            {data.form.signUpLabel}
          </Button>
        </div>

        <Card className="rounded-3xl border-[#dce9e5] bg-[#fbfdfc] p-5">
          {authErrorMessage ? (
            <Card className="mb-4 rounded-2xl border-[#f0d1cf] bg-[#fff6f5] p-3 text-sm text-[#8f352c]">
              {authErrorMessage}
            </Card>
          ) : null}

          {mode === "sign-up" && allowedDomains.length > 0 ? (
            <Card className="mb-4 rounded-2xl border-[#dce9e5] bg-[#eef7f4] p-3 text-xs text-[#4e6f66]">
              {data.form.allowedDomainsHint} {allowedDomains.map((domain) => `@${domain}`).join(", ")}.
            </Card>
          ) : null}

          {mode === "sign-in" ? (
            <SignIn
              routing="virtual"
              signUpUrl="/auth?mode=sign-up"
              fallbackRedirectUrl={redirectUrl}
              appearance={clerkAppearance}
            />
          ) : (
            <SignUp
              routing="virtual"
              signInUrl="/auth"
              fallbackRedirectUrl={redirectUrl}
              appearance={clerkAppearance}
            />
          )}
        </Card>
      </div>
    </section>
  );
}

function AuthPageSections({ data }: { data: AuthSectionsData }) {
  return (
    <main className="min-h-screen bg-[#edf4f1] p-4 md:p-8">
      <div className="mx-auto flex min-h-[calc(100vh-2rem)] max-w-[1280px] overflow-hidden rounded-[32px] border border-[#d7e7e1] bg-white shadow-xl">
        <AuthHeroSection data={data} />
        <AuthFormSection data={data} />
      </div>
    </main>
  );
}

export function AuthSections({ data }: { data: AuthSectionsData }) {
  return (
    <AuthSectionsProvider data={data}>
      <AuthPageSections data={data} />
    </AuthSectionsProvider>
  );
}
