"use client";

import { useEffect } from "react";

import { MenuStateCard } from "@/components/(menu)/menu-state-card";
import { frontendLogger } from "@/lib/frontend-logger";

export function MenuMissingConfigState() {
  useEffect(() => {
    frontendLogger.error({
      env: "NEXT_PUBLIC_API_BASE_URL",
      area: "menu",
      message: "missing required frontend environment variable",
    });
  }, []);

  return <MenuStateCard message="Missing `NEXT_PUBLIC_API_BASE_URL`. Check frontend environment settings." />;
}
