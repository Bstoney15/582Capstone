// Author: Charley Findling
// Created: 3/29/2026
// Description: Settings page - current functionality is listing associated merchants
//              with an option to sever the connection, and an option to remember dark mode.

import { useCallback, useEffect, useState } from "react";
import { useTheme } from "../../contexts/ThemeContext";
import {
  fetchUserMerchants,
  leaveMerchant,
} from "../../services/settingsService";

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

const sectionHeadingStyle = {
  marginTop: "2rem",
  marginBottom: "0.5rem",
};

const toggleContainerStyle = {
  display: "flex",
  alignItems: "center",
  gap: "0.75rem",
  marginBottom: "2rem",
};

const toggleTrackStyle = (isOn) => ({
  width: "48px",
  height: "26px",
  borderRadius: "13px",
  backgroundColor: isOn ? "#4f46e5" : "#d1d5db",
  position: "relative",
  cursor: "pointer",
  transition: "background-color 0.2s",
});

const toggleKnobStyle = (isOn) => ({
  width: "22px",
  height: "22px",
  borderRadius: "50%",
  backgroundColor: "#fff",
  position: "absolute",
  top: "2px",
  left: isOn ? "24px" : "2px",
  transition: "left 0.2s",
  boxShadow: "0 1px 3px rgba(0,0,0,0.2)",
});

/* ── Component ──────────────────────────────────────────────────────── */

export default function Settings() {
  /* ── Theme from context (FR 9.2) ───────────────────────────── */
  const { theme, toggleTheme } = useTheme();
  const isDark = theme === "dark";

  /* ── Merchant state (FR 9.1) ───────────────────────────────── */
  const [merchants, setMerchants] = useState([]);
  const [isLoading, setIsLoading] = useState(false);
  const [deletingId, setDeletingId] = useState(null);
  const [status, setStatus] = useState(null);

  /* ── Load merchants ────────────────────────────────────────── */

  const loadMerchants = useCallback(async () => {
    setIsLoading(true);
    try {
      const data = await fetchUserMerchants();
      setMerchants(data);
    } catch (err) {
      console.error("Failed to load merchants", err);
      setStatus({ type: "error", message: "Failed to load merchants." });
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    loadMerchants();
  }, [loadMerchants]);

  /* ── Leave merchant (FR 10.1) ──────────────────────────────── */

  const onLeaveMerchant = async (merchantId, merchantName) => {
    if (
      !window.confirm(
        `Are you sure you want to leave "${merchantName}"? This will revoke your access.`
      )
    )
      return;

    setDeletingId(merchantId);
    setStatus(null);

    try {
      await leaveMerchant(merchantId);
      setStatus({
        type: "success",
        message: `Successfully removed from "${merchantName}".`,
      });
      await loadMerchants();
    } catch (err) {
      console.error("Failed to leave merchant", err);
      setStatus({ type: "error", message: "Failed to remove merchant." });
    } finally {
      setDeletingId(null);
    }
  };

  /* ── Render ────────────────────────────────────────────────── */

  return (
    <div style={{ padding: "2rem", maxWidth: "760px" }}>
      <h1>Settings</h1>

      <h2 style={sectionHeadingStyle}>Appearance</h2>
      <div style={toggleContainerStyle}>
      <span>Dark Mode</span>
      <div
          role="switch"
          aria-checked={isDark}
          aria-label="Toggle dark mode"
          tabIndex={0}
          style={toggleTrackStyle(isDark)}
          onClick={toggleTheme}
          onKeyDown={(e) => {
          if (e.key === "Enter" || e.key === " ") {
              e.preventDefault();
              toggleTheme();
          }
          }}
      >
          <div style={toggleKnobStyle(isDark)} />
      </div>
      </div>

      {/* ── Merchant list (FR 9.1) ───────────────────────────── */}
      <h2 style={sectionHeadingStyle}>Your Merchants</h2>
      <p style={{ marginBottom: "1rem" }}>
        Manage merchant accounts linked to your profile. Leaving a merchant will
        revoke your access.
      </p>

      {status && (
        <p
          role="alert"
          style={{ ...STATUS_STYLES[status.type], marginBottom: "1rem" }}
        >
          {status.message}
        </p>
      )}

      {isLoading && <p>Loading merchants…</p>}

      {!isLoading && merchants.length > 0 && (
        <div style={tableContainerStyle}>
          <table style={{ width: "100%", borderCollapse: "collapse" }}>
            <thead>
              <tr>
                <th style={headerCellStyle}>Merchant Name</th>
                <th style={headerCellStyle}>Your Role</th>
                <th style={headerCellStyle}>Action</th>
              </tr>
            </thead>
            <tbody>
              {merchants.map((merchant) => (
                <tr
                  key={merchant.id}
                  style={{ borderTop: "1px solid var(--table-border)" }}
                >
                  <td style={bodyCellStyle}>{merchant.name}</td>
                  <td style={bodyCellStyle}>{merchant.role}</td>
                  <td style={bodyCellStyle}>
                    <button
                      type="button"
                      disabled={deletingId === merchant.id}
                      onClick={() =>
                        onLeaveMerchant(merchant.id, merchant.name)
                      }
                    >
                      {deletingId === merchant.id ? "Removing…" : "Leave"}
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {!isLoading && merchants.length === 0 && (
        <p style={{ color: "var(--table-text)", fontStyle: "italic" }}>
          You are not associated with any merchants.
        </p>
      )}
    </div>
  );
}