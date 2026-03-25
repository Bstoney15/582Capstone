// Author: Charley Findling
// Created: 3/24/2026
// Description: This services file provides functions (to be redefined later when the backend is implemented)
//              that are called by the Webhooks.jsx page to manage the storage of webhooks.

function getMerchantId(merchant) {
  return merchant?.id ?? merchant?.merchant_id ?? merchant?.name ?? "default";
}

function storageKey(merchant) {
  return `mock:merchant:webhooks:${getMerchantId(merchant)}`;
}

// ── Mock implementations (replace internals with real fetch calls) ───────────

// get the webhooks from the backend based on merchant ID
export async function fetchWebhooks(merchant) {
  // TODO: replace with GET /api/merchants/:id/webhooks
  const raw = localStorage.getItem(storageKey(merchant));
  if (!raw) return [];
  try {
    const parsed = JSON.parse(raw);
    return Array.isArray(parsed) ? parsed : [];
  } catch {
    return [];
  }
}

// create a webhook for the merchant based on the webhook info
export async function createWebhook(merchant, { url, secret }) {
  // TODO: replace with POST /api/merchants/:id/webhooks  body: { url, secret }
  const webhook = {
    id: crypto.randomUUID(),
    url,
    hasSecret: Boolean(secret),
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  };

  // Backend will decide whether multiple webhooks are allowed.
  // For now the mock stores only the latest one.
  // one should be fine per merchant ?
  localStorage.setItem(storageKey(merchant), JSON.stringify([webhook]));
  return webhook;
}

// delete the specified webhook from the db
export async function deleteWebhook(merchant, webhookId) {
  // TODO: replace with DELETE /api/merchants/:id/webhooks/:webhookId
  const current = await fetchWebhooks(merchant);
  const next = current.filter((w) => w.id !== webhookId);
  localStorage.setItem(storageKey(merchant), JSON.stringify(next));
}