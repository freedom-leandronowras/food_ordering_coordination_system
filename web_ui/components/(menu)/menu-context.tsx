"use client";

import { useRouter } from "next/navigation";
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";

import {
  clearSessionToken,
  parseJwtClaims,
  readSessionToken,
} from "@/lib/auth-session";
import {
  isEmailAllowedForDomains,
  isManagerRole,
  normalizeDomain,
  parseAllowedEmailDomains,
  isValidDomain,
} from "@/lib/auth-policy";
import type {
  AuthenticatedMember,
  CartLine,
  CreditsResponse,
  GrantCreditsResponse,
  ManagedMember,
  MembersByDomainResponse,
  MemberOrder,
  MenuItem,
  PlaceOrderResponse,
  VendorMenu,
} from "@/lib/menu-data";
import { formatMoney, getApiErrorMessage } from "@/lib/menu-data";

export type OrderHistoryEntry = {
  memberId: string;
  memberEmail: string;
  memberName: string;
  order: MemberOrder;
};

type ViewMode = "menu" | "history" | "management";

type MenuContextValue = {
  isCheckingJwt: boolean;
  isBootstrapping: boolean;
  isManager: boolean;
  viewMode: ViewMode;
  setViewMode: (mode: ViewMode) => void;
  memberId: string;
  vendors: VendorMenu[];
  selectedVendor: VendorMenu | null;
  selectedVendorId: string;
  selectVendor: (vendorId: string) => void;
  cartLines: CartLine[];
  credits: number;
  subtotal: number;
  creditsApplied: number;
  totalPay: number;
  coverage: number;
  canReserve: boolean;
  placingOrder: boolean;
  grantingCredits: boolean;
  showConfirmModal: boolean;
  showGrantModal: boolean;
  setShowConfirmModal: (open: boolean) => void;
  setShowGrantModal: (open: boolean) => void;
  deliveryNotes: string;
  setDeliveryNotes: (value: string) => void;
  grantAmount: string;
  setGrantAmount: (value: string) => void;
  grantReason: string;
  setGrantReason: (value: string) => void;
  grantInternalNote: string;
  setGrantInternalNote: (value: string) => void;
  errorMessage: string;
  statusMessage: string;
  historyLoading: boolean;
  orderHistory: OrderHistoryEntry[];
  addToCart: (vendorName: string, item: MenuItem) => void;
  updateCartQuantity: (itemId: string, quantity: number) => void;
  placeOrder: () => Promise<boolean>;
  refreshOrderHistory: (options?: { silent?: boolean }) => Promise<void>;
  grantCredits: () => Promise<boolean>;
  grantCreditsToMember: (targetMemberId: string, amount: number) => Promise<boolean>;
  lookupMembersByDomain: (domain: string) => Promise<ManagedMember[]>;
  signOut: () => void;
};

type MenuProviderProps = {
  children: ReactNode;
  apiBaseUrl: string;
};

const MenuContext = createContext<MenuContextValue | null>(null);

function buildApiUrl(baseUrl: string, path: string) {
  const normalizedPath = path.startsWith("/") ? path : `/${path}`;
  return `${baseUrl}${normalizedPath}`;
}

function withAuthHeader(token: string, headers?: HeadersInit): Headers {
  const nextHeaders = new Headers(headers);
  nextHeaders.set("Authorization", `Bearer ${token}`);
  return nextHeaders;
}

