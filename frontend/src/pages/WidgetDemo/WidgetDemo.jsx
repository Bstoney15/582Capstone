// Author: Benjamin Stonestreet
// Created: 2026-03-03
// This is a demo page showcasing the XRPay checkout widget integration. It simulates a simple storefront where users can select quantities of coffee-related items and proceed to checkout using the XRPay widget. The page handles loading the widget script, creating an invoice on the backend, and starting the widget checkout flow. Status messages are displayed to guide the user through the process.


import { useEffect, useMemo, useState } from "react";

// DOM id used to find/remove a previously injected widget <script> tag
const widgetScriptId = "xrpay-widget-script";
// URL of the checkout widget script served by the backend; can be overridden per environment.
const widgetScriptSrc = import.meta.env.VITE_WIDGET_SCRIPT_SRC || "/widget/checkout.js";
// Demo API key used when creating invoices from this storefront demo.
const demoInvoiceApiKey = import.meta.env.VITE_WIDGET_DEMO_INVOICE_API_KEY || "dev_demo_invoice_key";

// Static catalogue of items available in the demo storefront
const storeItems = [
  { id: "coffee-beans", name: "Single Origin Beans", description: "250g medium roast", price: 24.0 },
  { id: "dripper", name: "Ceramic Dripper", description: "V60-compatible pour over", price: 18.0 },
  { id: "filters", name: "Paper Filters", description: "100 count pack", price: 8.0 },
];

/**
 * WidgetDemo acts as the main storefront component for the checkout demo,
 * handling cart state, total calculations, and triggering the checkout widget.
 * @returns {JSX.Element} The rendered WidgetDemo component.
 */
