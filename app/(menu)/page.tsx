import {
  MenuSections,
} from "@/components/(menu)/menu-sections";
import type { MenuSectionsData } from "@/components/(menu)/menu-section-types";

const menuSectionsData: MenuSectionsData = {
  header: {
    brandName: "SoftLunch",
    subtitle: "Menu",
    addCreditsButtonLabel: "Add credits",
  },
  sidebar: {
    title: "Dietary",
    items: ["GF = Gluten free", "V = Vegan", "VG = Vegetarian"],
  },
  featured: {
    label: "Featured vendor",
    description:
      "Curated menus and team stipend ordering. Reserve before 11:30 AM.",
  },
  items: {
    title: "Popular Items",
    loadingText: "Loading menus from API...",
  },
  tray: {
    title: "Your Tray",
    emptyText: "Your tray is looking empty.",
    reserveLabel: "Reserve Lunch",
    submittingLabel: "Submitting...",
    deadlineText: "Order by 11:30 AM for 12:15 PM delivery.",
    subtotalLabel: "Subtotal",
    companyCreditLabel: "Company credit",
    totalPayLabel: "Total pay",
  },
  dialogs: {
    confirm: {
      title: "Confirm Order",
      description: "Review your lunch details before confirming.",
      paymentBreakdownLabel: "Payment breakdown",
      specialInstructionsLabel: "Special Instructions",
      specialInstructionsPlaceholder: "Leave at front desk, extra napkins...",
      submitLabel: "Slide to Order Lunch",
      submittingLabel: "Placing order...",
    },
    grant: {
      title: "Grant Credits",
      description: "Add funds to member account.",
      amountLabel: "Amount ($)",
      amountPlaceholder: "0.00",
      reasonLabel: "Reason for adjustment",
      reasons: ["Manual adjustment", "Monthly top-up", "Compensation"],
      internalNoteLabel: "Internal note (optional)",
      internalNotePlaceholder: "Compensation for Monday's technical delay",
      helperText:
        "This amount will be available immediately for the member to spend on their next order.",
      cancelLabel: "Cancel",
      submitLabel: "Grant Credits",
      submittingLabel: "Granting...",
    },
  },
  mobileNav: {
    items: ["Menu", "Orders", "Account"],
  },
  loading: {
    validatingSessionText: "Validating session...",
  },
};

export default function MenuHomePage() {
  const apiBaseUrl = process.env.NEXT_PUBLIC_API_BASE_URL ?? "";

  return (
    <MenuSections
      sectionsData={menuSectionsData}
      apiBaseUrl={apiBaseUrl}
    />
  );
}
