import { createContext, useContext, useState, useEffect } from "react";
import { useAuth } from "./AuthContext";

// Define role hierarchy to easily require minimum permission levels
const ROLE_HIERARCHY = {
    "Owner": 3,
    "Admin": 2,
    "Developer": 1
};

const MerchantContext = createContext();

export function MerchantProvider({ children }) {

    const { isAuthenticated } = useAuth();
    const [merchants, setMerchants] = useState([]);
    const [selectedMerchant, setSelectedMerchant] = useState(null);
    const [isLoading, setIsLoading] = useState(true);

    useEffect(() => {
        if (!isAuthenticated) return;

        const fetchMerchants = async () => {
            setIsLoading(true);
            try {
                const response = await fetch('/api/user/merchants');
                if (!response.ok) {
                    throw new Error(`Failed to fetch merchants: ${response.status}`);
                }

                const data = await response.json();

                setMerchants(data || []);
                if (data && data.length > 0) {
                    setSelectedMerchant(data[0]);
                }
            } catch (err) {
                console.error("Failed to fetch merchants:", err);
            } finally {
                setIsLoading(false);
            }
        };

        fetchMerchants();
    }, [isAuthenticated]);

    // Checks if the user's role on the selected merchant meets the required permission level
    const requireRole = (requiredRole) => {
        if (!selectedMerchant || !selectedMerchant.role) return false;

        const userLevel = ROLE_HIERARCHY[selectedMerchant.role] || 0;
        const requiredLevel = ROLE_HIERARCHY[requiredRole] || 0;

        return userLevel >= requiredLevel;
    };

    return (
        <MerchantContext.Provider value={{ merchants, selectedMerchant, setSelectedMerchant, isLoading, requireRole }}>
            {children}
        </MerchantContext.Provider>
    );
}

export function useMerchant() {
    return useContext(MerchantContext);
}
