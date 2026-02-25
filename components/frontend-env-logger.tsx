"use client";

import { useEffect } from "react";

import { frontendLogger } from "@/lib/frontend-logger";

type FrontendEnvVariable = {
  name: string;
  value: string | undefined;
  severity: "error" | "warn";
};

const frontendEnvVariables: FrontendEnvVariable[] = [
  {
    name: "NEXT_PUBLIC_API_BASE_URL",
    value: process.env.NEXT_PUBLIC_API_BASE_URL,
    severity: "error",
  },
  {
    name: "NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY",
    value: process.env.NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY,
    severity: "error",
  },
];

export function FrontendEnvLogger() {
  useEffect(() => {
    frontendEnvVariables.forEach((item) => {
      if (item.value) {
        return;
      }

      const details = {
        env: item.name,
        area: "frontend",
        message: "required frontend environment variable is missing",
      };
      if (item.severity === "error") {
        frontendLogger.error(details);
        return;
      }
      frontendLogger.warn(details);
    });
  }, []);

  return null;
}
