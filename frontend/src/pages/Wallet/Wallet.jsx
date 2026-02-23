import { Navigate } from "react-router-dom";
import { useMerchant } from "../../contexts/MerchantContext";

export default function Wallet() {
    const { requireRole, isLoading } = useMerchant();

    if (isLoading) {
        return <div style={{ padding: "2rem" }}>Loading...</div>;
    }

    if (!requireRole("Admin")) {
        return <Navigate to="/dashboard" replace />;
    }

    return (
        <div style={{ padding: "2rem" }}>
            <h1>Wallet</h1>
            <p>Welcome to the Wallet management page. This content is a placeholder.</p>
        </div>
    );
}
