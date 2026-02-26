"use client";

import { useMenuContext } from "@/components/(menu)/menu-context";
import type { MenuSectionsData } from "@/components/(menu)/menu-section-types";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { formatMoney } from "@/lib/menu-data";

type HeaderSectionProps = {
  data: MenuSectionsData["header"];
};

export function HeaderSection({ data }: HeaderSectionProps) {
  const { credits, memberId, isManager, viewMode, setViewMode, setShowGrantModal, signOut } =
    useMenuContext();

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
          {isManager ? (
            <div className="flex items-center gap-2 rounded-full border border-[#dbe9e4] bg-white p-1">
              <span className="px-2 text-xs text-[#617b74]">{data.viewLabel}</span>
              <Button
                type="button"
                size="sm"
                variant={viewMode === "menu" ? "default" : "ghost"}
                onClick={() => setViewMode("menu")}
              >
                {data.menuViewLabel}
              </Button>
              <Button
                type="button"
                size="sm"
                variant={viewMode === "management" ? "default" : "ghost"}
                onClick={() => setViewMode("management")}
              >
                {data.managementViewLabel}
              </Button>
            </div>
          ) : null}

          <span className="rounded-full border border-[#dbe9e4] bg-[#edf7f3] px-3 py-1 text-xs font-semibold text-[#245f55]">
            Credits: {formatMoney(credits)}
          </span>
          {isManager ? (
            <Button type="button" onClick={() => setShowGrantModal(true)} disabled={!memberId}>
              {data.addCreditsButtonLabel}
            </Button>
          ) : null}
          <Button type="button" variant="ghost" onClick={signOut}>
            Sign out
          </Button>
        </div>
      </div>
    </Card>
  );
}
