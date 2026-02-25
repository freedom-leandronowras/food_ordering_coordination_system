"use client";

import { AppShell } from "@/components/(menu)/app-shell";
import { useMenuContext } from "@/components/(menu)/menu-context";
import { MenuLoadingState } from "@/components/(menu)/menu-loading-state";
import { DialogsSection } from "@/components/(menu)/sections/dialogs-section";
import { FeedbackSection } from "@/components/(menu)/sections/feedback-section";
import { HeaderSection } from "@/components/(menu)/sections/header-section";
import { MainSection } from "@/components/(menu)/sections/main-section";
import { MobileNavSection } from "@/components/(menu)/sections/mobile-nav-section";
import { SidebarSection } from "@/components/(menu)/sections/sidebar-section";
import { TraySection } from "@/components/(menu)/sections/tray-section";
import type { MenuSectionsData } from "@/components/(menu)/menu-section-types";

type MenuContentProps = {
  sectionsData: MenuSectionsData;
};

export function MenuContent({ sectionsData }: MenuContentProps) {
  const { isCheckingJwt } = useMenuContext();

  if (isCheckingJwt) {
    return <MenuLoadingState sectionsData={sectionsData} />;
  }

  return (
    <>
      <AppShell
        header={<HeaderSection data={sectionsData.header} />}
        sidebar={<SidebarSection data={sectionsData.sidebar} />}
        main={<MainSection featured={sectionsData.featured} items={sectionsData.items} />}
        aside={<TraySection data={sectionsData.tray} />}
        mobileNav={<MobileNavSection data={sectionsData.mobileNav} />}
      />

      <FeedbackSection />
      <DialogsSection data={sectionsData.dialogs} />
    </>
  );
}
