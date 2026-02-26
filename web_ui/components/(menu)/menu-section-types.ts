export type MenuSectionsData = {
  header: {
    brandName: string;
    subtitle: string;
    addCreditsButtonLabel: string;
    menuViewLabel: string;
    managementViewLabel: string;
  };
  sidebar: {
    title: string;
    items: string[];
  };
  featured: {
    description: string;
    vendorDescriptions: Record<string, string>;
  };
  items: {
    title: string;
    loadingText: string;
  };
  tray: {
    title: string;
    emptyText: string;
    reserveLabel: string;
    submittingLabel: string;
    deadlineText: string;
    subtotalLabel: string;
    companyCreditLabel: string;
    totalPayLabel: string;
  };
  dialogs: {
    confirm: {
      title: string;
      description: string;
      paymentBreakdownLabel: string;
      specialInstructionsLabel: string;
      specialInstructionsPlaceholder: string;
      submitLabel: string;
      submittingLabel: string;
    };
    grant: {
      title: string;
      description: string;
      amountLabel: string;
      amountPlaceholder: string;
      reasonLabel: string;
      reasons: string[];
      internalNoteLabel: string;
      internalNotePlaceholder: string;
      helperText: string;
      cancelLabel: string;
      submitLabel: string;
      submittingLabel: string;
    };
  };
  management: {
    title: string;
    description: string;
    domainLabel: string;
    domainPlaceholder: string;
    searchLabel: string;
    searchingLabel: string;
    noResultsLabel: string;
    memberIdLabel: string;
    grantAmountLabel: string;
    grantLabel: string;
    grantingLabel: string;
  };
  mobileNav: {
    items: string[];
  };
  loading: {
    validatingSessionText: string;
  };
};
