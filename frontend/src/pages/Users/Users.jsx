import { Navigate } from "react-router-dom";
import { useMerchant } from "../../contexts/MerchantContext";

export default function Users() {
    const { requireRole, isLoading } = useMerchant();

    if (isLoading) {
        return <div style={{ padding: "2rem" }}>Loading...</div>;
    }

    if (!requireRole("Admin")) {
        return <Navigate to="/dashboard" replace />;
    }

    return (
        <div style={{ padding: "2rem" }}>
            <h1>Users</h1>
            <p>Welcome to the Users management page. This content is a placeholder.</p>
        </div>
    );
}
