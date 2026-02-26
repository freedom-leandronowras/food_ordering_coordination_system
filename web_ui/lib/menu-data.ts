export type MenuItem = {
  id: string;
  name: string;
  description: string;
  price: number;
  available: boolean;
};

export type VendorMenu = {
  service_id: string;
  service_name: string;
  items: MenuItem[];
  error?: string;
};

export type MemberOrderItem = {
  id: string;
  name: string;
  quantity: number;
  price: number;
};

export type MemberOrder = {
  order_id: string;
  status: string;
  total_price: number;
  delivery_notes: string;
  items: MemberOrderItem[];
};

export type PlaceOrderResponse = {
  order_id: string;
  status: string;
  total_price: number;
  remaining_credits: number;
};

export type CartLine = {
  item: MenuItem;
  vendorName: string;
  quantity: number;
};

export type CreditsResponse = {
  member_id: string;
  credits: number;
};

export type GrantCreditsResponse = {
  member_id: string;
  new_balance: number;
};

export type ManagedMember = {
  user_id: string;
  member_id: string;
  email: string;
  full_name: string;
  role: string;
  credits: number;
};

export type MembersByDomainResponse = {
  domain: string;
  members: ManagedMember[];
};

export type AuthenticatedMember = {
  user_id: string;
  member_id: string;
  email: string;
  full_name: string;
  role: string;
  credits: number;
};

export function formatMoney(value: number) {
  return `$${value.toFixed(2)}`;
}

export function getApiErrorMessage(payload: unknown) {
  if (typeof payload === "object" && payload !== null) {
    const maybeMessage = payload as { message?: string; code?: string };
    if (maybeMessage.message) {
      return maybeMessage.message;
    }
    if (maybeMessage.code) {
      return maybeMessage.code;
    }
  }

  return "request failed";
}
