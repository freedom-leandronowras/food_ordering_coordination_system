"use client";

import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { createContext, useContext, useMemo, useState, type FormEvent, type ReactNode } from "react";

import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { InlineFeedback } from "@/components/ui/inline-feedback";
import { isEmailAllowedForDomains, parseAllowedEmailDomains } from "@/lib/auth-policy";
import { parseJwtClaims, writeSessionToken } from "@/lib/auth-session";
import { getApiErrorMessage } from "@/lib/menu-data";
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

type AuthSessionPayload = {
  token: string;
  user: {
    email: string;
  };
};

const AuthSectionsContext = createContext<AuthSectionsContextValue | null>(null);

function getRedirectUrl(rawRedirect: string | null) {
  if (!rawRedirect || !rawRedirect.startsWith("/")) {
    return "/";
  }
  return rawRedirect;
}

function getAuthErrorMessage(rawError: string | null, allowedDomains: string[], blockedDomainMessage: string) {
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

async function requestAuth<T,>(apiBaseUrl: string, path: string, body: unknown): Promise<T> {
  const response = await fetch(`${apiBaseUrl}${path}`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(body),
  });

  const text = await response.text();
  let payload: unknown = null;
  if (text) {
    try {
      payload = JSON.parse(text) as unknown;
    } catch {
      payload = text;
    }
  }

  if (!response.ok) {
    throw new Error(getApiErrorMessage(payload));
  }

  return payload as T;
}

function AuthSectionsProvider({ children, data }: { children: ReactNode; data: AuthSectionsData }) {
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
    <section className="relative hidden w-[52%] overflow-hidden bg-gradient-to-br from-sl-d7ebe6 via-sl-cce4de to-sl-bdd9d2 p-10 lg:flex lg:flex-col lg:justify-between">
      <div className="flex items-center gap-3">
        <div className="flex h-10 w-10 items-center justify-center rounded-full bg-sl-1f6f64 text-sl-ffffff">#</div>
        <div>
          <p className="text-xl font-semibold text-sl-123830">{data.brand.name}</p>
          <p className="text-sm text-sl-4d6e66">{data.brand.tagline}</p>
        </div>
      </div>

      <Card className="mx-auto w-full max-w-[540px] rounded-[36px] border-0 bg-gradient-to-br from-sl-1a5148 via-sl-23665a to-sl-2b7367 p-8 text-sl-ffffff shadow-2xl">
        <p className="text-xs uppercase tracking-[0.18em] text-sl-c8e6dd">{data.hero.badge}</p>
        <h2 className="mt-4 text-4xl font-semibold leading-tight">{data.hero.title}</h2>
        <p className="mt-3 max-w-sm text-sm text-sl-d9eee8">{data.hero.description}</p>
        <div className="mt-8 inline-flex items-center rounded-full bg-sl-ffffff/15 px-3 py-1 text-xs">
          {data.hero.availabilityLabel}
        </div>
      </Card>

      {data.brand.versionLabel ? <p className="text-xs text-sl-5e7c75">{data.brand.versionLabel}</p> : null}
    </section>
  );
}

