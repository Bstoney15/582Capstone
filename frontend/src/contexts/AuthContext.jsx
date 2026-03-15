// Author: Benjamin Stonestreet
// Created: 2026-02-21


import { createContext, useContext, useState, useEffect } from "react";

const AuthContext = createContext();

/**
 * Provides authentication context to the application, managing user session state.
 * @param {Object} props
 * @param {React.ReactNode} props.children The child components.
 */
export function AuthProvider({ children }) {
    const [isAuthenticated, setIsAuthenticated] = useState(false);
    const [isLoading, setIsLoading] = useState(true);

    /**
     * Redirects authenticated users away from public entry points (like login/signup)
     * to the dashboard.
     */
    const redirectAuthenticatedUser = () => {
        const publicEntryRoutes = new Set(["/", "/login", "/signup"]); // Public routes
        const currentPath = window.location.pathname;

        if (publicEntryRoutes.has(currentPath)) {
            window.history.replaceState(null, "", "/dashboard"); // Redirect to dashboard
            window.dispatchEvent(new PopStateEvent("popstate"));
        }
    };

    /**
     * Checks the authentication status from the backend API and updates state.
     */
    const checkAuthStatus = async () => {
        try {
            const response = await fetch("/api/user/auth"); // Call auth API
            if (response.ok) {
                const data = await response.json();
                setIsAuthenticated(data.authenticated);
                if (data.authenticated) {
                    redirectAuthenticatedUser();
                }
            } else {
                setIsAuthenticated(false);
            }
        } catch (err) {
            console.error("Auth check failed:", err);
            setIsAuthenticated(false);
        } finally {
            setIsLoading(false); // Mark loading as complete
        }
    };

    // Effect hook to check auth status on mount and at regular intervals
    useEffect(() => {
        checkAuthStatus(); // Initial check

        // Set up polling for auth status every 30 seconds
        const intervalId = window.setInterval(() => {
            checkAuthStatus();
        }, 30000);

        return () => {
            window.clearInterval(intervalId); // Cleanup interval on unmount
        };
    }, []);

    return (
        <AuthContext.Provider value={{ isAuthenticated, setIsAuthenticated, isLoading, checkAuthStatus }}>
            {children}
        </AuthContext.Provider>
    );
}

/**
 * Custom hook to use the AuthContext.
 * @returns {Object} The current auth context values.
 */
export function useAuth() {
    return useContext(AuthContext);
}
