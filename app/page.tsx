"use client";

import { UserButton } from "@clerk/nextjs";
import { useMemo, useState } from "react";

type Role = "MEMBER" | "HIVE_MANAGER" | "INNOVATION_LEAD";

type MenuItem = {
  id: string;
  name: string;
  description: string;
  price: number;
  available: boolean;
};

type VendorMenu = {
  service_id: string;
  service_name: string;
  items: MenuItem[];
  error?: string;
};

type MemberOrderItem = {
  id: string;
  name: string;
  quantity: number;
  price: number;
};

type MemberOrder = {
  order_id: string;
  status: string;
  total_price: number;
  delivery_notes: string;
  items: MemberOrderItem[];
};

type PlaceOrderResponse = {
  order_id: string;
  status: string;
  total_price: number;
  remaining_credits: number;
};

type CartLine = {
  item: MenuItem;
  vendorName: string;
  quantity: number;
};

const SAMPLE_MEMBER_ID = "11111111-1111-4111-8111-111111111111";

const DEMO_MENUS: VendorMenu[] = [
  {
    service_id: "a1b2c3d4-1111-4000-a000-000000000001",
    service_name: "Bella Napoli Pizzeria",
    items: [
      {
        id: "f0000001-aaaa-4000-a000-000000000001",
        name: "Margherita",
        description: "San Marzano tomato, mozzarella di bufala, fresh basil",
        price: 12.5,
        available: true,
      },
      {
        id: "f0000001-aaaa-4000-a000-000000000002",
        name: "Diavola",
        description: "Spicy salami, roasted peppers, chilli oil",
        price: 14,
        available: true,
      },
      {
        id: "f0000001-aaaa-4000-a000-000000000004",
        name: "Capricciosa",
        description: "Artichoke, ham, mushrooms, olives",
        price: 16,
        available: false,
      },
    ],
  },
  {
    service_id: "a1b2c3d4-2222-4000-a000-000000000002",
    service_name: "Sakura Sushi Bar",
    items: [
      {
        id: "f0000002-bbbb-4000-a000-000000000001",
        name: "Salmon Nigiri (2pc)",
        description: "Fresh Atlantic salmon over seasoned rice",
        price: 8,
        available: true,
      },
      {
        id: "f0000002-bbbb-4000-a000-000000000002",
        name: "Dragon Roll (8pc)",
        description: "Shrimp tempura, avocado, eel sauce, tobiko",
        price: 16,
        available: true,
      },
      {
        id: "f0000002-bbbb-4000-a000-000000000005",
        name: "Miso Soup",
        description: "Traditional dashi broth with tofu, wakame, scallion",
        price: 4.5,
        available: true,
      },
    ],
  },
  {
    service_id: "a1b2c3d4-3333-4000-a000-000000000003",
    service_name: "El Fuego Taco Truck",
    items: [
      {
        id: "f0000003-cccc-4000-a000-000000000001",
        name: "Al Pastor Taco",
        description: "Marinated pork, pineapple, cilantro, onion",
        price: 4.5,
        available: true,
      },
      {
        id: "f0000003-cccc-4000-a000-000000000002",
        name: "Carnitas Burrito",
        description: "Slow-cooked pork, rice, beans, cheese, salsa verde",
        price: 11,
        available: true,
      },
      {
        id: "f0000003-cccc-4000-a000-000000000005",
        name: "Horchata",
        description: "Traditional rice and cinnamon drink",
        price: 3.5,
        available: true,
      },
    ],
  },
];

function formatMoney(value: number) {
  return `$${value.toFixed(2)}`;
}

function encodeBase64Url(value: string) {
  return btoa(value).replaceAll("+", "-").replaceAll("/", "_").replace(/=+$/, "");
}

function createDemoToken(memberId: string, role: Role) {
  const header = { alg: "HS256", typ: "JWT" };
  const payload = {
    sub: memberId,
    role,
    exp: Math.floor(Date.now() / 1000) + 60 * 60 * 12,
  };

  return `${encodeBase64Url(JSON.stringify(header))}.${encodeBase64Url(
    JSON.stringify(payload),
  )}.demo-signature`;
}

