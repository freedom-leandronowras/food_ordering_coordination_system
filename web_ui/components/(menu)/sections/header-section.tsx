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
    <Card className="motion-enter rounded-2xl border-sl-dce9e4 bg-sl-ffffff/95 px-4 py-3 backdrop-blur">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <div className="flex h-9 w-9 items-center justify-center rounded-full bg-sl-1f6f64 text-sl-ffffff">
            #
          </div>
          <div>
            <p className="text-lg font-semibold leading-tight">{data.brandName}</p>
            <p className="text-xs text-sl-6e837c">{data.subtitle}</p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <div className="flex items-center gap-2 rounded-full border border-sl-dbe9e4 bg-sl-ffffff p-1">
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
              variant={viewMode === "history" ? "default" : "ghost"}
              onClick={() => setViewMode("history")}
            >
              {data.historyViewLabel}
            </Button>
            {isManager ? (
              <Button
                type="button"
                size="sm"
                variant={viewMode === "management" ? "default" : "ghost"}
                onClick={() => setViewMode("management")}
              >
                {data.managementViewLabel}
              </Button>
            ) : null}
          </div>

          <span className="rounded-full border border-sl-dbe9e4 bg-sl-edf7f3 px-3 py-1 text-xs font-semibold text-sl-245f55">
            Credits: {formatMoney(credits)}
          </span>
          <Button type="button" onClick={() => setShowGrantModal(true)} disabled={!memberId}>
            {data.addCreditsButtonLabel}
          </Button>
          <Button type="button" variant="ghost" onClick={signOut}>
            Sign out
          </Button>
        </div>
      </div>
    </Card>
  );
}