function AuthFormSection({ data }: { data: AuthSectionsData }) {
  const router = useRouter();
  const { mode, setMode, redirectUrl, allowedDomains, authErrorMessage } = useAuthSectionsContext();
  const apiBaseUrl = process.env.NEXT_PUBLIC_API_BASE_URL ?? "";

  const [signInEmail, setSignInEmail] = useState("");
  const [signInPassword, setSignInPassword] = useState("");

  const [signUpName, setSignUpName] = useState("");
  const [signUpEmail, setSignUpEmail] = useState("");
  const [signUpPassword, setSignUpPassword] = useState("");
  const [signUpRole, setSignUpRole] = useState("MEMBER");

  const [status, setStatus] = useState("");
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const authBlockedMessage = useMemo(() => {
    if (!apiBaseUrl) {
      return "Missing NEXT_PUBLIC_API_BASE_URL. Configure the web environment.";
    }
    return "";
  }, [apiBaseUrl]);

  const completeAuth = (payload: AuthSessionPayload) => {
    if (!payload.token) {
      setError("Authentication response is missing the session token.");
      return;
    }

    if (
      allowedDomains.length > 0 &&
      payload.user?.email &&
      !isEmailAllowedForDomains(payload.user.email.toLowerCase(), allowedDomains)
    ) {
      setError(data.form.domainBlockedMessage);
      return;
    }

    const claims = parseJwtClaims(payload.token);
    if (!claims?.sub) {
      setError("Session token is invalid.");
      return;
    }

    writeSessionToken(payload.token);
    router.replace(redirectUrl || "/");
  };

  const onSignIn = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setStatus("");
    setError("");

    if (!apiBaseUrl) {
      setError("Missing NEXT_PUBLIC_API_BASE_URL.");
      return;
    }

    setSubmitting(true);
    try {
      const payload = await requestAuth<AuthSessionPayload>(apiBaseUrl, "/api/auth/login", {
        email: signInEmail,
        password: signInPassword,
      });
      completeAuth(payload);
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "Could not sign in.");
    } finally {
      setSubmitting(false);
    }
  };

  const onSignUp = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setStatus("");
    setError("");

    if (!apiBaseUrl) {
      setError("Missing NEXT_PUBLIC_API_BASE_URL.");
      return;
    }

    if (
      allowedDomains.length > 0 &&
      signUpEmail &&
      !isEmailAllowedForDomains(signUpEmail.toLowerCase(), allowedDomains)
    ) {
      setError(data.form.domainBlockedMessage);
      return;
    }

    setSubmitting(true);
    try {
      const payload = await requestAuth<AuthSessionPayload>(apiBaseUrl, "/api/auth/register", {
        email: signUpEmail,
        password: signUpPassword,
        full_name: signUpName,
        role: signUpRole,
      });
      setStatus("Account created. Redirecting...");
      completeAuth(payload);
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "Could not create account.");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <section className="w-full bg-sl-ffffff p-5 md:p-10 lg:w-[48%] lg:p-12">
      <div className="mx-auto w-full max-w-md">
        <div className="mb-6 flex items-center gap-3 lg:hidden">
          <div className="flex h-9 w-9 items-center justify-center rounded-full bg-sl-1f6f64 text-sl-ffffff">#</div>
          <div>
            <p className="text-lg font-semibold text-sl-123830">{data.brand.name}</p>
            <p className="text-xs text-sl-5d7871">{data.brand.tagline}</p>
          </div>
        </div>

        <div className="mb-5 flex rounded-full bg-sl-f2f7f5 p-1">
          <Button
            type="button"
            onClick={() => setMode("sign-in")}
            variant={mode === "sign-in" ? "default" : "ghost"}
            className={cn("flex-1", mode === "sign-in" ? "" : "text-sl-5e7871")}
          >
            {data.form.signInLabel}
          </Button>
          <Button
            type="button"
            onClick={() => setMode("sign-up")}
            variant={mode === "sign-up" ? "default" : "ghost"}
            className={cn("flex-1", mode === "sign-up" ? "" : "text-sl-5e7871")}
          >
            {data.form.signUpLabel}
          </Button>
        </div>

        <Card className="rounded-3xl border-sl-dce9e5 bg-sl-fbfdfc p-5">
          {authErrorMessage ? (
            <InlineFeedback message={authErrorMessage} tone="error" className="mb-4" />
          ) : null}

          {authBlockedMessage ? (
            <InlineFeedback message={authBlockedMessage} tone="error" className="mb-4" />
          ) : null}

          {mode === "sign-up" && allowedDomains.length > 0 ? (
            <Card className="mb-4 rounded-2xl border-sl-dce9e5 bg-sl-eef7f4 p-3 text-xs text-sl-4e6f66">
              {data.form.allowedDomainsHint} {allowedDomains.map((domain) => `@${domain}`).join(", ")}.
            </Card>
          ) : null}

          {status ? <InlineFeedback message={status} tone="success" className="mb-3" /> : null}
          {error ? <InlineFeedback message={error} tone="error" className="mb-3" /> : null}

          {mode === "sign-in" ? (
            <form className="space-y-3" onSubmit={onSignIn}>
              <label className="block text-sm font-medium text-sl-305a52">
                Email
                <Input
                  className="mt-1"
                  type="email"
                  required
                  autoComplete="email"
                  value={signInEmail}
                  onChange={(event) => setSignInEmail(event.target.value)}
                />
              </label>

              <label className="block text-sm font-medium text-sl-305a52">
                Password
                <Input
                  className="mt-1"
                  type="password"
                  required
                  autoComplete="current-password"
                  value={signInPassword}
                  onChange={(event) => setSignInPassword(event.target.value)}
                />
              </label>

              <Button type="submit" disabled={submitting || !apiBaseUrl} className="w-full">
                {submitting ? "Signing in..." : data.form.signInLabel}
              </Button>
            </form>
          ) : (
            <form className="space-y-3" onSubmit={onSignUp}>
              <label className="block text-sm font-medium text-sl-305a52">
                Full name
                <Input
                  className="mt-1"
                  required
                  autoComplete="name"
                  value={signUpName}
                  onChange={(event) => setSignUpName(event.target.value)}
                />
              </label>

              <label className="block text-sm font-medium text-sl-305a52">
                Email
                <Input
                  className="mt-1"
                  type="email"
                  required
                  autoComplete="email"
                  value={signUpEmail}
                  onChange={(event) => setSignUpEmail(event.target.value)}
                />
              </label>

              <label className="block text-sm font-medium text-sl-305a52">
                Password
                <Input
                  className="mt-1"
                  type="password"
                  required
                  minLength={8}
                  autoComplete="new-password"
                  value={signUpPassword}
                  onChange={(event) => setSignUpPassword(event.target.value)}
                />
              </label>

              <label className="block text-sm font-medium text-sl-305a52">
                Role
                <select
                  className="mt-1 w-full rounded-md border border-sl-dbe8e3 bg-sl-f8fbfa px-3 py-2 text-sm text-sl-123830"
                  value={signUpRole}
                  onChange={(event) => setSignUpRole(event.target.value)}
                >
                  <option value="MEMBER">Member</option>
                  <option value="HIVE_MANAGER">Hive Manager</option>
                </select>
              </label>

              <Button type="submit" disabled={submitting || !apiBaseUrl} className="w-full">
                {submitting ? "Creating account..." : data.form.signUpLabel}
              </Button>
            </form>
          )}
        </Card>
      </div>
    </section>
  );
}

function AuthPageSections({ data }: { data: AuthSectionsData }) {
  return (
    <main className="min-h-screen bg-sl-edf4f1 p-4 md:p-8">
      <div className="mx-auto flex min-h-[calc(100vh-2rem)] max-w-[1280px] overflow-hidden rounded-[32px] border border-sl-d7e7e1 bg-sl-ffffff shadow-xl">
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
