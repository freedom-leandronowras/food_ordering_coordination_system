import {
  MenuSections,
} from "@/components/(menu)/menu-sections";
import type { MenuSectionsData } from "@/components/(menu)/menu-section-types";

const menuSectionsData: MenuSectionsData = {
  header: {
    brandName: "SoftLunch",
    subtitle: "Menu",
    addCreditsButtonLabel: "Add credits",
    menuViewLabel: "Menu",
    managementViewLabel: "Management",
  },
  sidebar: {
    title: "Dietary",
    items: ["GF = Gluten free", "V = Vegan", "VG = Vegetarian"],
  },
  featured: {
    description: "Pick a vendor to browse today's menu.",
    vendorDescriptions: {
      "Bella Napoli Pizzeria": "Wood-fired Neapolitan pizzas with classic Italian toppings.",
      "Sakura Sushi Bar": "Fresh sushi, sashimi, and rolls prepared to order.",
      "El Fuego Taco Truck": "Street-style tacos, burritos, and quesadillas with bold flavor.",
    },
  },
  items: {
    title: "Popular Items",
    loadingText: "Loading menus from API...",
  },
  tray: {
    title: "Shopping Cart",
    emptyText: "Your tray is looking empty.",
    reserveLabel: "Reserve Lunch",
    submittingLabel: "Submitting...",
    deadlineText: "",
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
  management: {
    title: "Member Management",
    description:
      "Members are grouped by email domain by default. Optionally filter by domain and grant credits to a selected member account.",
    domainLabel: "Email domain",
    domainPlaceholder: "company.com",
    searchLabel: "Find members",
    searchingLabel: "Searching...",
    noResultsLabel: "No members found for the current filter.",
    memberIdLabel: "Member ID",
    grantAmountLabel: "Credits",
    grantLabel: "Grant credits",
    grantingLabel: "Granting...",
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
