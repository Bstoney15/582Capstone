// Author: Benjamin Stonestreet
// Created: 2026-02-20
// Description: This component is the navigation bar for the application.

import { useState, useRef, useEffect } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useAuth } from "../contexts/AuthContext";
import { useMerchant } from "../contexts/MerchantContext";
import { useTheme } from "../contexts/ThemeContext";
import Dropdown from "./Dropdown";
import { UserCircleIcon, SunIcon, MoonIcon } from "@heroicons/react/24/outline";
import "./NavBar.css";

/**
 * Navigation Bar component displaying links, merchant selector, and user profile menu.
 * @returns {JSX.Element} The rendered NavBar.
 */
export default function NavBar() {
    const { logout } = useAuth();
    const { merchants, selectedMerchant, setSelectedMerchant, requireRole } = useMerchant();
    const { theme, toggleTheme } = useTheme();
    const navigate = useNavigate();

    const [isProfileOpen, setIsProfileOpen] = useState(false);
    const profileRef = useRef(null);

    // Effect to close the profile menu when clicking outside of it
    useEffect(() => {
        /**
         * Handles clicks outside the profile dropdown to close it.
         * @param {MouseEvent} event The mouse event.
         */
        function handleClickOutside(event) {
            if (profileRef.current && !profileRef.current.contains(event.target)) {
                setIsProfileOpen(false); // Close the menu
            }
        }
        document.addEventListener("mousedown", handleClickOutside);
        return () => {
            document.removeEventListener("mousedown", handleClickOutside); // Cleanup listener
        };
    }, []);

    /**
     * Handles logout by clearing the client session state and redirecting.
     */
    const handleLogout = async () => {
        await logout();
    };

    return (
        <nav className="navbar-container">
            <div className="navbar-brand">XRPay</div>
            <div className="navbar-links">
                <Link to="/dashboard" className="navbar-link">Dashboard</Link>
                <Link to="/api-docs" className="navbar-link">API Docs</Link>
                {selectedMerchant && (
                    <>
                        {requireRole("Admin") && (
                            <>
                                <Link to="/wallet" className="navbar-link">Wallet</Link>
                                <Link to="/users" className="navbar-link">Users</Link>
                            </>
                        )}
                        <Link to="/customers" className="navbar-link">Customers</Link>
                        <Link to="/webhooks" className="navbar-link">Webhooks</Link>
                        <Link to="/api-keys" className="navbar-link">API Keys</Link>
                    </>
                )}
            </div>

            <div className="navbar-actions">

                {merchants && merchants.length > 0 && (
                    <Dropdown
                        options={merchants}
                        selectedItem={selectedMerchant}
                        onSelect={setSelectedMerchant}
                        placeholder="Select Account"
                    />
                )}

                <div className="profile-menu-container" ref={profileRef}>
                    <button
                        className="profile-menu-btn"
                        onClick={() => setIsProfileOpen(!isProfileOpen)}
                        aria-label="User Profile Menu"
                        aria-expanded={isProfileOpen}
                        aria-haspopup="menu"
                    >
                        <UserCircleIcon className="profile-icon" />
                    </button>

                    {isProfileOpen && (
                        <div className="custom-dropdown-menu profile-dropdown-menu" role="menu">
                            <button
                                className="custom-dropdown-item"
                                role="menuitem"
                                onClick={() => { setIsProfileOpen(false); navigate("/profile"); }}
                            >
                                Profile
                            </button>
                            <button
                                className="custom-dropdown-item"
                                role="menuitem"
                                onClick={() => { setIsProfileOpen(false); navigate("/settings"); }}
                            >
                                Settings
                            </button>
                            <div className="dropdown-divider"></div>
                            <button
                                className="custom-dropdown-item"
                                role="menuitem"
                                onClick={() => { setIsProfileOpen(false); navigate("/create-merchant"); }}
                            >
                                Create New Merchant
                            </button>
                            <div className="dropdown-divider"></div>
                            <button
                                className="custom-dropdown-item"
                                role="menuitem"
                                onClick={(e) => {
                                    e.stopPropagation();
                                    toggleTheme();
                                }}
                                style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}
                            >
                                <span>{theme === 'dark' ? 'Light Mode' : 'Dark Mode'}</span>
                                {theme === 'dark' ? (
                                    <SunIcon style={{ width: '1.2rem', height: '1.2rem' }} />
                                ) : (
                                    <MoonIcon style={{ width: '1.2rem', height: '1.2rem' }} />
                                )}
                            </button>
                            <div className="dropdown-divider"></div>
                            <button
                                className="custom-dropdown-item text-danger"
                                role="menuitem"
                                onClick={() => { setIsProfileOpen(false); void handleLogout(); }}
                            >
                                Logout
                            </button>
                        </div>
                    )}
                </div>

            </div>
        </nav>
    );
}
