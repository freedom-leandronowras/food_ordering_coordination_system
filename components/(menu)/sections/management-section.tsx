"use client";

import { useMemo, useState } from "react";

import { useMenuContext } from "@/components/(menu)/menu-context";
import type { MenuSectionsData } from "@/components/(menu)/menu-section-types";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { parseAllowedEmailDomains } from "@/lib/auth-policy";
import type { ManagedMember } from "@/lib/menu-data";

type ManagementSectionProps = {
  data: MenuSectionsData["management"];
};

export function ManagementSection({ data }: ManagementSectionProps) {
  const { isManager, lookupMembersByDomain, grantCreditsToMember, grantingCredits } = useMenuContext();
  const defaultDomain = useMemo(
    () => parseAllowedEmailDomains(process.env.NEXT_PUBLIC_ALLOWED_EMAIL_DOMAINS)[0] ?? "",
    [],
  );
  const [domain, setDomain] = useState(defaultDomain);
  const [searching, setSearching] = useState(false);
  const [members, setMembers] = useState<ManagedMember[]>([]);
  const [grantAmounts, setGrantAmounts] = useState<Record<string, string>>({});
  const [activeGrantMemberId, setActiveGrantMemberId] = useState("");

  const onSearch = async () => {
    setSearching(true);
    try {
      const fetchedMembers = await lookupMembersByDomain(domain);
      setMembers(fetchedMembers);
      setGrantAmounts((current) => {
        const next = { ...current };
        fetchedMembers.forEach((member) => {
          if (!next[member.member_id]) {
            next[member.member_id] = "100";
          }
        });
        return next;
      });
    } finally {
      setSearching(false);
    }
  };

  const onGrant = async (member: ManagedMember) => {
    const amount = Number(grantAmounts[member.member_id] ?? "0");
    setActiveGrantMemberId(member.member_id);
    try {
      await grantCreditsToMember(member.member_id, amount);
    } finally {
      setActiveGrantMemberId("");
    }
  };

  if (!isManager) {
    return (
      <Card className="rounded-3xl border-[#f0d1cf] bg-[#fff6f5] p-5 text-[#8f352c]">
        Manager access is required for domain member management.
      </Card>
    );
  }

  return (
    <div className="space-y-4">
      <Card className="rounded-3xl border-[#dce9e5] p-5">
        <h2 className="text-2xl font-semibold">{data.title}</h2>
        <p className="mt-2 text-sm text-[#5c746d]">{data.description}</p>

        <div className="mt-4 flex flex-col gap-3 sm:flex-row sm:items-end">
          <label className="flex-1 text-sm font-medium">
            {data.domainLabel}
            <Input
              className="mt-1"
              value={domain}
              placeholder={data.domainPlaceholder}
              onChange={(event) => setDomain(event.target.value)}
            />
          </label>
          <Button type="button" onClick={onSearch} disabled={searching} className="sm:min-w-40">
            {searching ? data.searchingLabel : data.searchLabel}
          </Button>
        </div>
      </Card>

      {members.length === 0 ? (
        <Card className="rounded-3xl border-[#dce9e5] p-5 text-sm text-[#607b74]">{data.noResultsLabel}</Card>
      ) : (
        <div className="space-y-3">
          {members.map((member) => {
            const isGrantingThisMember = grantingCredits && activeGrantMemberId === member.member_id;

            return (
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
                </div>

                <div className="flex w-full flex-col gap-2 sm:flex-row sm:items-end lg:w-auto">
                  <label className="text-sm font-medium">
                    {data.grantAmountLabel}
                    <Input
                      value={grantAmounts[member.member_id] ?? ""}
                      inputMode="decimal"
                      className="mt-1 w-full sm:w-32"
                      onChange={(event) =>
                        setGrantAmounts((current) => ({
                          ...current,
                          [member.member_id]: event.target.value,
                        }))
                      }
                    />
                  </label>

                  <Button type="button" disabled={isGrantingThisMember} onClick={() => onGrant(member)}>
                    {isGrantingThisMember ? data.grantingLabel : data.grantLabel}
                  </Button>
                </div>
              </Card>
            );
          })}
        </div>
      )}
    </div>
  );
}
