"use client";

import { UserButton } from "@clerk/nextjs";

import { useMenuContext } from "@/components/(menu)/menu-context";
import type { MenuSectionsData } from "@/components/(menu)/menu-section-types";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { formatMoney } from "@/lib/menu-data";

type HeaderSectionProps = {
  data: MenuSectionsData["header"];
};

export function HeaderSection({ data }: HeaderSectionProps) {
  const { credits, memberId, setShowGrantModal } = useMenuContext();

  return (
    <Card className="rounded-2xl border-[#dce9e4] bg-white/95 px-4 py-3 backdrop-blur">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <div className="flex h-9 w-9 items-center justify-center rounded-full bg-[#1f6f64] text-white">
            #
          </div>
          <div>
            <p className="text-lg font-semibold leading-tight">{data.brandName}</p>
            <p className="text-xs text-[#6e837c]">{data.subtitle}</p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <span className="rounded-full border border-[#dbe9e4] bg-[#edf7f3] px-3 py-1 text-xs font-semibold text-[#245f55]">
            Credits: {formatMoney(credits)}
          </span>
          <Button type="button" onClick={() => setShowGrantModal(true)} disabled={!memberId}>
            {data.addCreditsButtonLabel}
          </Button>
          <div className="rounded-full border border-[#dbe9e4] bg-white p-1">
            <UserButton afterSignOutUrl="/auth" />
          </div>
        </div>
      </div>
    </Card>
  );
}
