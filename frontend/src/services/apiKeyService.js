function getMerchantId(merchant) {
  return merchant?.id ?? merchant?.merchant_id ?? merchant?.name ?? "";
}

function toApiKeyModel(serverKey) {
  return {
    id: serverKey.id,
    name: serverKey.name,
    generatedAt: serverKey.generated_at,
  };
}

export async function fetchApiKeys(merchant) {
  const merchantId = getMerchantId(merchant);
  if (!merchantId) return [];

  const response = await fetch(`/api/merchant/api_key?merchant_id=${encodeURIComponent(merchantId)}`, {
    method: "GET",
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch API keys: ${response.status}`);
  }

  const payload = await response.json();
  if (!Array.isArray(payload)) return [];
  return payload.map(toApiKeyModel);
}

export async function createApiKey(merchant, { name }) {
  const merchantId = getMerchantId(merchant);
  if (!merchantId) {
    throw new Error("merchant_id is required");
  }

  const response = await fetch("/api/merchant/api_key", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      merchant_id: merchantId,
      name,
    }),
  });

  if (!response.ok) {
    throw new Error(`Failed to create API key: ${response.status}`);
  }

  const payload = await response.json();
  return {
    id: payload.id,
    name: payload.name,
    generatedAt: payload.generated_at,
    api_key: payload.api_key,
  };
}

export async function deleteApiKey(merchant, apiKeyId) {
  const merchantId = getMerchantId(merchant);
  if (!merchantId) {
    throw new Error("merchant_id is required");
  }

  const response = await fetch(
    `/api/merchant/api_key/${encodeURIComponent(apiKeyId)}?merchant_id=${encodeURIComponent(merchantId)}`,
    {
      method: "DELETE",
    }
  );

  if (!response.ok) {
    throw new Error(`Failed to delete API key: ${response.status}`);
  }
}