function getApiErrorMessage(payload: unknown) {
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

export default function Home() {
  const [apiBaseUrl] = useState("");
  const [memberId] = useState(SAMPLE_MEMBER_ID);
  const [memberToken] = useState(() => createDemoToken(SAMPLE_MEMBER_ID, "MEMBER"));
  const [managerToken] = useState(() => createDemoToken(SAMPLE_MEMBER_ID, "HIVE_MANAGER"));

  const [selectedVendorId, setSelectedVendorId] = useState(DEMO_MENUS[0]?.service_id ?? "");
  const [showConfirmModal, setShowConfirmModal] = useState(false);
  const [showGrantModal, setShowGrantModal] = useState(false);

  const [deliveryNotes, setDeliveryNotes] = useState("");
  const [grantReason, setGrantReason] = useState("Manual adjustment");
  const [grantInternalNote, setGrantInternalNote] = useState("");
  const [grantAmount, setGrantAmount] = useState("100");

  const [menus] = useState<VendorMenu[]>(DEMO_MENUS);
  const [cart, setCart] = useState<Record<string, CartLine>>({});
  const [credits, setCredits] = useState<number>(24);
  const [, setOrders] = useState<MemberOrder[]>([]);

  const [placingOrder, setPlacingOrder] = useState(false);
  const [grantingCredits, setGrantingCredits] = useState(false);
  const [errorMessage, setErrorMessage] = useState("");
  const [statusMessage, setStatusMessage] = useState("");

  const vendors = useMemo(
    () => menus.filter((vendor) => !vendor.error && vendor.items.length > 0),
    [menus],
  );
  const selectedVendor = useMemo(
    () => vendors.find((vendor) => vendor.service_id === selectedVendorId) ?? vendors[0] ?? null,
    [selectedVendorId, vendors],
  );

  const cartLines = useMemo(() => Object.values(cart), [cart]);
  const subtotal = useMemo(
    () =>
      cartLines.reduce((total, line) => total + line.quantity * line.item.price, 0),
    [cartLines],
  );
  const creditsApplied = useMemo(
    () => Math.min(credits, subtotal),
    [credits, subtotal],
  );
  const totalPay = useMemo(
    () => Math.max(0, subtotal - creditsApplied),
    [creditsApplied, subtotal],
  );
  const coverage = useMemo(() => {
    if (subtotal <= 0) {
      return 0;
    }
    return Math.min(100, Math.round((creditsApplied / subtotal) * 100));
  }, [creditsApplied, subtotal]);

  const canReserve = cartLines.length > 0 && !placingOrder;

  const resolveApiPath = (path: string) => {
    const base = apiBaseUrl.trim().replace(/\/$/, "");
    return `${base}${path}`;
  };

  const clearMessages = () => {
    setErrorMessage("");
    setStatusMessage("");
  };

  const addToCart = (vendorName: string, item: MenuItem) => {
    setCart((current) => {
      const previous = current[item.id];
      const quantity = previous ? previous.quantity + 1 : 1;
      return {
        ...current,
        [item.id]: { item, vendorName, quantity },
      };
    });
  };

  const updateCartQuantity = (itemId: string, quantity: number) => {
    setCart((current) => {
      if (quantity <= 0) {
        const next = { ...current };
        delete next[itemId];
        return next;
      }

      const existing = current[itemId];
      if (!existing) {
        return current;
      }

      return {
        ...current,
        [itemId]: {
          ...existing,
          quantity,
        },
      };
    });
  };

  const requestJson = async <T,>(
    path: string,
    options?: RequestInit,
    authRole?: Role,
  ): Promise<T> => {
    const headers = new Headers(options?.headers);
    headers.set("Content-Type", "application/json");

    if (authRole) {
      const token = authRole === "MEMBER" ? memberToken : managerToken;
      headers.set("Authorization", `Bearer ${token}`);
    }

    const response = await fetch(resolveApiPath(path), {
      ...options,
      headers,
    });

    const text = await response.text();
    let payload: unknown = null;
    if (text) {
      try {
        payload = JSON.parse(text) as unknown;
      } catch {
        payload = text;
      }
    }

    if (!response.ok) {
      throw new Error(getApiErrorMessage(payload));
    }

    return payload as T;
  };

  const refreshOrders = async () => {
    clearMessages();
    try {
      const payload = await requestJson<MemberOrder[]>(
        `/api/members/${memberId}/orders`,
        undefined,
        "MEMBER",
      );
      setOrders(payload);
      setStatusMessage("Order history refreshed.");
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : "Could not load orders.");
    }
  };

  const placeOrder = async () => {
    clearMessages();
    if (cartLines.length === 0) {
      setErrorMessage("Add at least one item to your cart first.");
      return false;
    }

    setPlacingOrder(true);

    try {
      const payload = await requestJson<PlaceOrderResponse>(
        "/api/orders",
        {
          method: "POST",
          body: JSON.stringify({
            member_id: memberId,
            delivery_notes: deliveryNotes,
            items: cartLines.map((line) => ({
              id: line.item.id,
              quantity: line.quantity,
              price: line.item.price,
            })),
          }),
        },
        "MEMBER",
      );

      setCart({});
      setDeliveryNotes("");
      setCredits(payload.remaining_credits);
      setStatusMessage(
        `Order ${payload.order_id} created (${payload.status}). Remaining credits: ${formatMoney(payload.remaining_credits)}.`,
      );
      await refreshOrders();
      return true;
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : "Could not place order.");
      return false;
    } finally {
      setPlacingOrder(false);
    }
  };

  const grantCredits = async () => {
    clearMessages();
    const amount = Number(grantAmount);
    if (!Number.isFinite(amount) || amount <= 0) {
      setErrorMessage("Grant amount must be a positive number.");
      return false;
    }

    setGrantingCredits(true);

    try {
      const payload = await requestJson<{ new_balance: number }>(
        `/api/members/${memberId}/credits`,
        {
          method: "POST",
          body: JSON.stringify({ amount }),
        },
        "HIVE_MANAGER",
      );
      setCredits(payload.new_balance);
      setStatusMessage(`Credits granted. New balance: ${formatMoney(payload.new_balance)}.`);
      return true;
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : "Could not grant credits.");
      return false;
    } finally {
      setGrantingCredits(false);
    }
  };

  const openConfirmOrder = () => {
    clearMessages();
    if (cartLines.length === 0) {
      setErrorMessage("Your tray is empty.");
      return;
    }
    setShowConfirmModal(true);
  };

  return (
    <main className="min-h-screen bg-[#f3f7f5] px-4 py-4 text-[#123830] md:px-8">
      <div className="mx-auto flex w-full max-w-[1240px] flex-col gap-5">
        <header className="rounded-2xl border border-[#dce9e4] bg-white/95 px-4 py-3 shadow-sm backdrop-blur">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div className="flex items-center gap-3">
              <div className="flex h-9 w-9 items-center justify-center rounded-full bg-[#1f6f64] text-white">
                #
              </div>
              <div>
                <p className="text-lg font-semibold leading-tight">SoftLunch</p>
                <p className="text-xs text-[#6e837c]">Marketplace</p>
              </div>
            </div>

            <div className="flex items-center gap-2">
              <span className="rounded-full border border-[#dbe9e4] bg-[#edf7f3] px-3 py-1 text-xs font-semibold text-[#245f55]">
                Credits: {formatMoney(credits)}
              </span>
              <button
                type="button"
                onClick={() => setShowGrantModal(true)}
                className="rounded-full bg-[#1f6f64] px-4 py-2 text-sm font-medium text-white"
              >
                Add credits
              </button>
              <div className="rounded-full border border-[#dbe9e4] bg-white p-1">
                <UserButton afterSignOutUrl="/auth" />
              </div>
            </div>
          </div>
        </header>

        <section className="grid gap-5 lg:grid-cols-[170px_minmax(0,1fr)_320px]">
          <aside className="space-y-4">
            <div className="rounded-2xl border border-[#dce9e5] bg-[#eef6f2] p-4">
              <p className="text-xs font-semibold uppercase tracking-wide text-[#44665e]">Dietary</p>
              <p className="mt-2 text-xs text-[#607b74]">GF = Gluten free</p>
              <p className="text-xs text-[#607b74]">V = Vegan</p>
              <p className="text-xs text-[#607b74]">VG = Vegetarian</p>
            </div>
          </aside>

          <div className="space-y-5">
            <div className="rounded-[28px] border border-[#d6e7e1] bg-gradient-to-r from-[#123830] via-[#1b5b52] to-[#286e63] p-6 text-white shadow-sm">
              <p className="text-xs uppercase tracking-wide text-[#b6ddd2]">Featured vendor</p>
              <h2 className="mt-3 text-4xl font-semibold leading-tight">
                {selectedVendor?.service_name ?? "Lunch Marketplace"}
              </h2>
              <p className="mt-2 max-w-xl text-sm text-[#d4ebe5]">
                Curated menus and team stipend ordering. Reserve before 11:30 AM.
              </p>
            </div>

            <div className="rounded-2xl border border-[#dce9e5] bg-white p-4 shadow-sm">
              <div className="flex flex-wrap gap-2">
                {vendors.map((vendor) => (
                  <button
                    key={vendor.service_id}
                    type="button"
                    onClick={() => setSelectedVendorId(vendor.service_id)}
                    className={`rounded-full px-3 py-1 text-sm ${
                      selectedVendor?.service_id === vendor.service_id
                        ? "bg-[#dff3ec] font-semibold text-[#1f6f64]"
                        : "bg-[#f2f7f5] text-[#4f6a64]"
                    }`}
                  >
                    {vendor.service_name}
                  </button>
                ))}
              </div>
            </div>

            <div className="space-y-3">
              <h3 className="text-3xl font-semibold">Popular Items</h3>
              <div className="space-y-3">
                {selectedVendor?.items.map((item) => (
                  <article
                    key={item.id}
                    className="rounded-2xl border border-[#dce9e5] bg-white p-4 shadow-sm"
                  >
                    <div className="flex items-start justify-between gap-3">
                      <div>
                        <p className="text-xl font-semibold">{item.name}</p>
                        <p className="mt-1 text-sm text-[#5a746d]">{item.description}</p>
                      </div>
                      <div className="text-right">
                        <p className="text-xl font-semibold text-[#1c5f54]">{formatMoney(item.price)}</p>
                        <button
                          type="button"
                          onClick={() =>
                            addToCart(selectedVendor?.service_name ?? "Vendor", item)
                          }
                          disabled={!item.available}
                          className="mt-3 rounded-full bg-[#1f6f64] px-3 py-2 text-sm font-medium text-white disabled:cursor-not-allowed disabled:bg-[#9eb7b1]"
                        >
                          {item.available ? "Add to tray" : "Unavailable"}
                        </button>
                      </div>
                    </div>
                  </article>
                ))}
              </div>
            </div>
          </div>

          <aside className="space-y-4">
            <div className="rounded-3xl border border-[#dce9e5] bg-white p-4 shadow-sm">
              <div className="mb-4 flex items-center justify-between">
                <h3 className="text-2xl font-semibold">Your Tray</h3>
                <span className="rounded-full bg-[#eef6f3] px-2 py-1 text-xs text-[#295b52]">
                  {cartLines.length} item{cartLines.length === 1 ? "" : "s"}
                </span>
              </div>

              {cartLines.length === 0 ? (
                <p className="rounded-2xl bg-[#f4f8f6] px-3 py-6 text-center text-sm text-[#6b847d]">
                  Your tray is looking empty.
                </p>
              ) : (
                <div className="space-y-3">
                  {cartLines.map((line) => (
                    <div
                      key={line.item.id}
                      className="rounded-2xl border border-[#e2ece8] bg-[#f9fcfb] p-3"
                    >
                      <div className="flex items-center justify-between gap-2">
                        <div>
                          <p className="font-semibold">{line.item.name}</p>
                          <p className="text-xs text-[#628078]">{line.vendorName}</p>
                        </div>
                        <p className="text-sm font-semibold">
                          {formatMoney(line.item.price * line.quantity)}
                        </p>
                      </div>
                      <div className="mt-2 flex items-center gap-2">
                        <button
                          type="button"
                          onClick={() => updateCartQuantity(line.item.id, line.quantity - 1)}
                          className="h-7 w-7 rounded-full border border-[#d6e6e0] text-sm"
                        >
                          -
                        </button>
                        <span className="text-sm">{line.quantity}</span>
                        <button
                          type="button"
                          onClick={() => updateCartQuantity(line.item.id, line.quantity + 1)}
                          className="h-7 w-7 rounded-full border border-[#d6e6e0] text-sm"
                        >
                          +
                        </button>
                      </div>
                    </div>
                  ))}
                </div>
              )}

              <div className="mt-4 space-y-2 border-t border-[#e0ebe7] pt-4 text-sm">
                <div className="flex justify-between">
                  <span>Subtotal</span>
                  <span>{formatMoney(subtotal)}</span>
                </div>
                <div className="flex justify-between text-[#1f6f64]">
                  <span>Company credit</span>
                  <span>-{formatMoney(creditsApplied)}</span>
                </div>
                <div className="flex justify-between border-t border-[#deebe6] pt-2 text-lg font-semibold">
                  <span>Total pay</span>
                  <span>{formatMoney(totalPay)}</span>
                </div>
              </div>

              <button
                type="button"
                onClick={openConfirmOrder}
                disabled={!canReserve}
                className="mt-4 w-full rounded-full bg-[#1f6f64] px-4 py-3 text-sm font-semibold text-white disabled:opacity-50"
              >
                {placingOrder ? "Submitting..." : "Reserve Lunch"}
              </button>
              <p className="mt-2 text-center text-xs text-[#6f8680]">
                Order by 11:30 AM for 12:15 PM delivery.
              </p>
            </div>
          </aside>
        </section>

        {(errorMessage || statusMessage) && (
          <section className="rounded-2xl border border-[#dce9e5] bg-white p-4 shadow-sm">
            {errorMessage && <p className="text-sm font-medium text-[#b33e2d]">{errorMessage}</p>}
            {statusMessage && <p className="text-sm font-medium text-[#245f55]">{statusMessage}</p>}
          </section>
        )}
      </div>

      {showConfirmModal && (
        <div className="fixed inset-0 z-40 flex items-center justify-center bg-black/30 p-4">
          <div className="w-full max-w-xl rounded-[28px] border border-[#d8e7e2] bg-white p-5 shadow-2xl">
            <div className="mb-4 flex items-start justify-between">
              <div>
                <h3 className="text-3xl font-semibold">Confirm Order</h3>
                <p className="text-sm text-[#5f7770]">Review your lunch details before confirming.</p>
              </div>
              <button
                type="button"
                onClick={() => setShowConfirmModal(false)}
                className="h-8 w-8 rounded-full bg-[#f3f8f6] text-lg"
              >
                x
              </button>
            </div>

            <div className="space-y-2">
              {cartLines.map((line) => (
                <div
                  key={line.item.id}
                  className="flex items-center justify-between rounded-xl bg-[#f7fbf9] px-3 py-2"
                >
                  <div>
                    <p className="font-medium">{line.item.name}</p>
                    <p className="text-xs text-[#67827a]">
                      {line.quantity}x - {line.vendorName}
                    </p>
                  </div>
                  <p className="font-semibold">{formatMoney(line.item.price * line.quantity)}</p>
                </div>
              ))}
            </div>

            <div className="mt-4 rounded-2xl bg-[#f5faf8] p-4">
              <p className="text-xs font-semibold uppercase tracking-wide text-[#67817a]">
                Payment breakdown
              </p>
              <div className="mt-2 flex items-center justify-between text-lg font-semibold">
                <span>{formatMoney(subtotal)} Total</span>
                <span className="text-[#1f6f64]">
                  {totalPay === 0 ? "Fully Covered" : `${formatMoney(totalPay)} Pay`}
                </span>
              </div>
              <div className="mt-2 h-2 rounded-full bg-[#dfece7]">
                <div className="h-2 rounded-full bg-[#2d6c60]" style={{ width: `${coverage}%` }} />
              </div>
              <p className="mt-1 text-xs text-[#607b74]">
                Credits ({formatMoney(credits)}) applied, {coverage}% covered.
              </p>
            </div>

            <label className="mt-4 block text-sm font-medium">
              Special Instructions
              <input
                className="mt-1 w-full rounded-2xl border border-[#d9e8e2] bg-[#f8fbfa] px-3 py-2"
                value={deliveryNotes}
                onChange={(event) => setDeliveryNotes(event.target.value)}
                placeholder="Leave at front desk, extra napkins..."
              />
            </label>

            <button
              type="button"
              disabled={placingOrder}
              onClick={async () => {
                const ok = await placeOrder();
                if (ok) {
                  setShowConfirmModal(false);
                }
              }}
              className="mt-4 w-full rounded-full bg-[#1f6f64] px-4 py-3 text-sm font-semibold text-white disabled:opacity-60"
            >
              {placingOrder ? "Placing order..." : "Slide to Order Lunch"}
            </button>
          </div>
        </div>
      )}

      {showGrantModal && (
        <div className="fixed inset-0 z-40 flex items-center justify-center bg-black/35 p-4">
          <div className="w-full max-w-md rounded-[28px] border border-[#d8e7e2] bg-white p-5 shadow-2xl">
            <div className="mb-4 flex items-start justify-between">
              <div>
                <h3 className="text-3xl font-semibold">Grant Credits</h3>
                <p className="text-sm text-[#5f7770]">Add funds to member account.</p>
              </div>
              <button
                type="button"
                onClick={() => setShowGrantModal(false)}
                className="h-8 w-8 rounded-full bg-[#f3f8f6] text-lg"
              >
                x
              </button>
            </div>

            <label className="block text-sm font-medium">
              Amount ($)
              <input
                value={grantAmount}
                onChange={(event) => setGrantAmount(event.target.value)}
                className="mt-1 w-full rounded-2xl border border-[#d7e6e1] bg-[#f8fbfa] px-3 py-2"
                inputMode="decimal"
                placeholder="0.00"
              />
            </label>

            <label className="mt-3 block text-sm font-medium">
              Reason for adjustment
              <select
                value={grantReason}
                onChange={(event) => setGrantReason(event.target.value)}
                className="mt-1 w-full rounded-2xl border border-[#d7e6e1] bg-[#f8fbfa] px-3 py-2"
              >
                <option>Manual adjustment</option>
                <option>Monthly top-up</option>
                <option>Compensation</option>
              </select>
            </label>

            <label className="mt-3 block text-sm font-medium">
              Internal note (optional)
              <textarea
                value={grantInternalNote}
                onChange={(event) => setGrantInternalNote(event.target.value)}
                className="mt-1 h-24 w-full rounded-2xl border border-[#d7e6e1] bg-[#f8fbfa] px-3 py-2"
                placeholder="Compensation for Monday's technical delay"
              />
            </label>

            <div className="mt-3 rounded-2xl bg-[#eef7f4] p-3 text-xs text-[#587771]">
              This amount will be available immediately for the member to spend on their next order.
            </div>

            <div className="mt-4 flex gap-2">
              <button
                type="button"
                onClick={() => setShowGrantModal(false)}
                className="flex-1 rounded-full border border-[#d8e7e2] px-3 py-2 text-sm"
              >
                Cancel
              </button>
              <button
                type="button"
                disabled={grantingCredits}
                onClick={async () => {
                  const ok = await grantCredits();
                  if (ok) {
                    setShowGrantModal(false);
                  }
                }}
                className="flex-1 rounded-full bg-[#20bea8] px-3 py-2 text-sm font-semibold text-white disabled:opacity-60"
              >
                {grantingCredits ? "Granting..." : "Grant Credits"}
              </button>
            </div>
          </div>
        </div>
      )}
    </main>
  );
}
