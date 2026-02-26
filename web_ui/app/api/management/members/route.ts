import { auth, clerkClient } from "@clerk/nextjs/server";
import { NextResponse } from "next/server";

import {
  getEmailDomain,
  isManagerRole,
  isValidDomain,
  normalizeDomain,
  normalizeRole,
  parseAllowedEmailDomains,
} from "@/lib/auth-policy";
import type { MembersByDomainResponse } from "@/lib/menu-data";

const uuidPattern =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
const membersPageSize = 100;
const maxUsersToScan = 500;

type ClerkEmailAddress = {
  id: string;
  emailAddress: string;
};

type ClerkUserPayload = {
  id: string;
  externalId: string | null;
  firstName: string | null;
  lastName: string | null;
  primaryEmailAddressId: string | null;
  emailAddresses: ClerkEmailAddress[];
  publicMetadata: unknown;
  unsafeMetadata: unknown;
};

type MetadataRecord = Record<string, unknown>;

function asMetadataRecord(input: unknown): MetadataRecord {
  if (!input || typeof input !== "object") {
    return {};
  }
  return input as MetadataRecord;
}

function readStringField(record: MetadataRecord, key: string) {
  const value = record[key];
  if (typeof value !== "string") {
    return "";
  }
  return value.trim();
}

function resolveMemberId(user: ClerkUserPayload) {
  const publicMetadata = asMetadataRecord(user.publicMetadata);
  const unsafeMetadata = asMetadataRecord(user.unsafeMetadata);

  const candidates = [
    readStringField(publicMetadata, "memberId"),
    readStringField(publicMetadata, "member_id"),
    readStringField(unsafeMetadata, "memberId"),
    readStringField(unsafeMetadata, "member_id"),
    typeof user.externalId === "string" ? user.externalId.trim() : "",
    user.id.trim(),
  ];

  return candidates.find((candidate) => candidate && uuidPattern.test(candidate)) ?? "";
}

function resolvePrimaryEmail(user: ClerkUserPayload) {
  if (!user.emailAddresses.length) {
    return "";
  }

  const selectedById = user.emailAddresses.find(
    (emailAddress) => emailAddress.id === user.primaryEmailAddressId,
  );
  const selected = selectedById ?? user.emailAddresses[0];
  return selected.emailAddress.trim().toLowerCase();
}

function resolveDisplayName(user: ClerkUserPayload) {
  const firstName = user.firstName?.trim() ?? "";
  const lastName = user.lastName?.trim() ?? "";
  const fullName = `${firstName} ${lastName}`.trim();
  if (fullName) {
    return fullName;
  }
  return "Unknown member";
}

function jsonError(status: number, code: string, message: string) {
  return NextResponse.json({ code, message }, { status });
}

export const runtime = "nodejs";

export async function GET(request: Request) {
  const { userId, sessionClaims } = await auth();
  if (!userId) {
    return jsonError(httpStatus.unauthorized, "UNAUTHORIZED", "authentication is required");
  }

  const claimsRecord = asMetadataRecord(sessionClaims);
  const role = normalizeRole(claimsRecord.role);
  if (!isManagerRole(role)) {
    return jsonError(httpStatus.forbidden, "FORBIDDEN", "manager role is required");
  }

  const { searchParams } = new URL(request.url);
  const domainParam = searchParams.get("domain") ?? "";
  const domain = normalizeDomain(domainParam);
  if (!domain || !isValidDomain(domain)) {
    return jsonError(httpStatus.badRequest, "INVALID_DOMAIN", "domain query is invalid");
  }

  const allowedDomains = parseAllowedEmailDomains(process.env.NEXT_PUBLIC_ALLOWED_EMAIL_DOMAINS);
  if (allowedDomains.length > 0 && !allowedDomains.includes(domain)) {
    return jsonError(httpStatus.forbidden, "DOMAIN_NOT_ALLOWED", "domain is not in the allowed list");
  }

  const clerk = await clerkClient();
  const responsePayload: MembersByDomainResponse = { domain, members: [] };

  let offset = 0;
  let scannedUsers = 0;
  while (scannedUsers < maxUsersToScan) {
    const page = await clerk.users.getUserList({ limit: membersPageSize, offset });
    const users = page.data as unknown as ClerkUserPayload[];

    if (users.length === 0) {
      break;
    }

    scannedUsers += users.length;
    offset += users.length;

    users.forEach((user) => {
      const email = resolvePrimaryEmail(user);
      if (!email || getEmailDomain(email) !== domain) {
        return;
      }

      const memberId = resolveMemberId(user);
      if (!memberId) {
        return;
      }

      responsePayload.members.push({
        user_id: user.id,
        member_id: memberId,
        email,
        full_name: resolveDisplayName(user),
      });
    });

    if (users.length < membersPageSize) {
      break;
    }
  }

  responsePayload.members.sort((a, b) => a.email.localeCompare(b.email));
  return NextResponse.json(responsePayload, { status: httpStatus.ok });
}

const httpStatus = {
  ok: 200,
  badRequest: 400,
  unauthorized: 401,
  forbidden: 403,
} as const;
