"use client";

import { useAuth } from "@clerk/nextjs";
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

import type {
  CartLine,
  CreditsResponse,
  GrantCreditsResponse,
  MemberOrder,
  MenuItem,
  PlaceOrderResponse,
  VendorMenu,
} from "@/lib/menu-data";
import { formatMoney, getApiErrorMessage } from "@/lib/menu-data";

type MenuContextValue = {
  isCheckingJwt: boolean;
  isBootstrapping: boolean;
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
  addToCart: (vendorName: string, item: MenuItem) => void;
  updateCartQuantity: (itemId: string, quantity: number) => void;
  placeOrder: () => Promise<boolean>;
  grantCredits: () => Promise<boolean>;
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

type JwtClaims = {
  sub?: string;
  exp?: number;
};

function parseJwtClaims(token: string): JwtClaims | null {
  const segments = token.split(".");
  if (segments.length < 2) {
    return null;
  }

  try {
    const payload = segments[1];
    const normalized = payload.replace(/-/g, "+").replace(/_/g, "/");
    const padded = normalized.padEnd(Math.ceil(normalized.length / 4) * 4, "=");
    const decoded = atob(padded);
    const claims = JSON.parse(decoded) as JwtClaims;
    return claims;
  } catch {
    return null;
  }
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
  const { isLoaded, getToken } = useAuth();
  const router = useRouter();

  const [memberId, setMemberId] = useState("");
  const [sessionToken, setSessionToken] = useState("");
  const [selectedVendorId, setSelectedVendorId] = useState("");
  const [menus, setMenus] = useState<VendorMenu[]>([]);
  const [cart, setCart] = useState<Record<string, CartLine>>({});
  const [credits, setCredits] = useState<number>(0);
  const [, setOrders] = useState<MemberOrder[]>([]);

  const [deliveryNotes, setDeliveryNotes] = useState("");
  const [grantReason, setGrantReason] = useState("Manual adjustment");
  const [grantInternalNote, setGrantInternalNote] = useState("");
  const [grantAmount, setGrantAmount] = useState("100");

  const [isBootstrapping, setIsBootstrapping] = useState(true);
  const [isCheckingJwt, setIsCheckingJwt] = useState(true);
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

  const refreshOrders = useCallback(
    async (targetMemberId: string, token: string, options?: { silent?: boolean }) => {
      if (!targetMemberId) {
        return;
      }
      if (!options?.silent) {
        clearMessages();
      }

      try {
        const payload = await requestJson<MemberOrder[]>(
          apiBaseUrl,
          `/members/${encodeURIComponent(targetMemberId)}/orders`,
          {
            headers: withAuthHeader(token),
          },
        );
        setOrders(payload);
        if (!options?.silent) {
          setStatusMessage("Order history refreshed.");
        }
      } catch (error) {
        setErrorMessage(error instanceof Error ? error.message : "Could not load orders.");
      }
    },
    [apiBaseUrl, clearMessages],
  );

  const refreshCredits = useCallback(
    async (targetMemberId: string, token: string) => {
      if (!targetMemberId) {
        return;
      }

      try {
        const payload = await requestJson<CreditsResponse>(
          apiBaseUrl,
          `/members/${encodeURIComponent(targetMemberId)}/credits`,
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

  const bootstrapApiData = useCallback(async (targetMemberId: string, token: string) => {
    clearMessages();
    setIsBootstrapping(true);
    try {
      const menusPayload = await requestJson<VendorMenu[]>(apiBaseUrl, "/menus");

      setMenus(menusPayload);
      setSelectedVendorId(menusPayload[0]?.service_id ?? "");

      await Promise.all([
        refreshCredits(targetMemberId, token),
        refreshOrders(targetMemberId, token, { silent: true }),
      ]);
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : "Could not load API data.");
    } finally {
      setIsBootstrapping(false);
    }
  }, [apiBaseUrl, clearMessages, refreshCredits, refreshOrders]);

  useEffect(() => {
    let isMounted = true;

    const validateJwt = async () => {
      if (!isLoaded) {
        return;
      }

      const token = await getToken();
      if (!token) {
        if (isMounted) {
          setIsCheckingJwt(false);
        }
        router.replace("/auth");
        return;
      }

      const claims = parseJwtClaims(token);
      if (!claims?.sub) {
        if (isMounted) {
          setErrorMessage("Session token is missing a valid member id (sub claim).");
          setIsCheckingJwt(false);
        }
        return;
      }

      if (claims.exp && claims.exp <= Math.floor(Date.now() / 1000)) {
        if (isMounted) {
          setIsCheckingJwt(false);
        }
        router.replace("/auth");
        return;
      }

      if (isMounted) {
        setSessionToken(token);
        setMemberId(claims.sub);
        setIsCheckingJwt(false);
      }
    };

    void validateJwt();

    return () => {
      isMounted = false;
    };
  }, [getToken, isLoaded, router]);

  useEffect(() => {
    if (!memberId || !sessionToken) {
      return;
    }
    void bootstrapApiData(memberId, sessionToken);
  }, [bootstrapApiData, memberId, sessionToken]);

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
      const responsePayload = await requestJson<PlaceOrderResponse>(apiBaseUrl, "/orders", {
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
      await refreshOrders(memberId, sessionToken, { silent: true });
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
    if (!memberId) {
      setErrorMessage("Member id is not available in the session token.");
      return false;
    }
    if (!sessionToken) {
      setErrorMessage("Session token is missing.");
      return false;
    }

    const amount = Number(grantAmount);
    if (!Number.isFinite(amount) || amount <= 0) {
      setErrorMessage("Grant amount must be a positive number.");
      return false;
    }

    setGrantingCredits(true);
    try {
      const payload = await requestJson<GrantCreditsResponse>(
        apiBaseUrl,
        `/members/${encodeURIComponent(memberId)}/credits`,
        {
          method: "POST",
          headers: withAuthHeader(sessionToken),
          body: JSON.stringify({ amount }),
        },
      );
      const updatedCredits = Number(payload.new_balance ?? 0);
      setCredits(updatedCredits);
      setStatusMessage(`Credits granted. New balance: ${formatMoney(updatedCredits)}.`);
      return true;
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : "Could not grant credits.");
      return false;
    } finally {
      setGrantingCredits(false);
    }
  };

  const value: MenuContextValue = {
    isCheckingJwt,
    isBootstrapping,
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
    addToCart,
    updateCartQuantity,
    placeOrder,
    grantCredits,
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
