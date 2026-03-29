// Author: Charley Findling, Benjamin Stonestreet
// Created: 3/24/2026
// Description: This is the webhooks page that allows the user to view, create, and delete (currently) one webhook.
//              Calls async functions from webhookService.js to abstract away the server communications.

import { useMerchant } from "../../contexts/MerchantContext";
import { useCallback, useEffect, useState } from "react";
import {
  fetchWebhooks,
  createWebhook,
  deleteWebhook,
} from "../../services/webhookService";

/* ── Styles ─────────────────────────────────────────────────────────── */

const tableContainerStyle = {
  border: "1px solid var(--table-border)",
  borderRadius: "8px",
  overflow: "hidden",
  backgroundColor: "var(--table-surface)",
};

const headerCellStyle = {
  textAlign: "left",
  padding: "0.75rem",
  color: "var(--table-text)",
  backgroundColor: "var(--table-header)",
};

const bodyCellStyle = {
  padding: "0.75rem",
  color: "var(--table-text)",
  backgroundColor: "var(--table-cell)",
};

const STATUS_STYLES = {
  error: { color: "#b91c1c" },
  success: { color: "#166534" },
};

/* ── Component ──────────────────────────────────────────────────────── */

export default function Webhooks() {
  const { selectedMerchant, isLoading: isMerchantLoading } = useMerchant();

  // Form state
  const [webhookUrl, setWebhookUrl] = useState("");
  const [webhookSecret, setWebhookSecret] = useState("");

  // Data state
  const [webhooks, setWebhooks] = useState([]);
  const [isLoadingWebhooks, setIsLoadingWebhooks] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [deletingId, setDeletingId] = useState(null);

  // Feedback
  const [status, setStatus] = useState(null); // { type: "error"|"success", message: string }

  /* ── Fetch webhooks whenever they change ─────────────────── */

  const loadWebhooks = useCallback(async () => {
    if (!selectedMerchant) {
      setWebhooks([]);
      return;
    }
    setIsLoadingWebhooks(true);
    try {
      const data = await fetchWebhooks(selectedMerchant);
      setWebhooks(data);
    } catch (err) {
      console.error("Failed to load webhooks", err);
      setStatus({ type: "error", message: "Failed to load webhooks." });
    } finally {
      setIsLoadingWebhooks(false);
    }
  }, [selectedMerchant]);

  useEffect(() => {
    setStatus(null);
    setWebhookUrl("");
    setWebhookSecret("");
    loadWebhooks();
  }, [loadWebhooks]);

  /* ── Save ──────────────────────────────────────────────────────────── */

  const onSaveWebhook = async (event) => {
    event.preventDefault();
    const url = webhookUrl.trim();
    const secret = webhookSecret.trim();

    if (!url) {
      setStatus({ type: "error", message: "Webhook URL is required." });
      return;
    }
    if (!secret) {
      setStatus({ type: "error", message: "Webhook secret key is required." });
      return;
    }

    setIsSaving(true);
    setStatus(null);

    try {
      await createWebhook(selectedMerchant, { url, secret });
      setWebhookUrl("");
      setWebhookSecret("");
      setStatus({ type: "success", message: "Webhook saved successfully." });
      await loadWebhooks(); // refresh list from source of truth
    } catch (err) {
      console.error("Failed to save webhook", err);
      setStatus({ type: "error", message: "Failed to save webhook. Please try again." });
    } finally {
      setIsSaving(false);
    }
  };

  /* ── Delete ────────────────────────────────────────────────────────── */

  const onDeleteWebhook = async (id) => {
    if (!window.confirm("Are you sure you want to delete this webhook?")) return;

    setDeletingId(id);
    setStatus(null);

    try {
      await deleteWebhook(selectedMerchant, id);
      setStatus({ type: "success", message: "Webhook deleted." });
      await loadWebhooks();
    } catch (err) {
      console.error("Failed to delete webhook", err);
      setStatus({ type: "error", message: "Failed to delete webhook." });
    } finally {
      setDeletingId(null);
    }
  };

  /* ── Render ────────────────────────────────────────────────────────── */

  if (isMerchantLoading) {
    return <div style={{ padding: "2rem" }}>Loading...</div>;
  }

  return (
    <div style={{ padding: "2rem", maxWidth: "760px" }}>
      <h1>Webhooks</h1>

      {selectedMerchant ? (
        <p style={{ marginBottom: "1rem" }}>
          Configure webhook delivery URL and shared secret key for real-time
          invoice events.
        </p>
      ) : (
        <p style={{ marginBottom: "1rem" }}>
          Select a merchant account to configure webhooks.
        </p>
      )}

      {status && (
        <p
          role="alert"
          style={{ ...STATUS_STYLES[status.type], marginBottom: "1rem" }}
        >
          {status.message}
        </p>
      )}

      {selectedMerchant && (
        <form onSubmit={onSaveWebhook} style={{ marginBottom: "1.25rem" }}>
          <label
            htmlFor="webhook-url"
            style={{ display: "block", marginBottom: "0.5rem" }}
          >
            Webhook URL
          </label>
          <input
            id="webhook-url"
            type="url"
            required
            value={webhookUrl}
            onChange={(e) => setWebhookUrl(e.target.value)}
            placeholder="https://merchant.example.com/webhook"
            disabled={isSaving}
            style={{
              width: "100%",
              padding: "0.625rem",
              marginBottom: "0.75rem",
            }}
          />

          <label
            htmlFor="webhook-secret"
            style={{ display: "block", marginBottom: "0.5rem" }}
          >
            Webhook Secret Key
          </label>
          <input
            id="webhook-secret"
            type="password"
            required
            value={webhookSecret}
            onChange={(e) => setWebhookSecret(e.target.value)}
            placeholder="Enter webhook shared secret"
            disabled={isSaving}
            style={{
              width: "100%",
              padding: "0.625rem",
              marginBottom: "0.75rem",
            }}
          />

          <button type="submit" disabled={isSaving}>
            {isSaving ? "Saving…" : "Save Webhook"}
          </button>
        </form>
      )}

      {selectedMerchant && isLoadingWebhooks && (
        <p>Loading webhooks…</p>
      )}

      {selectedMerchant && !isLoadingWebhooks && webhooks.length > 0 && (
        <div style={tableContainerStyle}>
          <table style={{ width: "100%", borderCollapse: "collapse" }}>
            <thead>
              <tr>
                <th style={headerCellStyle}>Webhook URL</th>
                <th style={headerCellStyle}>Key</th>
                <th style={headerCellStyle}>Updated</th>
                <th style={headerCellStyle}>Action</th>
              </tr>
            </thead>
            <tbody>
              {webhooks.map((webhook) => (
                <tr
                  key={webhook.id}
                  style={{ borderTop: "1px solid var(--table-border)" }}
                >
                  <td style={bodyCellStyle}>{webhook.url}</td>
                  <td style={bodyCellStyle}>
                    {webhook.hasSecret ? "••••••••" : "None"}
                  </td>
                  <td style={bodyCellStyle}>
                    {new Date(webhook.updatedAt).toLocaleString()}
                  </td>
                  <td style={bodyCellStyle}>
                    <button
                      type="button"
                      disabled={deletingId === webhook.id}
                      onClick={() => onDeleteWebhook(webhook.id)}
                    >
                      {deletingId === webhook.id ? "Deleting…" : "Delete"}
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {selectedMerchant && !isLoadingWebhooks && webhooks.length === 0 && (
        <p style={{ color: "var(--table-text)", fontStyle: "italic" }}>
          No webhooks configured yet.
        </p>
      )}
    </div>
  );
}