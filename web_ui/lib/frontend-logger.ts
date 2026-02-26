"use client";

import pino from "pino/browser";

export const frontendLogger = pino({
  name: "frontend",
  level: process.env.NODE_ENV === "production" ? "info" : "debug",
  browser: {
    asObject: true,
  },
});
