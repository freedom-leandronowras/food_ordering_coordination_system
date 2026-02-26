"use client";

import { useMenuContext } from "@/components/(menu)/menu-context";
import { Card } from "@/components/ui/card";

export function FeedbackSection() {
  const { errorMessage, statusMessage } = useMenuContext();

  if (!errorMessage && !statusMessage) {
    return null;
  }

  return (
    <section className="fixed bottom-24 left-4 right-4 z-30 mx-auto max-w-[1240px] lg:bottom-8">
      <Card className="p-4">
        {errorMessage && <p className="text-sm font-medium text-[#b33e2d]">{errorMessage}</p>}
        {statusMessage && <p className="text-sm font-medium text-[#245f55]">{statusMessage}</p>}
      </Card>
    </section>
  );
}
