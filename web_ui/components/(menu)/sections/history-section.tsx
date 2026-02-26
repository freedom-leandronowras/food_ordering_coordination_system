"use client";

import { useMemo } from "react";

import { useMenuContext, type OrderHistoryEntry } from "@/components/(menu)/menu-context";
import type { MenuSectionsData } from "@/components/(menu)/menu-section-types";
import { Card } from "@/components/ui/card";
import { formatMoney } from "@/lib/menu-data";

type HistorySectionProps = {
  data: MenuSectionsData["history"];
};

type GroupedHistory = {
  memberId: string;
  memberLabel: string;
  entries: OrderHistoryEntry[];
};

export function HistorySection({ data }: HistorySectionProps) {
  const { isManager, historyLoading, orderHistory } = useMenuContext();

  const groupedHistory = useMemo(() => {
    const groups = new Map<string, GroupedHistory>();
    orderHistory.forEach((entry) => {
      const existing = groups.get(entry.memberId);
      if (existing) {
        existing.entries.push(entry);
        return;
      }

      const label = entry.memberName || entry.memberEmail || entry.memberId;
      groups.set(entry.memberId, {
        memberId: entry.memberId,
        memberLabel: label,
        entries: [entry],
      });
    });
    return Array.from(groups.values());
  }, [orderHistory]);

  if (historyLoading) {
    return <Card className="motion-enter rounded-3xl p-5 text-sm text-sl-607b74">{data.loadingLabel}</Card>;
  }

  if (groupedHistory.length === 0) {
    return (
      <Card className="motion-enter rounded-3xl p-5 text-sm text-sl-607b74">
        {isManager ? data.emptyManagerLabel : data.emptyMemberLabel}
      </Card>
    );
  }

  return (
    <div className="space-y-4 motion-enter">
      <Card className="rounded-3xl border-sl-dce9e5 bg-sl-ffffff/95 p-5">
        <h3 className="text-2xl font-semibold text-sl-1a4d45">{data.title}</h3>
        <p className="mt-1 text-sm text-sl-4f6f67">{isManager ? data.managerDescription : data.memberDescription}</p>
      </Card>

      {groupedHistory.map((group) => (
        <Card key={group.memberId} className="rounded-3xl p-4">
          <div className="mb-3">
            <p className="text-base font-semibold text-sl-1a4d45">{group.memberLabel}</p>
            <p className="text-xs text-sl-4f6f67">
              {data.memberIdLabel}: {group.memberId}
            </p>
          </div>

          <div className="space-y-3">
            {group.entries.map(({ order }) => (
              <Card key={`${group.memberId}-${order.order_id}`} className="rounded-2xl border-sl-e0ebe7 bg-sl-f9fcfb p-3">
                <div className="flex flex-wrap items-start justify-between gap-2">
                  <div>
                    <p className="text-sm font-semibold text-sl-1a4d45">
                      {data.orderIdLabel}: {order.order_id}
                    </p>
                    <p className="text-xs text-sl-4f6f67">
                      {data.statusLabel}: {order.status}
                    </p>
                  </div>
                  <p className="text-sm font-semibold text-sl-1c5f54">
                    {data.totalLabel}: {formatMoney(Number(order.total_price ?? 0))}
                  </p>
                </div>

                <p className="mt-2 text-xs text-sl-4f6f67">
                  {data.notesLabel}: {order.delivery_notes?.trim() ? order.delivery_notes : data.noNotesLabel}
                </p>

                <p className="mt-2 text-xs text-sl-4f6f67">
                  {data.itemsLabel}:{" "}
                  {order.items.length > 0
                    ? order.items.map((item) => `${item.quantity}x ${item.name}`).join(", ")
                    : "-"}
                </p>
              </Card>
            ))}
          </div>
        </Card>
      ))}
    </div>
  );
}
