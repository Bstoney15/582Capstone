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

function generateMockKeyValue() {
    const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789";
    let token = "sk_live_";
    for (let i = 0; i < 32; i += 1) {
        token += alphabet[Math.floor(Math.random() * alphabet.length)];
    }
    return token;
}

export default function ApiKeys() {
    const { selectedMerchant, isLoading } = useMerchant();
    const [keys, setKeys] = useState([]);
    const [newKeyName, setNewKeyName] = useState("");
    const [statusMessage, setStatusMessage] = useState("");

    const storageKey = useMemo(() => {
        const merchantId = selectedMerchant?.id ?? selectedMerchant?.merchant_id ?? selectedMerchant?.name ?? "default";
        return `mock:merchant:api_keys:${merchantId}`;
    }, [selectedMerchant]);

    useEffect(() => {
        const raw = localStorage.getItem(storageKey);
        if (!raw) {
            const seededKeys = [{
                id: crypto.randomUUID(),
                key: generateMockKeyValue(),
                name: "Production Checkout",
                generatedAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 3).toISOString()
            }];
            localStorage.setItem(storageKey, JSON.stringify(seededKeys));
            setKeys(seededKeys);
            return;
        }

        try {
            const parsed = JSON.parse(raw);
            setKeys(Array.isArray(parsed) ? parsed : []);
        } catch {
            setKeys([]);
        }
        setStatusMessage("");
        setNewKeyName("");
    }, [storageKey]);

    const persistKeys = (nextKeys) => {
        setKeys(nextKeys);
        localStorage.setItem(storageKey, JSON.stringify(nextKeys));
    };

    const onGenerateKey = (event) => {
        event.preventDefault();
        const name = newKeyName.trim();

        if (!name) {
            setStatusMessage("API key name is required.");
            return;
        }

        const newKey = {
            id: crypto.randomUUID(),
            key: generateMockKeyValue(),
            name,
            generatedAt: new Date().toISOString()
        };

        // Only one key is shown/managed in this mock UI.
        persistKeys([newKey]);
        setNewKeyName("");
        setStatusMessage("API key generated (mock POST /api/merchant/api_key). Secret shown once: " + newKey.key);
    };

    const onDeleteKey = (id) => {
        persistKeys(keys.filter((key) => key.id !== id));
        setStatusMessage("API key deleted (mock DELETE /api/merchant/api_key/<api_key>). ");
    };

    if (isLoading) {
        return <div style={{ padding: "2rem" }}>Loading...</div>;
    }

    return (
        <div style={{ padding: "2rem", maxWidth: "840px" }}>
            <h1>API Keys</h1>

            {selectedMerchant ? (
                <p style={{ marginBottom: "1rem" }}>
                    Create and delete API keys for invoice creation. Existing keys only show name and generation time.
                </p>
            ) : (
                <p style={{ marginBottom: "1rem" }}>Select a merchant account to manage API keys.</p>
            )}

            {statusMessage && (
                <p style={{ color: statusMessage.includes("required") ? "#b91c1c" : "#166534", marginBottom: "1rem", wordBreak: "break-word" }}>
                    {statusMessage}
                </p>
            )}

            {selectedMerchant && (
                <form onSubmit={onGenerateKey} style={{ marginBottom: "1.25rem", display: "flex", gap: "0.75rem", flexWrap: "wrap" }}>
                    <input
                        type="text"
                        value={newKeyName}
                        onChange={(event) => setNewKeyName(event.target.value)}
                        placeholder="Key name (e.g. Production Checkout)"
                        style={{ minWidth: "260px", flex: "1", padding: "0.625rem" }}
                    />
                    <button type="submit">Generate New Key</button>
                </form>
            )}

            {selectedMerchant && keys.length === 0 && <p>No API keys yet.</p>}

            {selectedMerchant && keys.length > 0 && (
                <div style={tableContainerStyle}>
                    <table style={{ width: "100%", borderCollapse: "collapse" }}>
                        <thead>
                            <tr>
                                <th style={headerCellStyle}>Key Name</th>
                                <th style={headerCellStyle}>Generated</th>
                                <th style={headerCellStyle}>Action</th>
                            </tr>
                        </thead>
                        <tbody>
                            {keys.map((key) => (
                                <tr key={key.id} style={{ borderTop: "1px solid var(--table-border)" }}>
                                    <td style={bodyCellStyle}>{key.name}</td>
                                    <td style={bodyCellStyle}>{new Date(key.generatedAt).toLocaleString()}</td>
                                    <td style={bodyCellStyle}>
                                        <button type="button" onClick={() => onDeleteKey(key.id)}>
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
