const managerRoles = new Set(["HIVE_MANAGER", "INNOVATION_LEAD"]);

export function normalizeRole(role: unknown) {
  if (typeof role !== "string") {
    return "";
  }
  return role.trim().toUpperCase();
}

export function isManagerRole(role: unknown) {
  return managerRoles.has(normalizeRole(role));
}

export function normalizeDomain(rawDomain: string) {
  return rawDomain.trim().toLowerCase().replace(/^@+/, "");
}

export function isValidDomain(domain: string) {
  const normalized = normalizeDomain(domain);
  // Keeps the validation simple while blocking obviously malformed values.
  return /^[a-z0-9](?:[a-z0-9-]*[a-z0-9])?(?:\.[a-z0-9](?:[a-z0-9-]*[a-z0-9])?)+$/.test(
    normalized,
  );
}

export function parseAllowedEmailDomains(rawValue: string | undefined) {
  if (!rawValue) {
    return [];
  }

  const uniqueDomains = new Set<string>();
  rawValue
    .split(",")
    .map((domain) => normalizeDomain(domain))
    .filter((domain) => domain.length > 0)
    .forEach((domain) => uniqueDomains.add(domain));

  return Array.from(uniqueDomains);
}

export function getEmailDomain(emailAddress: string) {
  const normalized = emailAddress.trim().toLowerCase();
  const atIndex = normalized.lastIndexOf("@");
  if (atIndex === -1 || atIndex === normalized.length - 1) {
    return "";
  }
  return normalized.slice(atIndex + 1);
}

export function isEmailAllowedForDomains(emailAddress: string, allowedDomains: string[]) {
  if (allowedDomains.length === 0) {
    return true;
  }

  const emailDomain = getEmailDomain(emailAddress);
  if (!emailDomain) {
    return false;
  }

  return allowedDomains.includes(emailDomain);
}
