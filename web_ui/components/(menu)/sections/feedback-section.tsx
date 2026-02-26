"use client";

import { useMenuContext } from "@/components/(menu)/menu-context";
import { InlineFeedback } from "@/components/ui/inline-feedback";

export function FeedbackSection() {
  const { errorMessage, statusMessage } = useMenuContext();

  if (!errorMessage && !statusMessage) {
    return null;
  }

  return (
    <section className="space-y-2">
      {errorMessage ? <InlineFeedback message={errorMessage} tone="error" /> : null}
      {statusMessage ? <InlineFeedback message={statusMessage} tone="success" /> : null}
    </section>
  );
}
