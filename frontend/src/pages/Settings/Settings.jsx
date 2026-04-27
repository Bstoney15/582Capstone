// Settings.jsx – settings page for managing appearance preferences and merchant memberships.
// Author: Charley Findling
// Created: 2026-02-20

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

/**
 * Returns inline styles for the toggle track based on the on/off state.
 * @param {boolean} isOn Whether the toggle is active.
 * @returns {Object} Style object for the track element.
 */
const toggleTrackStyle = (isOn) => ({
  width: "48px",
  height: "26px",
  borderRadius: "13px",
  backgroundColor: isOn ? "#4f46e5" : "#d1d5db",
  position: "relative",
  cursor: "pointer",
  transition: "background-color 0.2s",
});

/**
 * Returns inline styles for the toggle knob based on the on/off state.
 * @param {boolean} isOn Whether the toggle is active.
 * @returns {Object} Style object for the knob element.
 */
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

/**
 * Settings renders the user settings page with appearance (dark mode toggle) and
 * merchant membership management (leave a merchant).
 * @returns {JSX.Element} The rendered Settings page.
 */
export default function Settings() {
  const { theme, toggleTheme } = useTheme();
  const isDark = theme === "dark";

  const [merchants, setMerchants] = useState([]);
  const [isLoading, setIsLoading] = useState(false);
  const [deletingId, setDeletingId] = useState(null);
  const [status, setStatus] = useState(null);

  /**
   * Fetches the list of merchants the current user belongs to.
   */
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

  /**
   * Removes the current user's role from the specified merchant after confirmation.
   * @param {string} merchantId The ID of the merchant to leave.
   * @param {string} merchantName The display name of the merchant (used in the confirm dialog).
   */
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

      {/* Merchant list */}
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
