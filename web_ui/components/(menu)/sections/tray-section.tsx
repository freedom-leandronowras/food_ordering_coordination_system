"use client";

import { ReservationCard } from "@/components/(menu)/reservation-card";
import { useMenuContext } from "@/components/(menu)/menu-context";
import type { MenuSectionsData } from "@/components/(menu)/menu-section-types";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { formatMoney } from "@/lib/menu-data";

type TraySectionProps = {
  data: MenuSectionsData["tray"];
};

export function TraySection({ data }: TraySectionProps) {
  const {
    cartLines,
    subtotal,
    creditsApplied,
    totalPay,
    canReserve,
    placingOrder,
    setShowConfirmModal,
    updateCartQuantity,
  } = useMenuContext();

  return (
    <Card className="motion-enter-delayed rounded-3xl p-4">
      <CardHeader className="mb-2 p-0">
        <div className="flex items-center justify-between">
          <CardTitle className="text-2xl">{data.title}</CardTitle>
          <span className="rounded-full bg-sl-eef6f3 px-2 py-1 text-xs text-sl-295b52">
            {cartLines.length} item{cartLines.length === 1 ? "" : "s"}
          </span>
        </div>
      </CardHeader>

      <CardContent className="space-y-3 p-0">
        {cartLines.length === 0 ? (
          <p className="rounded-2xl bg-sl-f4f8f6 px-3 py-6 text-center text-sm text-sl-6b847d">
            {data.emptyText}
          </p>
        ) : (
          cartLines.map((line) => (
            <ReservationCard
              key={line.item.id}
              line={line}
              onDecrease={() => updateCartQuantity(line.item.id, line.quantity - 1)}
              onIncrease={() => updateCartQuantity(line.item.id, line.quantity + 1)}
            />
          ))
        )}

        <div className="mt-4 space-y-2 border-t border-sl-e0ebe7 pt-4 text-sm">
          <div className="flex justify-between">
            <span>{data.subtotalLabel}</span>
            <span>{formatMoney(subtotal)}</span>
          </div>
          <div className="flex justify-between text-sl-1f6f64">
            <span>{data.companyCreditLabel}</span>
            <span>-{formatMoney(creditsApplied)}</span>
          </div>
          <div className="flex justify-between border-t border-sl-deebe6 pt-2 text-lg font-semibold">
            <span>{data.totalPayLabel}</span>
            <span>{formatMoney(totalPay)}</span>
          </div>
        </div>

        <Button type="button" onClick={() => setShowConfirmModal(true)} disabled={!canReserve} className="w-full">
          {placingOrder ? data.submittingLabel : data.reserveLabel}
        </Button>
        {data.deadlineText ? <p className="text-center text-xs text-sl-6f8680">{data.deadlineText}</p> : null}
      </CardContent>
    </Card>
  );
}
