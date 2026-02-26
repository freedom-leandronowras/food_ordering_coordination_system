"use client";

import { useCallback, useEffect, useMemo, useState } from "react";

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
import { getEmailDomain } from "@/lib/auth-policy";
import { formatMoney, type ManagedMember } from "@/lib/menu-data";

type ManagementSectionProps = {
  data: MenuSectionsData["management"];
};

export function ManagementSection({ data }: ManagementSectionProps) {
  const { isManager, lookupMembersByDomain, grantCreditsToMember, grantingCredits } = useMenuContext();
  const [loadingMembers, setLoadingMembers] = useState(false);
  const [members, setMembers] = useState<ManagedMember[]>([]);

  const [selectedMember, setSelectedMember] = useState<ManagedMember | null>(null);
  const [grantAmount, setGrantAmount] = useState("100");
  const [grantOpen, setGrantOpen] = useState(false);

  const groupedMembers = useMemo(() => {
    const groups = new Map<string, ManagedMember[]>();
    members.forEach((member) => {
      const key = getEmailDomain(member.email) || "unknown";
      const existing = groups.get(key) ?? [];
      existing.push(member);
      groups.set(key, existing);
    });

    return Array.from(groups.entries())
      .sort(([a], [b]) => a.localeCompare(b))
      .map(([domainKey, domainMembers]) => ({
        domain: domainKey,
        members: domainMembers.sort((a, b) => a.email.localeCompare(b.email)),
      }));
  }, [members]);

  const loadMembers = useCallback(async () => {
    setLoadingMembers(true);
    try {
      const fetchedMembers = await lookupMembersByDomain("");
      setMembers(fetchedMembers);
    } finally {
      setLoadingMembers(false);
    }
  }, [lookupMembersByDomain]);

  useEffect(() => {
    if (!isManager) {
      return;
    }
    void loadMembers();
  }, [isManager, loadMembers]);

  const openGrantModal = (member: ManagedMember) => {
    setSelectedMember(member);
    setGrantAmount("100");
    setGrantOpen(true);
  };

  const submitGrant = async () => {
    if (!selectedMember) {
      return;
    }

    const amount = Number(grantAmount);
    if (!Number.isFinite(amount) || amount <= 0) {
      return;
    }

    const ok = await grantCreditsToMember(selectedMember.member_id, amount);
    if (!ok) {
      return;
    }

    setGrantOpen(false);
    setSelectedMember(null);
    await loadMembers();
  };

  if (!isManager) {
    return (
      <Card className="rounded-3xl border-[#f0d1cf] bg-[#fff6f5] p-5 text-[#8f352c]">
        Manager access is required for domain member management.
      </Card>
    );
  }

  return (
    <>
      <div className="space-y-4">
        <Card className="rounded-3xl border-[#dce9e5] p-5">
          <h2 className="text-2xl font-semibold">{data.title}</h2>
          <p className="mt-2 text-sm text-[#5c746d]">{data.description}</p>
        </Card>

        {loadingMembers ? (
          <Card className="rounded-3xl border-[#dce9e5] p-5 text-sm text-[#607b74]">Loading members...</Card>
        ) : null}

        {groupedMembers.length === 0 ? (
          <Card className="rounded-3xl border-[#dce9e5] p-5 text-sm text-[#607b74]">{data.noResultsLabel}</Card>
        ) : (
          <div className="space-y-3">
            {groupedMembers.map((group) => (
              <Card key={group.domain} className="rounded-3xl border-[#dce9e5] p-4">
                <p className="mb-3 text-sm font-semibold uppercase tracking-wide text-[#54726a]">@{group.domain}</p>
                <div className="space-y-3">
                  {group.members.map((member) => (
                    <Card
                      key={member.user_id}
                      className="flex flex-col gap-3 rounded-3xl border-[#dce9e5] p-4 lg:flex-row lg:items-end lg:justify-between"
                    >
                      <div>
                        <p className="text-lg font-semibold text-[#1a4d45]">{member.full_name}</p>
                        <p className="text-sm text-[#4f6f67]">{member.email}</p>
                        <p className="mt-1 text-xs text-[#69827b]">
                          {data.memberIdLabel}: {member.member_id}
                        </p>
                        <p className="mt-1 text-xs font-semibold text-[#245f55]">
                          Current Credits: {formatMoney(Number(member.credits ?? 0))}
                        </p>
                      </div>

                      <Button type="button" onClick={() => openGrantModal(member)}>
                        {data.grantLabel}
                      </Button>
                    </Card>
                  ))}
                </div>
              </Card>
            ))}
          </div>
        )}
      </div>

      <Dialog open={grantOpen} onOpenChange={setGrantOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Grant Credits</DialogTitle>
            <DialogDescription>
              {selectedMember
                ? `Add credits to ${selectedMember.full_name} (${selectedMember.email}).`
                : "Add credits to member account."}
            </DialogDescription>
          </DialogHeader>

          <label className="block text-sm font-medium">
            {data.grantAmountLabel}
            <Input
              className="mt-1"
              value={grantAmount}
              inputMode="decimal"
              onChange={(event) => setGrantAmount(event.target.value)}
            />
          </label>

          <DialogFooter>
            <DialogClose asChild>
              <Button type="button" variant="outline">
                Cancel
              </Button>
            </DialogClose>
            <Button type="button" disabled={grantingCredits || !selectedMember} onClick={() => void submitGrant()}>
              {grantingCredits ? data.grantingLabel : data.grantLabel}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
