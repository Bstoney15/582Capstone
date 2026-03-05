import { useEffect, useMemo, useState } from "react";
import { useMerchant } from "../../contexts/MerchantContext";

const tableContainerStyle = {
    border: "1px solid var(--table-border)",
    borderRadius: "8px",
    overflow: "hidden",
    backgroundColor: "var(--table-surface)"
};

const headerCellStyle = {
    textAlign: "left",
    padding: "0.75rem",
    color: "var(--table-text)",
    backgroundColor: "var(--table-header)"
};

const bodyCellStyle = {
    padding: "0.75rem",
    color: "var(--table-text)",
    backgroundColor: "var(--table-cell)"
};

export default function Webhooks() {
    const { selectedMerchant, isLoading } = useMerchant();
    const [webhookUrl, setWebhookUrl] = useState("");
    const [webhookSecret, setWebhookSecret] = useState("");
    const [webhooks, setWebhooks] = useState([]);
    const [statusMessage, setStatusMessage] = useState("");

    const storageKey = useMemo(() => {
        const merchantId = selectedMerchant?.id ?? selectedMerchant?.merchant_id ?? selectedMerchant?.name ?? "default";
        return `mock:merchant:webhook:${merchantId}`;
    }, [selectedMerchant]);

    useEffect(() => {
        const raw = localStorage.getItem(storageKey);
        if (!raw) {
            const seededWebhooks = [{
                id: crypto.randomUUID(),
                url: "https://merchant.example.com/webhooks/invoices",
                hasSecret: true,
                updatedAt: new Date(Date.now() - 1000 * 60 * 60 * 6).toISOString()
            }];
            localStorage.setItem(storageKey, JSON.stringify(seededWebhooks));
            setWebhooks(seededWebhooks);
            setWebhookUrl("");
            setWebhookSecret("");
            setStatusMessage("");
            return;
        }

        try {
            const parsed = JSON.parse(raw);
            setWebhooks(Array.isArray(parsed) ? parsed : []);
            setWebhookUrl("");
            setWebhookSecret("");
        } catch {
            setWebhooks([]);
            setWebhookUrl("");
            setWebhookSecret("");
        }
        setStatusMessage("");
    }, [storageKey]);

    const persistWebhooks = (nextWebhooks) => {
        setWebhooks(nextWebhooks);
        localStorage.setItem(storageKey, JSON.stringify(nextWebhooks));
    };

    const onSaveWebhook = (event) => {
        event.preventDefault();
        const nextUrl = webhookUrl.trim();
        const nextSecret = webhookSecret.trim();

        if (!nextUrl || !nextSecret) {
            setStatusMessage("Webhook URL and key are required.");
            return;
        }

        const nextConfig = {
            id: crypto.randomUUID(),
            url: nextUrl,
            hasSecret: true,
            updatedAt: new Date().toISOString()
        };

        // Only one webhook is allowed in this mock UI.
        persistWebhooks([nextConfig]);
        setWebhookUrl("");
        setWebhookSecret("");
        setStatusMessage("Webhook saved (mock POST /api/merchant/webhook).");
    };

    const onDeleteWebhook = (id) => {
        const nextWebhooks = webhooks.filter((webhook) => webhook.id !== id);
        persistWebhooks(nextWebhooks);
        setStatusMessage("Webhook config deleted (mock DELETE /api/merchant/webhook).");
    };

    if (isLoading) {
        return <div style={{ padding: "2rem" }}>Loading...</div>;
    }

    return (
        <div style={{ padding: "2rem", maxWidth: "760px" }}>
            <h1>Webhooks</h1>

            {selectedMerchant ? (
                <p style={{ marginBottom: "1rem" }}>
                    Configure webhook delivery URL and shared secret key for real-time invoice events.
                </p>
            ) : (
                <p style={{ marginBottom: "1rem" }}>Select a merchant account to configure webhooks.</p>
            )}

            {statusMessage && (
                <p style={{ color: statusMessage.includes("required") ? "#b91c1c" : "#166534", marginBottom: "1rem" }}>
                    {statusMessage}
                </p>
            )}

            {selectedMerchant && (
                <form onSubmit={onSaveWebhook} style={{ marginBottom: "1.25rem" }}>
                    <label htmlFor="webhook-url" style={{ display: "block", marginBottom: "0.5rem" }}>
                        Webhook URL
                    </label>
                    <input
                        id="webhook-url"
                        type="url"
                        value={webhookUrl}
                        onChange={(event) => setWebhookUrl(event.target.value)}
                        placeholder="https://merchant.example.com/webhook"
                        style={{ width: "100%", padding: "0.625rem", marginBottom: "0.75rem" }}
                    />

                    <label htmlFor="webhook-secret" style={{ display: "block", marginBottom: "0.5rem" }}>
                        Webhook Secret Key
                    </label>
                    <input
                        id="webhook-secret"
                        type="password"
                        value={webhookSecret}
                        onChange={(event) => setWebhookSecret(event.target.value)}
                        placeholder="Enter webhook shared secret"
                        style={{ width: "100%", padding: "0.625rem", marginBottom: "0.75rem" }}
                    />

                    <button type="submit">Save Webhook</button>
                </form>
            )}

            {selectedMerchant && webhooks.length > 0 && (
                <div style={tableContainerStyle}>
                    <table style={{ width: "100%", borderCollapse: "collapse" }}>
                        <thead>
                            <tr>
                                <th style={headerCellStyle}>Webhook URL</th>
                                <th style={headerCellStyle}>Secret</th>
                                <th style={headerCellStyle}>Updated</th>
                                <th style={headerCellStyle}>Action</th>
                            </tr>
                        </thead>
                        <tbody>
                            {webhooks.map((webhook) => (
                                <tr key={webhook.id} style={{ borderTop: "1px solid var(--table-border)" }}>
                                    <td style={bodyCellStyle}>{webhook.url}</td>
                                    <td style={bodyCellStyle}>Configured</td>
                                    <td style={bodyCellStyle}>{new Date(webhook.updatedAt).toLocaleString()}</td>
                                    <td style={bodyCellStyle}>
                                        <button type="button" onClick={() => onDeleteWebhook(webhook.id)}>
                                            Delete
                                        </button>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            )}
        </div>
    );
}
