"use client";

import { MenuStateCard } from "@/components/(menu)/menu-state-card";
import type { MenuSectionsData } from "@/components/(menu)/menu-section-types";

type MenuLoadingStateProps = {
  sectionsData: MenuSectionsData;
};

export function MenuLoadingState({ sectionsData }: MenuLoadingStateProps) {
  return <MenuStateCard message={sectionsData.loading.validatingSessionText} />;
}
