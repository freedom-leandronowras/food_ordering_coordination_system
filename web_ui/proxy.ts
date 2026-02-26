import { NextResponse, type NextRequest } from "next/server";

import { sessionCookieName } from "@/lib/auth-session";

function isPublicRoute(pathname: string) {
  return pathname.startsWith("/auth");
}

export default function proxy(req: NextRequest) {
  if (isPublicRoute(req.nextUrl.pathname)) {
    return NextResponse.next();
  }

  const token = req.cookies.get(sessionCookieName)?.value;
  if (token) {
    return NextResponse.next();
  }

  const redirectUrl = req.nextUrl.clone();
  redirectUrl.pathname = "/auth";
  const relativePath = `${req.nextUrl.pathname}${req.nextUrl.search}`;
  if (relativePath.startsWith("/")) {
    redirectUrl.searchParams.set("redirect_url", relativePath);
  }
  return NextResponse.redirect(redirectUrl);
}

export const config = {
  matcher: [
    "/((?!_next|[^?]*\\.(?:html?|css|js(?!on)|jpe?g|webp|png|gif|svg|ttf|woff2?|ico|csv|docx?|xlsx?|zip|webmanifest)).*)",
    "/(api|trpc)(.*)",
  ],
};
