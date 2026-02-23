import { useEffect } from "react";
import { Navigate, Outlet } from "react-router-dom";
import { useAuth } from "../contexts/AuthContext";
import { MerchantProvider } from "../contexts/MerchantContext";
import NavBar from "./NavBar";

export default function ProtectedRoute() {
    const { isAuthenticated, isLoading, checkAuthStatus } = useAuth();

    useEffect(() => {
        // Re-check authentication when the protected route is visited/rendered
        checkAuthStatus();
    }, []);

    if (isLoading) {
        return <div style={{ display: "flex", justifyContent: "center", alignItems: "center", height: "100vh" }}>Loading...</div>;
    }

    if (!isAuthenticated) {
        return <Navigate to="/login" replace />;
    }

    return (
        <MerchantProvider>
            <NavBar />
            <div className="authenticated-content-wrapper">
                <Outlet />
            </div>
        </MerchantProvider>
    );
}
