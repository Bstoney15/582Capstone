// Author: Charley Findling
// Created: 3/29/2026
// Description: Service functions for the Settings page.
//              Handles fetching user merchants and leaving a merchant.

const API_BASE = "/api";

// Fetch all merchants associated with the current user
export async function fetchUserMerchants() {
  const res = await fetch(`${API_BASE}/user/merchants`, {
    credentials: "include",
  });
  if (!res.ok) throw new Error("Failed to fetch merchants");
  return res.json();
}

// Remove the current user's association with a merchant (FR 10.1)
export async function leaveMerchant(merchantId) {
  const res = await fetch(`${API_BASE}/user/merchants/${merchantId}`, {
    method: "DELETE",
    credentials: "include",
  });
  if (!res.ok) {
    const body = await res.text();
    console.error(`Leave merchant failed: ${res.status} — ${body}`);
    throw new Error(body || "Failed to leave merchant");
  }
  return res.json();
}