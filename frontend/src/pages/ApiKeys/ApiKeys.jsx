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

const modalOverlayStyle = {
    position: "fixed",
    inset: 0,
    backgroundColor: "color-mix(in srgb, var(--color-base) 60%, transparent)",
    display: "flex",
    alignItems: "center",
    justifyContent: "center",
    zIndex: 1000,
    padding: "1rem"
};

const modalCardStyle = {
    width: "min(640px, 100%)",
    backgroundColor: "var(--color-base-variant)",
    color: "var(--color-text)",
    border: "1px solid var(--table-border)",
    borderRadius: "10px",
    padding: "1rem",
    boxShadow: "0 14px 36px color-mix(in srgb, var(--color-base) 70%, transparent)"
};

const tokenStyle = {
    margin: 0,
    wordBreak: "break-all",
    fontFamily: "monospace",
    backgroundColor: "var(--table-cell)",
    color: "var(--table-text)",
    border: "1px solid var(--table-border)",
    borderRadius: "8px",
    padding: "0.75rem"
};

const mockJwt = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJtZXJjaGFudF9pZCI6ImRlbW8tMTIzIiwic2NvcGUiOiJpbnZvaWNlOndyaXRlIiwiaWF0IjoxNzEwMDAwMDAwLCJleHAiOjE4NjAwMDAwMDB9.kfV9L42wP8o2nQ6uT1c8sB1yKj0h7mR3pD4aN5xY6zQ";

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
    const [showTokenModal, setShowTokenModal] = useState(false);
    const [copyMessage, setCopyMessage] = useState("");

    const storageKey = useMemo(() => {
        const merchantId = selectedMerchant?.id ?? selectedMerchant?.merchant_id ?? selectedMerchant?.name ?? "default";
        return `mock:merchant:api_keys:${merchantId}`;
    }, [selectedMerchant]);

    useEffect(() => {
        const raw = localStorage.getItem(storageKey);
        if (!raw) {
            setKeys([]);
            setStatusMessage("");
            setNewKeyName("");
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
        setStatusMessage("");
        setCopyMessage("");
        setShowTokenModal(true);
    };

    const onDeleteKey = (id) => {
        persistKeys(keys.filter((key) => key.id !== id));
        setStatusMessage("");
    };

    const onCopyToken = async () => {
        try {
            await navigator.clipboard.writeText(mockJwt);
        } catch {
            setCopyMessage("Copy failed. Please copy manually.");
        }
    };

    if (isLoading) {
        return <div style={{ padding: "2rem" }}>Loading...</div>;
    }

    return (
        <div style={{ padding: "2rem", maxWidth: "840px" }}>
            <h1>API Keys</h1>

            {showTokenModal && (
                <div style={modalOverlayStyle} role="dialog" aria-modal="true" aria-labelledby="api-key-modal-title">
                    <div style={modalCardStyle}>
                        <h2 id="api-key-modal-title" style={{ marginTop: 0, marginBottom: "0.75rem" }}>
                            Copy This Key:
                        </h2>
                        <p style={tokenStyle}>{mockJwt}</p>
                        {copyMessage && <p style={{ marginTop: "0.75rem", marginBottom: 0 }}>{copyMessage}</p>}
                        <div style={{ marginTop: "1rem", display: "flex", justifyContent: "flex-end", gap: "0.5rem" }}>
                            <button type="button" onClick={onCopyToken}>
                                Copy
                            </button>
                            <button
                                type="button"
                                onClick={() => {
                                    setShowTokenModal(false);
                                    setCopyMessage("");
                                }}
                            >
                                Close
                            </button>
                        </div>
                    </div>
                </div>
            )}

            {!selectedMerchant && <p style={{ marginBottom: "1rem" }}>Select a merchant account to manage API keys.</p>}

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
