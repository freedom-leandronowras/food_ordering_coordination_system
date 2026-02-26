import type { MenuSectionsData } from "@/components/(menu)/menu-section-types";
import { Card } from "@/components/ui/card";

type SidebarSectionProps = {
  data: MenuSectionsData["sidebar"];
};

export function SidebarSection({ data }: SidebarSectionProps) {
  return (
    <Card className="rounded-2xl border-sl-dce9e5 bg-sl-eef6f2 p-4">
      <p className="text-xs font-semibold uppercase tracking-wide text-sl-44665e">{data.title}</p>
      {data.items.map((item) => (
        <p key={item} className="mt-2 text-xs text-sl-607b74">
          {item}
        </p>
      ))}
    </Card>
  );
}
