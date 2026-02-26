export const sessionCookieName = "focs_session";

export type SessionJwtClaims = {
  sub?: string;
  exp?: number;
  role?: string;
  email?: string;
  user_id?: string;
};

function decodeBase64Url(value: string) {
  const normalized = value.replace(/-/g, "+").replace(/_/g, "/");
  const padded = normalized.padEnd(Math.ceil(normalized.length / 4) * 4, "=");
  return atob(padded);
}

export function parseJwtClaims(token: string): SessionJwtClaims | null {
  const segments = token.split(".");
  if (segments.length < 2) {
    return null;
  }

  try {
    const decoded = decodeBase64Url(segments[1]);
    return JSON.parse(decoded) as SessionJwtClaims;
  } catch {
    return null;
  }
}

export function readSessionToken() {
  if (typeof document === "undefined") {
    return "";
  }

  const entries = document.cookie.split(";").map((entry) => entry.trim());
  for (const entry of entries) {
    if (!entry.startsWith(`${sessionCookieName}=`)) {
      continue;
    }
    return decodeURIComponent(entry.slice(sessionCookieName.length + 1));
  }

  return "";
}

export function writeSessionToken(token: string, maxAgeSeconds = 60 * 60 * 24 * 7) {
  if (typeof document === "undefined") {
    return;
  }

  const secure = typeof window !== "undefined" && window.location.protocol === "https:";
  const securePart = secure ? "; Secure" : "";
  document.cookie = `${sessionCookieName}=${encodeURIComponent(token)}; Max-Age=${Math.max(0, Math.floor(maxAgeSeconds))}; Path=/; SameSite=Lax${securePart}`;
}

export function clearSessionToken() {
  if (typeof document === "undefined") {
    return;
  }

  const secure = typeof window !== "undefined" && window.location.protocol === "https:";
  const securePart = secure ? "; Secure" : "";
  document.cookie = `${sessionCookieName}=; Max-Age=0; Path=/; SameSite=Lax${securePart}`;
}
