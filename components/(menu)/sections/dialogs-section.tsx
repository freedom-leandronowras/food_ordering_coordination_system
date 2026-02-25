"use client";

import { useMenuContext } from "@/components/(menu)/menu-context";
import type { MenuSectionsData } from "@/components/(menu)/menu-section-types";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { formatMoney } from "@/lib/menu-data";
import { cn } from "@/lib/utils";

type DialogsSectionProps = {
  data: MenuSectionsData["dialogs"];
};

export function DialogsSection({ data }: DialogsSectionProps) {
  const {
    showConfirmModal,
    setShowConfirmModal,
    showGrantModal,
    setShowGrantModal,
    cartLines,
    totalPay,
    subtotal,
    coverage,
    credits,
    deliveryNotes,
    setDeliveryNotes,
    placeOrder,
    placingOrder,
    memberId,
    grantAmount,
    setGrantAmount,
    grantReason,
    setGrantReason,
    grantInternalNote,
    setGrantInternalNote,
    grantingCredits,
    grantCredits,
  } = useMenuContext();

  return (
    <>
      <Dialog open={showConfirmModal} onOpenChange={setShowConfirmModal}>
        <DialogContent className="max-w-xl">
          <DialogHeader>
            <div className="flex items-start justify-between gap-4">
              <div>
                <DialogTitle>{data.confirm.title}</DialogTitle>
                <DialogDescription>{data.confirm.description}</DialogDescription>
              </div>
              <DialogClose asChild>
                <Button type="button" size="icon" variant="outline">
                  x
                </Button>
              </DialogClose>
            </div>
          </DialogHeader>

          <div className="space-y-2">
            {cartLines.map((line) => (
              <Card
                key={line.item.id}
                className="flex items-center justify-between rounded-xl border-0 bg-[#f7fbf9] px-3 py-2"
              >
                <div>
                  <p className="font-medium">{line.item.name}</p>
                  <p className="text-xs text-[#67827a]">
                    {line.quantity}x - {line.vendorName}
                  </p>
                </div>
                <p className="font-semibold">{formatMoney(line.item.price * line.quantity)}</p>
              </Card>
            ))}
          </div>

          <Card className="mt-4 rounded-2xl border-0 bg-[#f5faf8] p-4">
            <p className="text-xs font-semibold uppercase tracking-wide text-[#67817a]">
              {data.confirm.paymentBreakdownLabel}
            </p>
            <div className="mt-2 flex items-center justify-between text-lg font-semibold">
              <span>{formatMoney(subtotal)} Total</span>
              <span className="text-[#1f6f64]">
                {totalPay === 0 ? "Fully Covered" : `${formatMoney(totalPay)} Pay`}
              </span>
            </div>
            <div className="mt-2 h-2 rounded-full bg-[#dfece7]">
              <div className="h-2 rounded-full bg-[#2d6c60]" style={{ width: `${coverage}%` }} />
            </div>
            <p className="mt-1 text-xs text-[#607b74]">
              Credits ({formatMoney(credits)}) applied, {coverage}% covered.
            </p>
          </Card>

          <label className="mt-4 block text-sm font-medium">
            {data.confirm.specialInstructionsLabel}
            <Input
              className="mt-1"
              value={deliveryNotes}
              onChange={(event) => setDeliveryNotes(event.target.value)}
              placeholder={data.confirm.specialInstructionsPlaceholder}
            />
          </label>

          <DialogFooter>
            <Button
              type="button"
              disabled={placingOrder || !memberId}
              onClick={async () => {
                const ok = await placeOrder();
                if (ok) {
                  setShowConfirmModal(false);
                }
              }}
              className="w-full"
            >
              {placingOrder ? data.confirm.submittingLabel : data.confirm.submitLabel}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={showGrantModal} onOpenChange={setShowGrantModal}>
        <DialogContent>
          <DialogHeader>
            <div className="flex items-start justify-between gap-4">
              <div>
                <DialogTitle>{data.grant.title}</DialogTitle>
                <DialogDescription>{data.grant.description}</DialogDescription>
              </div>
              <DialogClose asChild>
                <Button type="button" size="icon" variant="outline">
                  x
                </Button>
              </DialogClose>
            </div>
          </DialogHeader>

          <label className="block text-sm font-medium">
            {data.grant.amountLabel}
            <Input
              value={grantAmount}
              onChange={(event) => setGrantAmount(event.target.value)}
              className="mt-1"
              inputMode="decimal"
              placeholder={data.grant.amountPlaceholder}
            />
          </label>

          <label className="mt-3 block text-sm font-medium">
            {data.grant.reasonLabel}
            <select
              value={grantReason}
              onChange={(event) => setGrantReason(event.target.value)}
              className={cn(
                "mt-1 flex h-10 w-full rounded-2xl border border-[#d7e6e1] bg-[#f8fbfa] px-3 py-2 text-sm text-[#123830] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#1f6f64]",
              )}
            >
              {data.grant.reasons.map((reason) => (
                <option key={reason}>{reason}</option>
              ))}
            </select>
          </label>

          <label className="mt-3 block text-sm font-medium">
            {data.grant.internalNoteLabel}
            <textarea
              value={grantInternalNote}
              onChange={(event) => setGrantInternalNote(event.target.value)}
              className={cn(
                "mt-1 h-24 w-full rounded-2xl border border-[#d7e6e1] bg-[#f8fbfa] px-3 py-2 text-sm text-[#123830] placeholder:text-[#7a908a] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#1f6f64]",
              )}
              placeholder={data.grant.internalNotePlaceholder}
            />
          </label>

          <Card className="mt-3 rounded-2xl border-0 bg-[#eef7f4] p-3 text-xs text-[#587771]">
            {data.grant.helperText}
          </Card>

          <DialogFooter>
            <DialogClose asChild>
              <Button type="button" variant="outline" className="flex-1">
                {data.grant.cancelLabel}
              </Button>
            </DialogClose>
            <Button
              type="button"
              disabled={grantingCredits || !memberId}
              onClick={async () => {
                const ok = await grantCredits();
                if (ok) {
                  setShowGrantModal(false);
                }
              }}
              className="flex-1"
            >
              {grantingCredits ? data.grant.submittingLabel : data.grant.submitLabel}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
