import { createContext, useContext, useState, useEffect } from "react";

const AuthContext = createContext();

export function AuthProvider({ children }) {
    const [isAuthenticated, setIsAuthenticated] = useState(false);
    const [isLoading, setIsLoading] = useState(true);

    const redirectAuthenticatedUser = () => {
        const publicEntryRoutes = new Set(["/", "/login", "/signup"]);
        const currentPath = window.location.pathname;

        if (publicEntryRoutes.has(currentPath)) {
            window.history.replaceState(null, "", "/dashboard");
            window.dispatchEvent(new PopStateEvent("popstate"));
        }
    };

    const checkAuthStatus = async () => {
        try {
            const response = await fetch("/api/user/auth");
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
            setIsLoading(false);
        }
    };

    useEffect(() => {
        checkAuthStatus();

        const intervalId = window.setInterval(() => {
            checkAuthStatus();
        }, 30000);

        return () => {
            window.clearInterval(intervalId);
        };
    }, []);

    return (
        <AuthContext.Provider value={{ isAuthenticated, setIsAuthenticated, isLoading, checkAuthStatus }}>
            {children}
        </AuthContext.Provider>
    );
}

export function useAuth() {
    return useContext(AuthContext);
}
