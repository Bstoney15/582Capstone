// Author: Benjamin Stonestreet
// Created: 2026-02-28
// Description: This context provides a way to manage merchant data and permissions for the application.


import { createContext, useContext, useState, useEffect } from "react";
import { useAuth } from "./AuthContext";

// Define role hierarchy to easily require minimum permission levels
const ROLE_HIERARCHY = {
    "Owner": 3,
    "Admin": 2,
    "Developer": 1
};

const MerchantContext = createContext();

/**
 * Provides merchant context, managing the user's available merchants and selected merchant.
 * @param {Object} props
 * @param {React.ReactNode} props.children The child components.
 */
export function MerchantProvider({ children }) {

    const { isAuthenticated } = useAuth();
    const [merchants, setMerchants] = useState([]);
    const [selectedMerchant, setSelectedMerchant] = useState(null);
    const [isLoading, setIsLoading] = useState(true);

    // Effect hook to fetch merchants when the user becomes authenticated
    useEffect(() => {
        if (!isAuthenticated) return;

        /**
         * Fetches merchants from the API and sets the initial selected merchant.
         */
        const fetchMerchants = async () => {
            setIsLoading(true); // Start loading state
            try {
                const response = await fetch('/api/user/merchants');
                if (!response.ok) {
                    throw new Error(`Failed to fetch merchants: ${response.status}`);
                }

                const data = await response.json();

                setMerchants(data || []);
                if (data && data.length > 0) {
                    setSelectedMerchant(data[0]); // Select first merchant by default
                }
            } catch (err) {
                console.error("Failed to fetch merchants:", err);
            } finally {
                setIsLoading(false); // End loading state
            }
        };

        fetchMerchants();
    }, [isAuthenticated]);

    /**
     * Checks if the user's role on the selected merchant meets the required permission level.
     * @param {string} requiredRole The role required to pass the check.
     * @returns {boolean} True if the user has the required role, false otherwise.
     */
    const requireRole = (requiredRole) => {
        if (!selectedMerchant || !selectedMerchant.role) return false;

        const userLevel = ROLE_HIERARCHY[selectedMerchant.role] || 0;
        const requiredLevel = ROLE_HIERARCHY[requiredRole] || 0;

        return userLevel >= requiredLevel; // Compare hierarchy levels
    };

    return (
        <MerchantContext.Provider value={{ merchants, selectedMerchant, setSelectedMerchant, isLoading, requireRole }}>
            {children}
        </MerchantContext.Provider>
    );
}

/**
 * Custom hook to use the MerchantContext.
 * @returns {Object} The current merchant context values.
 */
export function useMerchant() {
    return useContext(MerchantContext);
}