async function requestJson<T,>(baseUrl: string, path: string, options?: RequestInit): Promise<T> {
  const headers = new Headers(options?.headers);
  if (options?.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  const response = await fetch(buildApiUrl(baseUrl, path), {
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
}

export function MenuProvider({ children, apiBaseUrl }: MenuProviderProps) {
  const router = useRouter();
  const allowedEmailDomains = useMemo(
    () => parseAllowedEmailDomains(process.env.NEXT_PUBLIC_ALLOWED_EMAIL_DOMAINS),
    [],
  );

  const [memberId, setMemberId] = useState("");
  const [memberEmail, setMemberEmail] = useState("");
  const [memberName, setMemberName] = useState("");
  const [isManager, setIsManager] = useState(false);
  const [viewMode, setViewMode] = useState<ViewMode>("menu");
  const [sessionToken, setSessionToken] = useState("");
  const [selectedVendorId, setSelectedVendorId] = useState("");
  const [menus, setMenus] = useState<VendorMenu[]>([]);
  const [cart, setCart] = useState<Record<string, CartLine>>({});
  const [credits, setCredits] = useState<number>(0);
  const [orderHistory, setOrderHistory] = useState<OrderHistoryEntry[]>([]);

  const [deliveryNotes, setDeliveryNotes] = useState("");
  const [grantReason, setGrantReason] = useState("Manual adjustment");
  const [grantInternalNote, setGrantInternalNote] = useState("");
  const [grantAmount, setGrantAmount] = useState("100");

  const [isBootstrapping, setIsBootstrapping] = useState(true);
  const [isCheckingJwt, setIsCheckingJwt] = useState(true);
  const [historyLoading, setHistoryLoading] = useState(false);
  const [placingOrder, setPlacingOrder] = useState(false);
  const [grantingCredits, setGrantingCredits] = useState(false);
  const [showConfirmModal, setShowConfirmModal] = useState(false);
  const [showGrantModal, setShowGrantModal] = useState(false);
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
    () => cartLines.reduce((total, line) => total + line.quantity * line.item.price, 0),
    [cartLines],
  );
  const creditsApplied = useMemo(() => Math.min(credits, subtotal), [credits, subtotal]);
  const totalPay = useMemo(() => Math.max(0, subtotal - creditsApplied), [creditsApplied, subtotal]);
  const coverage = useMemo(() => {
    if (subtotal <= 0) {
      return 0;
    }
    return Math.min(100, Math.round((creditsApplied / subtotal) * 100));
  }, [creditsApplied, subtotal]);
  const canReserve = cartLines.length > 0 && !placingOrder && !isBootstrapping;

  const clearMessages = useCallback(() => {
    setErrorMessage("");
    setStatusMessage("");
  }, []);

  const signOut = useCallback(() => {
    clearSessionToken();
    setSessionToken("");
    setMemberId("");
    setMemberEmail("");
    setMemberName("");
    setIsManager(false);
    setOrderHistory([]);
    router.replace("/auth");
  }, [router]);

  const fetchMemberOrders = useCallback(
    async (targetMemberId: string, token: string): Promise<MemberOrder[]> => {
      if (!targetMemberId) {
        return [];
      }
      return requestJson<MemberOrder[]>(apiBaseUrl, `/api/members/${encodeURIComponent(targetMemberId)}/orders`, {
        headers: withAuthHeader(token),
      });
    },
    [apiBaseUrl],
  );

  const refreshCredits = useCallback(
    async (targetMemberId: string, token: string) => {
      if (!targetMemberId) {
        return;
      }

      try {
        const payload = await requestJson<CreditsResponse>(
          apiBaseUrl,
          `/api/members/${encodeURIComponent(targetMemberId)}/credits`,
          {
            headers: withAuthHeader(token),
          },
        );
        setCredits(Number(payload.credits ?? 0));
      } catch (error) {
        const message = error instanceof Error ? error.message : "Could not load credits.";
        if (message === "MEMBER_NOT_FOUND" || message.includes("no credit account")) {
          setCredits(0);
          return;
        }
        throw error;
      }
    },
    [apiBaseUrl],
  );

  const bootstrapApiData = useCallback(
    async (targetMemberId: string, token: string, managerRole: boolean) => {
      clearMessages();
      setIsBootstrapping(true);
      try {
        const menusPayload = await requestJson<VendorMenu[]>(apiBaseUrl, "/api/menus");

        setMenus(menusPayload);
        setSelectedVendorId(menusPayload[0]?.service_id ?? "");

        if (managerRole) {
          await refreshCredits(targetMemberId, token);
          setOrderHistory([]);
        } else {
          const [, memberOrders] = await Promise.all([
            refreshCredits(targetMemberId, token),
            fetchMemberOrders(targetMemberId, token),
          ]);
          setOrderHistory(
            memberOrders.map((order) => ({
              memberId: targetMemberId,
              memberEmail,
              memberName,
              order,
            })),
          );
        }
      } catch (error) {
        setErrorMessage(error instanceof Error ? error.message : "Could not load API data.");
      } finally {
        setIsBootstrapping(false);
      }
    },
    [apiBaseUrl, clearMessages, fetchMemberOrders, memberEmail, memberName, refreshCredits],
  );

  useEffect(() => {
    let isMounted = true;

    const validateSession = async () => {
      if (!apiBaseUrl) {
        if (isMounted) {
          setErrorMessage("Missing NEXT_PUBLIC_API_BASE_URL.");
          setIsCheckingJwt(false);
        }
        return;
      }

      const token = readSessionToken();
      if (!token) {
        if (isMounted) {
          setIsCheckingJwt(false);
        }
        router.replace("/auth");
        return;
      }

      const claims = parseJwtClaims(token);
      if (!claims?.sub) {
        clearSessionToken();
        if (isMounted) {
          setIsCheckingJwt(false);
        }
        router.replace("/auth");
        return;
      }
      if (claims.exp && claims.exp <= Math.floor(Date.now() / 1000)) {
        clearSessionToken();
        if (isMounted) {
          setIsCheckingJwt(false);
        }
        router.replace("/auth");
        return;
      }

      try {
        const profile = await requestJson<AuthenticatedMember>(apiBaseUrl, "/api/auth/me", {
          headers: withAuthHeader(token),
          cache: "no-store",
        });

        const emailAddress = profile.email?.toLowerCase() ?? "";
        if (
          allowedEmailDomains.length > 0 &&
          (!emailAddress || !isEmailAllowedForDomains(emailAddress, allowedEmailDomains))
        ) {
          clearSessionToken();
          if (isMounted) {
            setIsCheckingJwt(false);
          }
          router.replace("/auth?error=EMAIL_DOMAIN_NOT_ALLOWED");
          return;
        }

        if (isMounted) {
          setSessionToken(token);
          setMemberId(profile.member_id || claims.sub || "");
          setMemberEmail(profile.email || "");
          setMemberName(profile.full_name || "");
          const managerRole = isManagerRole(profile.role || claims.role);
          setIsManager(managerRole);
          if (!managerRole) {
            setViewMode("menu");
          }
          setIsCheckingJwt(false);
        }
      } catch {
        clearSessionToken();
        if (isMounted) {
          setIsCheckingJwt(false);
        }
        router.replace("/auth");
      }
    };

    void validateSession();

    return () => {
      isMounted = false;
    };
  }, [allowedEmailDomains, apiBaseUrl, router]);

  useEffect(() => {
    if (!memberId || !sessionToken) {
      return;
    }
    void bootstrapApiData(memberId, sessionToken, isManager);
  }, [bootstrapApiData, isManager, memberId, sessionToken]);

  const refreshOrderHistory = useCallback(
    async (options?: { silent?: boolean }) => {
      if (!memberId || !sessionToken) {
        return;
      }
      if (!options?.silent) {
        clearMessages();
      }
      setHistoryLoading(true);

      try {
        if (isManager) {
          const payload = await requestJson<MembersByDomainResponse>(apiBaseUrl, "/api/auth/members", {
            headers: withAuthHeader(sessionToken),
            cache: "no-store",
          });
          const members = Array.isArray(payload.members)
            ? payload.members.filter((member) => (member.role || "").toUpperCase() === "MEMBER")
            : [];
          const memberOrders = await Promise.all(
            members.map(async (member) => {
              const orders = await fetchMemberOrders(member.member_id, sessionToken);
              return orders.map((order) => ({
                memberId: member.member_id,
                memberEmail: member.email || "",
                memberName: member.full_name || "",
                order,
              }));
            }),
          );
          setOrderHistory(memberOrders.flat());
        } else {
          const memberOrders = await fetchMemberOrders(memberId, sessionToken);
          setOrderHistory(
            memberOrders.map((order) => ({
              memberId,
              memberEmail,
              memberName,
              order,
            })),
          );
        }

        if (!options?.silent) {
          setStatusMessage("Order history refreshed.");
        }
      } catch (error) {
        setErrorMessage(error instanceof Error ? error.message : "Could not load orders.");
      } finally {
        setHistoryLoading(false);
      }
    },
    [apiBaseUrl, clearMessages, fetchMemberOrders, isManager, memberEmail, memberId, memberName, sessionToken],
  );

  useEffect(() => {
    if (viewMode !== "history") {
      return;
    }
    void refreshOrderHistory({ silent: true });
  }, [refreshOrderHistory, viewMode]);

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

  const placeOrder = async () => {
    clearMessages();
    if (!memberId) {
      setErrorMessage("Member id is not available in the session token.");
      return false;
    }
    if (!sessionToken) {
      setErrorMessage("Session token is missing.");
      return false;
    }
    if (cartLines.length === 0) {
      setErrorMessage("Add at least one item to your cart first.");
      return false;
    }

    setPlacingOrder(true);

    try {
      const responsePayload = await requestJson<PlaceOrderResponse>(apiBaseUrl, "/api/orders", {
        method: "POST",
        headers: withAuthHeader(sessionToken),
        body: JSON.stringify({
          member_id: memberId,
          delivery_notes: deliveryNotes,
          items: cartLines.map((line) => ({
            id: line.item.id,
            quantity: line.quantity,
            price: line.item.price,
          })),
        }),
      });

      setCart({});
      setDeliveryNotes("");
      setCredits(responsePayload.remaining_credits);

      setStatusMessage(
        `Order ${responsePayload.order_id} created (${responsePayload.status}). Remaining credits: ${formatMoney(responsePayload.remaining_credits)}.`,
      );
      await refreshOrderHistory({ silent: true });
      return true;
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : "Could not place order.");
      return false;
    } finally {
      setPlacingOrder(false);
    }
  };

  const grantCreditsToMember = useCallback(
    async (targetMemberId: string, amount: number) => {
      clearMessages();
      if (!isManager) {
        setErrorMessage("Only managers can grant credits.");
        return false;
      }
      if (!targetMemberId) {
        setErrorMessage("Target member id is required.");
        return false;
      }
      if (!sessionToken) {
        setErrorMessage("Session token is missing.");
        return false;
      }
      if (!Number.isFinite(amount) || amount <= 0) {
        setErrorMessage("Grant amount must be a positive number.");
        return false;
      }

      setGrantingCredits(true);
      try {
        const payload = await requestJson<GrantCreditsResponse>(
          apiBaseUrl,
          `/api/members/${encodeURIComponent(targetMemberId)}/credits`,
          {
            method: "POST",
            headers: withAuthHeader(sessionToken),
            body: JSON.stringify({ amount }),
          },
        );
        const updatedCredits = Number(payload.new_balance ?? 0);
        if (targetMemberId === memberId) {
          setCredits(updatedCredits);
        }
        setStatusMessage(
          `Granted ${formatMoney(amount)} to ${targetMemberId}. New balance: ${formatMoney(updatedCredits)}.`,
        );
        return true;
      } catch (error) {
        setErrorMessage(error instanceof Error ? error.message : "Could not grant credits.");
        return false;
      } finally {
        setGrantingCredits(false);
      }
    },
    [apiBaseUrl, clearMessages, isManager, memberId, sessionToken],
  );

  const grantCredits = async () => {
    clearMessages();
    if (!memberId) {
      setErrorMessage("Member id is not available in the session token.");
      return false;
    }
    const amount = Number(grantAmount);
    return grantCreditsToMember(memberId, amount);
  };

  const lookupMembersByDomain = useCallback(
    async (domain: string) => {
      clearMessages();
      if (!isManager) {
        setErrorMessage("Only managers can access member management.");
        return [];
      }
      if (!sessionToken) {
        setErrorMessage("Session token is missing.");
        return [];
      }

      const normalizedDomain = normalizeDomain(domain);
      if (normalizedDomain && !isValidDomain(normalizedDomain)) {
        setErrorMessage("Please enter a valid email domain.");
        return [];
      }

      const path = normalizedDomain
        ? `/api/auth/members?domain=${encodeURIComponent(normalizedDomain)}`
        : "/api/auth/members";

      try {
        const payload = await requestJson<MembersByDomainResponse>(apiBaseUrl, path, {
          headers: withAuthHeader(sessionToken),
          cache: "no-store",
        });
        const members = Array.isArray(payload.members) ? payload.members : [];
        if (payload.domain) {
          setStatusMessage(`Loaded ${members.length} member(s) for ${payload.domain}.`);
        } else {
          setStatusMessage(`Loaded ${members.length} member(s) across all domains.`);
        }
        return members;
      } catch (error) {
        setErrorMessage(error instanceof Error ? error.message : "Could not load members by domain.");
        return [];
      }
    },
    [apiBaseUrl, clearMessages, isManager, sessionToken],
  );

  const value: MenuContextValue = {
    isCheckingJwt,
    isBootstrapping,
    isManager,
    viewMode,
    setViewMode,
    memberId,
    vendors,
    selectedVendor,
    selectedVendorId,
    selectVendor: setSelectedVendorId,
    cartLines,
    credits,
    subtotal,
    creditsApplied,
    totalPay,
    coverage,
    canReserve,
    placingOrder,
    grantingCredits,
    showConfirmModal,
    showGrantModal,
    setShowConfirmModal,
    setShowGrantModal,
    deliveryNotes,
    setDeliveryNotes,
    grantAmount,
    setGrantAmount,
    grantReason,
    setGrantReason,
    grantInternalNote,
    setGrantInternalNote,
    errorMessage,
    statusMessage,
    historyLoading,
    orderHistory,
    addToCart,
    updateCartQuantity,
    placeOrder,
    refreshOrderHistory,
    grantCredits,
    grantCreditsToMember,
    lookupMembersByDomain,
    signOut,
  };

  return <MenuContext.Provider value={value}>{children}</MenuContext.Provider>;
}

export function useMenuContext() {
  const context = useContext(MenuContext);
  if (!context) {
    throw new Error("useMenuContext must be used inside MenuProvider.");
  }
  return context;
}
