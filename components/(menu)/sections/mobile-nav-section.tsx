import type { MenuSectionsData } from "@/components/(menu)/menu-section-types";
import { Button } from "@/components/ui/button";

type MobileNavSectionProps = {
  data: MenuSectionsData["mobileNav"];
};

export function MobileNavSection({ data }: MobileNavSectionProps) {
  return (
    <div className="grid grid-cols-3 gap-2">
      {data.items.slice(0, 3).map((item) => (
        <Button key={item} variant="ghost" size="sm" className="w-full">
          {item}
        </Button>
      ))}
    </div>
  );
}
