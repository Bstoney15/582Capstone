import { useEffect, useMemo, useState } from "react";

const widgetScriptId = "xrpay-widget-script";
const widgetScriptSrc = "http://localhost:8080/widget/checkout.js?v=20260301-2";
const backendApiBaseUrl = "http://localhost:8080";
const devInvoiceApiKey = "dev_demo_invoice_key";

const storeItems = [
  { id: "coffee-beans", name: "Single Origin Beans", description: "250g medium roast", price: 24.0 },
  { id: "dripper", name: "Ceramic Dripper", description: "V60-compatible pour over", price: 18.0 },
  { id: "filters", name: "Paper Filters", description: "100 count pack", price: 8.0 },
];

export default function WidgetDemo() {
  const [invoiceId, setInvoiceId] = useState("(created at checkout)");
  const [scriptReady, setScriptReady] = useState(false);
  const [initialized, setInitialized] = useState(false);
  const [isCreatingInvoice, setIsCreatingInvoice] = useState(false);
  const [statusMessage, setStatusMessage] = useState("Loading checkout widget script...");
  const [quantities, setQuantities] = useState({
    "coffee-beans": 1,
    dripper: 1,
    filters: 1,
  });

  const successUrl = useMemo(() => `${window.location.origin}/widget-demo?payment=queued`, []);

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

  const subtotal = useMemo(() => lineItems.reduce((sum, item) => sum + item.lineTotal, 0), [lineItems]);

  useEffect(() => {
    const hasLatestMethods =
      !!window.MyPay &&
      typeof window.MyPay.setInvoiceId === "function" &&
      typeof window.MyPay.start === "function";

    if (hasLatestMethods) {
      setScriptReady(true);
      setStatusMessage("Widget script loaded. Click Checkout with XRPay.");
      return;
    }

    if (window.MyPay) {
      delete window.MyPay;
    }

    const existingScript = document.getElementById(widgetScriptId);
    if (existingScript) {
      existingScript.remove();
    }

    const script = document.createElement("script");
    script.id = widgetScriptId;
    script.src = widgetScriptSrc;
    script.async = true;

    const handleLoad = () => {
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

    script.onerror = () => {
      setStatusMessage("Failed to load widget script from backend. Start backend on :8080 and retry.");
    };

    document.body.appendChild(script);

    return () => {
      script.onload = null;
      script.onerror = null;
    };
  }, []);

  useEffect(() => {
    if (!scriptReady || initialized || !window.MyPay) {
      return;
    }

    try {
      window.MyPay.init({
        invoiceId: "seed-invoice-001",
        triggerSelector: "#widget-demo-hidden-trigger",
        successUrl,
        apiBaseUrl: backendApiBaseUrl,
        debug: true,
      });

      setInitialized(true);
      setStatusMessage("Checkout ready. Click Checkout with XRPay to create invoice + start payment.");
    } catch (error) {
      setStatusMessage(error instanceof Error ? error.message : "Failed to initialize checkout widget.");
    }
  }, [scriptReady, initialized, successUrl]);

  const createInvoiceForCheckout = async (amountXRP) => {
    const amountUSD = amountXRP.toFixed(2);
    const compatibilityAmountXRP = amountXRP.toFixed(4);

    const response = await fetch(`${backendApiBaseUrl}/api/invoices`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Accept: "application/json",
      },
      body: JSON.stringify({
        merchant_api_key: devInvoiceApiKey,
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

      if (createdInvoice?.amount_xrp) {
        setStatusMessage(
          `Invoice ${createdInvoiceId.slice(0, 8)}... created: $${subtotal.toFixed(2)} -> ${createdInvoice.amount_xrp} XRP (${createdInvoice.pricing_source || "pricing"}).`
        );
      }

      if (typeof window.MyPay.setInvoiceId !== "function" || typeof window.MyPay.start !== "function") {
        throw new Error("Widget API is outdated (missing setInvoiceId/start). Hard refresh the page and retry.");
      }

      window.MyPay.setInvoiceId(createdInvoiceId);
      await window.MyPay.start();
    } catch (error) {
      setStatusMessage(error instanceof Error ? error.message : "Checkout failed.");
    } finally {
      setIsCreatingInvoice(false);
    }
  };

  const updateQuantity = (itemId, nextValue) => {
    const parsed = Number(nextValue);
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
      <button id="widget-demo-hidden-trigger" type="button" style={{ display: "none" }} aria-hidden="true" tabIndex={-1}>
        Hidden widget trigger
      </button>
    </div>
  );
}