"use client";

import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { CartLine, formatMoney } from "@/lib/menu-data";

type ReservationCardProps = {
  line: CartLine;
  onIncrease: () => void;
  onDecrease: () => void;
};

export function ReservationCard({ line, onIncrease, onDecrease }: ReservationCardProps) {
  return (
    <Card className="rounded-2xl border-sl-e2ece8 bg-sl-f9fcfb p-3">
      <div className="flex items-center justify-between gap-2">
        <div>
          <p className="font-semibold">{line.item.name}</p>
          <p className="text-xs text-sl-628078">{line.vendorName}</p>
        </div>
        <p className="text-sm font-semibold">{formatMoney(line.item.price * line.quantity)}</p>
      </div>
      <div className="mt-2 flex items-center gap-2">
        <Button type="button" size="icon" variant="outline" onClick={onDecrease} className="h-7 w-7">
          -
        </Button>
        <span className="text-sm">{line.quantity}</span>
        <Button type="button" size="icon" variant="outline" onClick={onIncrease} className="h-7 w-7">
          +
        </Button>
      </div>
    </Card>
  );
}
