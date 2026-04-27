// apiKeyService.js – service functions for fetching, creating, and deleting merchant API keys.
// Author: Benjamin Stonestreet
// Created: 2026-03-29

/**
 * Resolves the merchant ID from any of the known merchant object shapes.
 * @param {Object} merchant The merchant object from context.
 * @returns {string} The merchant ID, or an empty string if not resolvable.
 */
function getMerchantId(merchant) {
  return merchant?.id ?? merchant?.merchant_id ?? merchant?.name ?? "";
}

/**
 * Maps a server API key response object to the internal model shape.
 * @param {Object} serverKey The raw API key record from the server.
 * @returns {Object} Normalized key model with id, name, and generatedAt.
 */
function toApiKeyModel(serverKey) {
  return {
    id: serverKey.id,
    name: serverKey.name,
    generatedAt: serverKey.generated_at,
  };
}

/**
 * Fetches all active (non-revoked) API keys for the given merchant.
 * @param {Object} merchant The selected merchant object.
 * @returns {Promise<Array>} Array of normalized API key models.
 */
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

/**
 * Creates a new API key for the given merchant and returns the full response including the plaintext key.
 * @param {Object} merchant The selected merchant object.
 * @param {Object} options Key creation options.
 * @param {string} options.name The display name for the key.
 * @returns {Promise<Object>} The created key including the one-time plaintext api_key field.
 */
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

/**
 * Revokes (soft-deletes) an API key for the given merchant.
 * @param {Object} merchant The selected merchant object.
 * @param {string} apiKeyId The ID of the key to revoke.
 * @returns {Promise<void>}
 */
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
