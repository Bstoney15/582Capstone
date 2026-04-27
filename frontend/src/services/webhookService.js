function getMerchantId(merchant) {
  return merchant?.id ?? merchant?.merchant_id ?? merchant?.name ?? "";
}

export async function fetchWebhooks(merchant) {
  const id = getMerchantId(merchant);
  const res = await fetch(`/api/merchant/webhooks?merchant_id=${encodeURIComponent(id)}`);
  if (!res.ok) throw new Error("Failed to fetch webhooks");
  return res.json();
}

export async function createWebhook(merchant, { url, secret }) {
  const id = getMerchantId(merchant);
  const res = await fetch("/api/merchant/webhooks", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ merchant_id: id, url, secret }),
  });
  if (!res.ok) throw new Error("Failed to save webhook");
  return res.json();
}

export async function deleteWebhook(merchant, webhookId) {
  const id = getMerchantId(merchant);
  const res = await fetch(
    `/api/merchant/webhooks/${encodeURIComponent(webhookId)}?merchant_id=${encodeURIComponent(id)}`,
    { method: "DELETE" }
  );
  if (!res.ok) throw new Error("Failed to delete webhook");
}

export async function fetchWebhookLogs(merchant) {
  const id = getMerchantId(merchant);
  const res = await fetch(`/api/merchant/webhook_logs?merchant_id=${encodeURIComponent(id)}`);
  if (!res.ok) throw new Error("Failed to fetch webhook logs");
  return res.json();
}

export async function resendWebhook(merchant, logId) {
  const id = getMerchantId(merchant);
  const res = await fetch(
    `/api/merchant/webhook_logs/${encodeURIComponent(logId)}/resend?merchant_id=${encodeURIComponent(id)}`,
    { method: "POST" }
  );
  const data = await res.json();
  if (!res.ok && res.status !== 502) throw new Error("Resend request failed");
  return data;
}
