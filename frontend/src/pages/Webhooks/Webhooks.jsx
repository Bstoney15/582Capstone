// Author: Charley Findling, Benjamin Stonestreet
// Created: 3/24/2026

import { useMerchant } from "../../contexts/MerchantContext";
import { useCallback, useEffect, useState } from "react";
import {
  fetchWebhooks,
  createWebhook,
  deleteWebhook,
  fetchWebhookLogs,
  resendWebhook,
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

  // Config data state
  const [webhooks, setWebhooks] = useState([]);
  const [isLoadingWebhooks, setIsLoadingWebhooks] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [deletingId, setDeletingId] = useState(null);

  // Log state
  const [logs, setLogs] = useState([]);
  const [isLoadingLogs, setIsLoadingLogs] = useState(false);
  const [resendingId, setResendingId] = useState(null);

  // Feedback
  const [status, setStatus] = useState(null); // { type: "error"|"success", message: string }

  /* ── Loaders ──────────────────────────────────────────────────────── */

  const loadWebhooks = useCallback(async () => {
    if (!selectedMerchant) {
      setWebhooks([]);
      return;
    }
    setIsLoadingWebhooks(true);
    try {
      const data = await fetchWebhooks(selectedMerchant);
      setWebhooks(data);
    } catch {
      setStatus({ type: "error", message: "Failed to load webhooks." });
    } finally {
      setIsLoadingWebhooks(false);
    }
  }, [selectedMerchant]);

  const loadLogs = useCallback(async () => {
    if (!selectedMerchant) {
      setLogs([]);
      return;
    }
    setIsLoadingLogs(true);
    try {
      const data = await fetchWebhookLogs(selectedMerchant);
      setLogs(Array.isArray(data) ? data : []);
    } catch {
      // Non-critical — don't surface as a blocking error
    } finally {
      setIsLoadingLogs(false);
    }
  }, [selectedMerchant]);

  useEffect(() => {
    setStatus(null);
    setWebhookUrl("");
    setWebhookSecret("");
    loadWebhooks();
    loadLogs();
  }, [loadWebhooks, loadLogs]);

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
      await loadWebhooks();
    } catch {
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
    } catch {
      setStatus({ type: "error", message: "Failed to delete webhook." });
    } finally {
      setDeletingId(null);
    }
  };

  /* ── Resend ────────────────────────────────────────────────────────── */

  const onResend = async (logId) => {
    setResendingId(logId);
    setStatus(null);
    try {
      const result = await resendWebhook(selectedMerchant, logId);
      if (result.succeeded) {
        setStatus({ type: "success", message: "Webhook resent successfully." });
      } else {
        setStatus({ type: "error", message: `Resend failed: ${result.error || "unknown error"}` });
      }
      await loadLogs();
    } catch {
      setStatus({ type: "error", message: "Failed to resend webhook." });
    } finally {
      setResendingId(null);
    }
  };

  /* ── Render ────────────────────────────────────────────────────────── */

  if (isMerchantLoading) {
    return <div style={{ padding: "2rem" }}>Loading...</div>;
  }

  return (
    <div style={{ padding: "2rem", maxWidth: "900px" }}>
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

      {selectedMerchant && isLoadingWebhooks && <p>Loading webhooks…</p>}

      {selectedMerchant && !isLoadingWebhooks && webhooks.length > 0 && (
        <div style={tableContainerStyle}>
          <table style={{ width: "100%", borderCollapse: "collapse" }}>
            <thead>
              <tr>
                <th style={headerCellStyle}>Webhook URL</th>
                <th style={headerCellStyle}>Key</th>
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

      {/* ── Event Log ──────────────────────────────────────────────── */}

      {selectedMerchant && (
        <>
          <h2 style={{ marginTop: "2rem", marginBottom: "0.75rem" }}>Event Log</h2>
          <p style={{ marginBottom: "1rem" }}>
            History of outbound webhook dispatch attempts. Use Resend to
            manually re-deliver any event.
          </p>

          {isLoadingLogs && <p>Loading event log…</p>}

          {!isLoadingLogs && logs.length === 0 && (
            <p style={{ color: "var(--table-text)", fontStyle: "italic" }}>
              No webhook events recorded yet.
            </p>
          )}

          {!isLoadingLogs && logs.length > 0 && (
            <div style={tableContainerStyle}>
              <table style={{ width: "100%", borderCollapse: "collapse" }}>
                <thead>
                  <tr>
                    <th style={headerCellStyle}>Event</th>
                    <th style={headerCellStyle}>Invoice</th>
                    <th style={headerCellStyle}>Status</th>
                    <th style={headerCellStyle}>HTTP</th>
                    <th style={headerCellStyle}>Attempts</th>
                    <th style={headerCellStyle}>Time</th>
                    <th style={headerCellStyle}>Action</th>
                  </tr>
                </thead>
                <tbody>
                  {logs.map((log) => (
                    <tr
                      key={log.id}
                      style={{ borderTop: "1px solid var(--table-border)" }}
                    >
                      <td style={bodyCellStyle}>{log.event_type}</td>
                      <td style={{ ...bodyCellStyle, fontFamily: "monospace", fontSize: "0.8rem" }}>
                        {log.invoice_id ? log.invoice_id.slice(0, 8) + "…" : "—"}
                      </td>
                      <td style={bodyCellStyle}>
                        <span style={{ color: log.succeeded ? "#166534" : "#b91c1c", fontWeight: 600 }}>
                          {log.succeeded ? "OK" : "Failed"}
                        </span>
                      </td>
                      <td style={bodyCellStyle}>
                        {log.status_code > 0 ? log.status_code : "—"}
                      </td>
                      <td style={bodyCellStyle}>{log.attempts > 0 ? log.attempts : "—"}</td>
                      <td style={{ ...bodyCellStyle, fontSize: "0.85rem" }}>
                        {new Date(log.created_at).toLocaleString()}
                      </td>
                      <td style={bodyCellStyle}>
                        <button
                          type="button"
                          disabled={resendingId === log.id}
                          onClick={() => onResend(log.id)}
                        >
                          {resendingId === log.id ? "Sending…" : "Resend"}
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </>
      )}
    </div>
  );
}
