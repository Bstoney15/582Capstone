// Author: Benjamin Stonestreet
// Created: 2026-02-20
// Description: This component protects routes that require authentication.

import { useEffect } from "react";
import { Navigate, Outlet } from "react-router-dom";
import { useAuth } from "../contexts/AuthContext";
import { MerchantProvider } from "../contexts/MerchantContext";
import NavBar from "./NavBar";

/**
 * A wrapper component that protects routes requiring authentication.
 * Redirects to the login page if the user is not authenticated.
 * @returns {JSX.Element} The rendered protected route or redirect.
 */
export default function ProtectedRoute() {
    const { isAuthenticated, isLoading, checkAuthStatus } = useAuth();

    // Effect to re-check authentication on route visit
    useEffect(() => {
        // Re-check authentication when the protected route is visited/rendered
        checkAuthStatus();
    }, []);

    if (isLoading) {
        return <div style={{ display: "flex", justifyContent: "center", alignItems: "center", height: "100vh" }}>Loading...</div>; // Show loading state
    }

    if (!isAuthenticated) {
        return <Navigate to="/login" replace />; // Redirect to login if not authenticated
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
