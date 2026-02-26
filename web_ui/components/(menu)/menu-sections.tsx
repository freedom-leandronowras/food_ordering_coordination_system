"use client";

import { MenuContent } from "@/components/(menu)/menu-content";
import { MenuProvider } from "@/components/(menu)/menu-context";
import { MenuMissingConfigState } from "@/components/(menu)/menu-missing-config-state";
import type { MenuSectionsData } from "@/components/(menu)/menu-section-types";

type MenuSectionsProps = {
  sectionsData: MenuSectionsData;
  apiBaseUrl: string;
};

export function MenuSections({ sectionsData, apiBaseUrl }: MenuSectionsProps) {
  if (!apiBaseUrl) {
    return <MenuMissingConfigState />;
  }

  return (
    <MenuProvider apiBaseUrl={apiBaseUrl}>
      <MenuContent sectionsData={sectionsData} />
    </MenuProvider>
  );
}
