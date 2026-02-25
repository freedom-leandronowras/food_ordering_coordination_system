"use client";

import { SignIn, SignUp } from "@clerk/nextjs";
import { usePathname, useRouter, useSearchParams } from "next/navigation";

function getRedirectUrl(rawRedirect: string | null) {
  if (!rawRedirect || !rawRedirect.startsWith("/")) {
    return "/";
  }
  return rawRedirect;
}

export default function AuthPage() {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();

  const mode = searchParams.get("mode") === "sign-up" ? "sign-up" : "sign-in";
  const redirectUrl = getRedirectUrl(searchParams.get("redirect_url"));

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
    <main className="min-h-screen bg-[#edf4f1] p-4 md:p-8">
      <div className="mx-auto flex min-h-[calc(100vh-2rem)] max-w-[1280px] overflow-hidden rounded-[32px] border border-[#d7e7e1] bg-white shadow-xl">
        <section className="relative hidden w-[52%] overflow-hidden bg-gradient-to-br from-[#d7ebe6] via-[#cce4de] to-[#bdd9d2] p-10 lg:flex lg:flex-col lg:justify-between">
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-full bg-[#1f6f64] text-white">
              #
            </div>
            <div>
              <p className="text-xl font-semibold text-[#123830]">SoftLunch</p>
              <p className="text-sm text-[#4d6e66]">Fuel for your best work</p>
            </div>
          </div>

          <div className="mx-auto w-full max-w-[540px] rounded-[36px] bg-gradient-to-br from-[#1a5148] via-[#23665a] to-[#2b7367] p-8 text-white shadow-2xl">
            <p className="text-xs uppercase tracking-[0.18em] text-[#c8e6dd]">Trending today</p>
            <h2 className="mt-4 text-4xl font-semibold leading-tight">Spicy Tuna Poke Bowl</h2>
            <p className="mt-3 max-w-sm text-sm text-[#d9eee8]">
              Sign in to browse curated menus, reserve lunch, and manage team credits.
            </p>
            <div className="mt-8 inline-flex items-center rounded-full bg-white/15 px-3 py-1 text-xs">
              Available until 11:30 AM
            </div>
          </div>

          <p className="text-xs text-[#5e7c75]">Internal Platform v2.4</p>
        </section>

        <section className="w-full bg-white p-5 md:p-10 lg:w-[48%] lg:p-12">
          <div className="mx-auto w-full max-w-md">
            <div className="mb-6 flex items-center gap-3 lg:hidden">
              <div className="flex h-9 w-9 items-center justify-center rounded-full bg-[#1f6f64] text-white">
                #
              </div>
              <div>
                <p className="text-lg font-semibold text-[#123830]">SoftLunch</p>
                <p className="text-xs text-[#5d7871]">Fuel for your best work</p>
              </div>
            </div>

            <div className="mb-5 flex rounded-full bg-[#f2f7f5] p-1">
              <button
                type="button"
                onClick={() => setMode("sign-in")}
                className={`flex-1 rounded-full px-4 py-2 text-sm ${
                  mode === "sign-in"
                    ? "bg-[#1f6f64] font-semibold text-white"
                    : "text-[#5e7871]"
                }`}
              >
                Sign in
              </button>
              <button
                type="button"
                onClick={() => setMode("sign-up")}
                className={`flex-1 rounded-full px-4 py-2 text-sm ${
                  mode === "sign-up"
                    ? "bg-[#1f6f64] font-semibold text-white"
                    : "text-[#5e7871]"
                }`}
              >
                Sign up
              </button>
            </div>

            <div className="rounded-3xl border border-[#dce9e5] bg-[#fbfdfc] p-5">
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
            </div>
          </div>
        </section>
      </div>
    </main>
  );
}
