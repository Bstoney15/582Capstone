// Authors: Ben Stonestreet, Ryan Grimsley
// Created: 02/20/26
// Description: wallet page where users can view and edit wallet associated with a merchant
import { useEffect, useState } from "react";
import { Navigate } from "react-router-dom";
import { useMerchant } from "../../contexts/MerchantContext";

export default function Wallet() {
    const { requireRole, isLoading, selectedMerchant } = useMerchant();
    const [walletAddress, setWalletAddress] = useState("");
    const [draftWalletAddress, setDraftWalletAddress] = useState("");
    const [editing, setEditing] = useState(true);

    const [message, setMessage] = useState("");
    const [error, setError] = useState("");

    useEffect(() => {
        async function fetchWallet() {
            try {
                setError("");
                setMessage("");

                const merchantId =
                    selectedMerchant?.id ??
                    selectedMerchant?.merchant_id;

                const params = new URLSearchParams({
                    merchant_id: merchantId
                });

                const response = await fetch(
                    `/api/merchant/get-wallet?${params}`
                );

                if (!response.ok) {
                    throw new Error();
                }

                const data = await response.json();
                setWalletAddress(data.wallet_address || "");
                setDraftWalletAddress(data.wallet_address || "");
                setEditing(!data.wallet_address);
            } catch {
                setError("Failed to load wallet.");
            }
        }

        if (selectedMerchant) {
            fetchWallet();
        }
    }, [selectedMerchant]);

    const onSubmitWallet = async (event) => {
        event.preventDefault();

        const nextWallet = draftWalletAddress.trim();

        setError("");
        setMessage("");

        if (!nextWallet) {
            setError("Wallet address is required.");
            return;
        }

        try {
            const merchantId =
                selectedMerchant?.id ??
                selectedMerchant?.merchant_id;

            const params = new URLSearchParams({
                merchant_id: merchantId
            });

            const response = await fetch(
                `/api/merchant/set-wallet?${params}`,
                {
                    method: "PATCH",
                    headers: {
                        "Content-Type": "application/json"
                    },
                    body: JSON.stringify({
                        wallet_address: nextWallet
                    })
                }
            );

            if (!response.ok) {
                throw new Error();
            }

            const data = await response.json();

            setWalletAddress(data.walletAddress ?? nextWallet);
            setEditing(false);
            setMessage("Wallet saved.");
        } catch {
            setError("Failed to save wallet.");
        }
    };

    if (isLoading) {
        return <div style={{ padding: "2rem" }}>Loading...</div>;
    }

    if (!requireRole("Admin")) {
        return <Navigate to="/dashboard" replace />;
    }

    if (!selectedMerchant) {
        return (
            <div style={{ padding: "2rem" }}>
                <h1>Wallet</h1>
                <p>Select a merchant account to manage wallet details.</p>
            </div>
        );
    }

    return (
        <div style={{ padding: "2rem", maxWidth: "760px" }}>
            <h1>Wallet</h1>
            <p style={{ marginBottom: "1rem" }}>
                The wallet address where this merchant receives customer payments.
            </p>

            {error && (
                <p style={{ color: "#b91c1c", marginBottom: "1rem" }}>
                    {error}
                </p>
            )}

            {message && (
                <p style={{ color: "#166534", marginBottom: "1rem" }}>
                    {message}
                </p>
            )}

            {!editing && walletAddress ? (
                <div style={{ border: "1px solid #d1d5db", borderRadius: "8px", padding: "1rem" }}>
                    <p style={{ marginBottom: "0.5rem" }}>
                        <strong>Current wallet address</strong>
                    </p>
                    <code style={{ display: "block", marginBottom: "1rem", wordBreak: "break-all" }}>
                        {walletAddress}
                    </code>
                    <div style={{ display: "flex", gap: "0.75rem", flexWrap: "wrap" }}>
                        <button type="button" onClick={() => {
                            setEditing(true);
                            setError("");
                            setMessage("");
                        }}>
                            Replace Wallet
                        </button>

                        <a
                            href={`https://livenet.xrpl.org/accounts/${encodeURIComponent(walletAddress)}`}
                            target="_blank"
                            rel="noreferrer"
                        >
                            View on XRPL Explorer
                        </a>
                    </div>
                </div>
            ) : (
                <form onSubmit={onSubmitWallet}>
                    <label htmlFor="walletAddress" style={{ display: "block", marginBottom: "0.5rem" }}>
                        Wallet Address
                    </label>
                    <input
                        id="walletAddress"
                        type="text"
                        value={draftWalletAddress}
                        onChange={(event) => setDraftWalletAddress(event.target.value)}
                        placeholder="Enter wallet address"
                        style={{ width: "100%", padding: "0.625rem", marginBottom: "0.75rem" }}
                    />

                    <div style={{ display: "flex", gap: "0.75rem", flexWrap: "wrap" }}>
                        <button type="submit">
                            {walletAddress ? "Replace Wallet" : "Save Wallet"}
                        </button>

                        {walletAddress && (
                            <button
                                type="button"
                                onClick={() => {
                                    setDraftWalletAddress(walletAddress);
                                    setEditing(false);
                                    setError("");
                                    setMessage("");
                                }}
                            >
                                Cancel
                            </button>
                        )}
                    </div>
                </form>
            )}
        </div>
    );
}