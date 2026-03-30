import { useEffect, useState } from "react";
import { useMerchant } from "../../contexts/MerchantContext";
import { createApiKey, deleteApiKey, fetchApiKeys } from "../../services/apiKeyService";

export default function ApiKeys() {
    const { selectedMerchant, isLoading } = useMerchant();
    const [keys, setKeys] = useState([]);
    const [newKeyName, setNewKeyName] = useState("");
    const [status, setStatus] = useState(null);
    const [isLoadingKeys, setIsLoadingKeys] = useState(false);
    const [isGenerating, setIsGenerating] = useState(false);
    const [deletingId, setDeletingId] = useState(null);
    const [showTokenModal, setShowTokenModal] = useState(false);
    const [latestToken, setLatestToken] = useState("");
    const [copyMessage, setCopyMessage] = useState("");

    useEffect(() => {
        const loadKeys = async () => {
            if (!selectedMerchant) {
                setKeys([]);
                setStatus(null);
                setNewKeyName("");
                return;
            }

            setIsLoadingKeys(true);
            try {
                const data = await fetchApiKeys(selectedMerchant);
                setKeys(Array.isArray(data) ? data : []);
                setStatus(null);
            } catch (error) {
                console.error("Failed to fetch API keys", error);
                setStatus({ type: "error", message: "Failed to load API keys." });
                setKeys([]);
            } finally {
                setIsLoadingKeys(false);
            }
        };

        setShowTokenModal(false);
        setLatestToken("");
        setCopyMessage("");
        setNewKeyName("");
        loadKeys();
    }, [selectedMerchant]);

    const loadKeys = async () => {
        if (!selectedMerchant) {
            setKeys([]);
            setStatus(null);
            return;
        }

        try {
            const data = await fetchApiKeys(selectedMerchant);
            setKeys(Array.isArray(data) ? data : []);
        } catch {
            setKeys([]);
            setStatus({ type: "error", message: "Failed to load API keys." });
        }
    };

    const onGenerateKey = async (event) => {
        event.preventDefault();
        const name = newKeyName.trim();

        if (!name) {
            setStatus({ type: "error", message: "API key name is required." });
            return;
        }

        setIsGenerating(true);
        setStatus(null);

        try {
            const created = await createApiKey(selectedMerchant, { name });
            await loadKeys();
            setNewKeyName("");
            setLatestToken(created.api_key || "");
            setCopyMessage("");
            setShowTokenModal(Boolean(created.api_key));
            setStatus({ type: "success", message: "API key generated." });
        } catch (error) {
            console.error("Failed to create API key", error);
            setStatus({ type: "error", message: "Failed to generate API key." });
        } finally {
            setIsGenerating(false);
        }
    };

    const onDeleteKey = async (id) => {
        if (!window.confirm("Are you sure you want to delete this API key?")) return;

        setDeletingId(id);
        setStatus(null);
        try {
            await deleteApiKey(selectedMerchant, id);
            await loadKeys();
            setStatus({ type: "success", message: "API key deleted." });
        } catch (error) {
            console.error("Failed to delete API key", error);
            setStatus({ type: "error", message: "Failed to delete API key." });
        } finally {
            setDeletingId(null);
        }
    };

    const onCopyToken = async () => {
        try {
            await navigator.clipboard.writeText(latestToken);
            setCopyMessage("Copied to clipboard.");
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
                        <p style={tokenStyle}>{latestToken}</p>
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

            {status && (
                <p style={{ color: status.type === "error" ? "#b91c1c" : "#166534", marginBottom: "1rem", wordBreak: "break-word" }}>
                    {status.message}
                </p>
            )}

            {selectedMerchant && (
                <form onSubmit={onGenerateKey} style={{ marginBottom: "1.25rem", display: "flex", gap: "0.75rem", flexWrap: "wrap" }}>
                    <input
                        type="text"
                        value={newKeyName}
                        onChange={(event) => setNewKeyName(event.target.value)}
                        placeholder="Key name (e.g. Production Checkout)"
                        disabled={isGenerating}
                        style={{ minWidth: "260px", flex: "1", padding: "0.625rem" }}
                    />
                    <button type="submit" disabled={isGenerating}>{isGenerating ? "Generating..." : "Generate New Key"}</button>
                </form>
            )}

            {selectedMerchant && isLoadingKeys && <p>Loading API keys...</p>}

            {selectedMerchant && !isLoadingKeys && keys.length === 0 && <p>No API keys yet.</p>}

            {selectedMerchant && !isLoadingKeys && keys.length > 0 && (
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
                                        <button type="button" disabled={deletingId === key.id} onClick={() => onDeleteKey(key.id)}>
                                            {deletingId === key.id ? "Deleting..." : "Delete"}
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