export default function WidgetDemo() {
  // Invoice ID returned by the backend after creating an invoice for the current cart
  const [invoiceId, setInvoiceId] = useState("(created at checkout)");
  // True once the widget <script> has loaded and exposed the expected MyPay API
  const [scriptReady, setScriptReady] = useState(false);
  // True once window.MyPay.init() has been called successfully
  const [initialized, setInitialized] = useState(false);
  // True while the invoice creation request is in-flight
  const [isCreatingInvoice, setIsCreatingInvoice] = useState(false);
  // Human-readable status string shown at the bottom of the page
  const [statusMessage, setStatusMessage] = useState("Loading checkout widget script...");
  // Per-item quantity selections keyed by item id
  const [quantities, setQuantities] = useState({
    "coffee-beans": 1,
    dripper: 1,
    filters: 1,
  });

  // URL the widget will redirect to after a successful payment; stable across renders
  const successUrl = useMemo(() => `${window.location.origin}/widget-demo?payment=queued`, []);

  // Derive line-item objects (with quantity and computed line total) from the static catalogue
  const lineItems = useMemo(
    () =>
      storeItems.map((item) => {
        const quantity = quantities[item.id] || 0;
        return {
          ...item,
          quantity,
          lineTotal: quantity * item.price,
        };
      }),
    [quantities]
  );

  // Sum of all line totals; recomputed whenever quantities change
  const subtotal = useMemo(() => lineItems.reduce((sum, item) => sum + item.lineTotal, 0), [lineItems]);

  // Effect: load the checkout widget <script> from the backend on mount.
  // If MyPay is already present with the required API surface (e.g. hot-reload), skip re-injection.
  useEffect(() => {
    // Check whether the current window.MyPay already exposes the expected v2 API methods
    const hasLatestMethods =
      !!window.MyPay &&
      typeof window.MyPay.setInvoiceId === "function" &&
      typeof window.MyPay.start === "function";

    if (hasLatestMethods) {
      setScriptReady(true);
      setStatusMessage("Widget script loaded. Click Checkout with XRPay.");
      return;
    }

    // Remove a stale MyPay global so the freshly injected script can redefine it cleanly
    if (window.MyPay) {
      delete window.MyPay;
    }

    // Remove any previously injected widget script tag before re-injecting
    const existingScript = document.getElementById(widgetScriptId);
    if (existingScript) {
      existingScript.remove();
    }

    const script = document.createElement("script");
    script.id = widgetScriptId;
    script.src = widgetScriptSrc;
    script.async = true;

    const handleLoad = () => {
      // After the script loads, verify it exposed the v2 API surface before marking it ready
      const loadedHasLatestMethods =
        !!window.MyPay &&
        typeof window.MyPay.setInvoiceId === "function" &&
        typeof window.MyPay.start === "function";

      if (!loadedHasLatestMethods) {
        setStatusMessage("Loaded widget script is outdated. Hard refresh and retry.");
        return;
      }

        setScriptReady(true);
      setStatusMessage("Widget script loaded. Click Checkout with XRPay.");
    };

    script.onload = handleLoad;

    // Provide a user-friendly error if the backend is not running or the path is wrong
    script.onerror = () => {
      setStatusMessage("Failed to load widget script from backend. Start backend on :8080 and retry.");
    };

    document.body.appendChild(script);

    // Cleanup: detach handlers to avoid stale-closure memory leaks (script stays in DOM intentionally)
    return () => {
      script.onload = null;
      script.onerror = null;
    };
  }, []);

  // Effect: initialise the widget once the script is ready and MyPay is available.
  // Runs only once per page lifecycle (guarded by the `initialized` flag).
  useEffect(() => {
    if (!scriptReady || initialized || !window.MyPay) {
      return;
    }

    try {
      // Pass a placeholder invoice id during init; the real id is set via setInvoiceId at checkout time
      window.MyPay.init({
        invoiceId: "seed-invoice-001",
        triggerSelector: "#widget-demo-hidden-trigger",
        successUrl,
        apiBaseUrl: "",
        debug: true,
      });

      setInitialized(true);
      setStatusMessage("Checkout ready. Click Checkout with XRPay to create invoice + start payment.");
    } catch (error) {
      setStatusMessage(error instanceof Error ? error.message : "Failed to initialize checkout widget.");
    }
  }, [scriptReady, initialized, successUrl]);

  /**
   * Creates a new invoice on the backend for the given cart total and returns the response body.
   * Throws an Error with a descriptive message on any failure so the caller can surface it.
   * @param {number} amountXRP The amount in XRP for the invoice.
   * @returns {Promise<Object>} The created invoice data.
   */
  const createInvoiceForCheckout = async (amountXRP) => {
    const amountUSD = amountXRP.toFixed(2);
    // Four-decimal precision required for XRP amounts by the backend API
    const compatibilityAmountXRP = amountXRP.toFixed(4);

    const response = await fetch("/api/invoices", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Accept: "application/json",
      },
      body: JSON.stringify({
        merchant_api_key: demoInvoiceApiKey,
        amount_usd: amountUSD,
        amount_xrp: compatibilityAmountXRP,
      }),
    });

    if (!response.ok) {
      const errorBody = await response.text();
      const detail = errorBody?.trim() ? `: ${errorBody.trim()}` : "";
      throw new Error(`Invoice creation failed (${response.status})${detail}`);
    }

    const responseBody = await response.json();
    if (!responseBody?.invoice_id || typeof responseBody.invoice_id !== "string") {
      throw new Error("Invoice creation response is missing invoice_id.");
    }

    return responseBody;
  };

  /**
   * Click handler for the "Checkout with XRPay" button.
   * Guards against uninitialized state and empty carts, then:
   *   1. Creates an invoice on the backend for the current cart subtotal.
   *   2. Passes the new invoice id to the widget via setInvoiceId.
   *   3. Starts the widget payment flow via MyPay.start().
   */
  const handleCheckout = async () => {
    if (!initialized || !window.MyPay) {
      setStatusMessage("Checkout is not initialized yet.");
      return;
    }

    if (subtotal <= 0) {
      setStatusMessage("Add at least one item before checkout.");
      return;
    }

    setIsCreatingInvoice(true);
    setStatusMessage("Creating invoice for cart total...");

    try {
      const createdInvoice = await createInvoiceForCheckout(subtotal);
      const createdInvoiceId = createdInvoice.invoice_id;
      setInvoiceId(createdInvoiceId);

      // Update status with conversion details when the backend returns XRP amount
      if (createdInvoice?.amount_xrp) {
        setStatusMessage(
          `Invoice ${createdInvoiceId.slice(0, 8)}... created: $${subtotal.toFixed(2)} -> ${createdInvoice.amount_xrp} XRP (${createdInvoice.pricing_source || "pricing"}).`
        );
      }

      // Safety check: ensure the runtime widget still exposes the v2 API before calling it
      if (typeof window.MyPay.setInvoiceId !== "function" || typeof window.MyPay.start !== "function") {
        throw new Error("Widget API is outdated (missing setInvoiceId/start). Hard refresh the page and retry.");
      }

      // Wire the freshly created invoice into the widget, then open the payment modal
      window.MyPay.setInvoiceId(createdInvoiceId);
      await window.MyPay.start();
    } catch (error) {
      setStatusMessage(error instanceof Error ? error.message : "Checkout failed.");
    } finally {
      setIsCreatingInvoice(false);
    }
  };

  /**
   * Updates the quantity for a single item, clamping the value to the range [0, 5].
   * @param {string} itemId The unique ID of the item.
   * @param {string|number} nextValue The new quantity value.
   */
  const updateQuantity = (itemId, nextValue) => {
    const parsed = Number(nextValue);
    // Clamp to valid range and floor to integer; default to 0 for non-finite inputs
    const safe = Number.isFinite(parsed) ? Math.max(0, Math.min(5, Math.floor(parsed))) : 0;
    setQuantities((previous) => ({
      ...previous,
      [itemId]: safe,
    }));
  };

  return (
    <div style={{ maxWidth: "920px", margin: "2rem auto", padding: "1rem", display: "grid", gap: "1rem" }}>
      <h1 style={{ margin: 0 }}>Coffee Shop Demo Storefront</h1>
      <p style={{ margin: 0, opacity: 0.85 }}>Pick a few items and checkout with the XRPay widget.</p>

      <div style={{ display: "grid", gridTemplateColumns: "2fr 1fr", gap: "1rem", alignItems: "start" }}>
        <div style={{ display: "grid", gap: "0.75rem" }}>
          {lineItems.map((item) => (
            <article
              key={item.id}
              style={{
                border: "1px solid #d1d5db",
                borderRadius: "10px",
                padding: "0.9rem",
                display: "grid",
                gap: "0.45rem",
              }}
            >
              <strong>{item.name}</strong>
              <span style={{ fontSize: "0.92rem", opacity: 0.8 }}>{item.description}</span>
              <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
                <span>${item.price.toFixed(2)}</span>
                <label style={{ display: "flex", alignItems: "center", gap: "0.4rem" }}>
                  Qty
                  <input
                    type="number"
                    min={0}
                    max={5}
                    value={item.quantity}
                    onChange={(event) => updateQuantity(item.id, event.target.value)}
                    style={{ width: "64px", padding: "0.3rem", border: "1px solid #d1d5db", borderRadius: "6px" }}
                  />
                </label>
              </div>
            </article>
          ))}
        </div>

        <aside style={{ border: "1px solid #d1d5db", borderRadius: "10px", padding: "1rem", display: "grid", gap: "0.8rem" }}>
          <h2 style={{ margin: 0, fontSize: "1.1rem" }}>Order Summary</h2>
          <div style={{ display: "grid", gap: "0.45rem", fontSize: "0.94rem" }}>
            {lineItems.map((item) => (
              <div key={`${item.id}-summary`} style={{ display: "flex", justifyContent: "space-between" }}>
                <span>{item.quantity} × {item.name}</span>
                <span>${item.lineTotal.toFixed(2)}</span>
              </div>
            ))}
          </div>
          <hr style={{ width: "100%", border: 0, borderTop: "1px solid #e5e7eb" }} />
          <div style={{ display: "flex", justifyContent: "space-between", fontWeight: 700 }}>
            <span>Total</span>
            <span>${subtotal.toFixed(2)}</span>
          </div>

          {/* Checkout button: disabled until widget is initialised, cart is non-empty, and no request is in-flight */}
      <button
            id="widget-demo-pay-button"
            type="button"
            onClick={handleCheckout}
            disabled={!initialized || subtotal <= 0 || isCreatingInvoice}
            style={{
              padding: "0.7rem 0.9rem",
              borderRadius: "8px",
              border: "1px solid #16a34a",
              background: "#16a34a",
              color: "white",
              cursor: initialized && subtotal > 0 && !isCreatingInvoice ? "pointer" : "not-allowed",
              opacity: initialized && subtotal > 0 && !isCreatingInvoice ? 1 : 0.7,
            }}
          >
            {isCreatingInvoice ? "Creating Invoice..." : "Checkout with XRPay"}
          </button>

          <p style={{ margin: 0, fontSize: "0.88rem", opacity: 0.75 }}>Demo invoice: {invoiceId}</p>
        </aside>
      </div>

      <p style={{ margin: 0, fontSize: "0.95rem" }}><strong>Status:</strong> {statusMessage}</p>
      {/* Hidden button used as the widget's triggerSelector target; never shown to the user */}
      <button id="widget-demo-hidden-trigger" type="button" style={{ display: "none" }} aria-hidden="true" tabIndex={-1}>
        Hidden widget trigger
      </button>
    </div>
  );
}