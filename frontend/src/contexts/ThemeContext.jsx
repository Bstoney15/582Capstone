// Author: Benjamin Stonestreet, Ryan Grimsley
// Created: 2026-02-20
// Last Updated: 2026-03-28
// Description: This context provides a way to toggle between light and dark themes for the application.


import { createContext, useContext, useState, useEffect } from "react";

const STORAGE_KEY = 'theme';

/**
 * Gets the system's preferred color scheme.
 * @returns {string} 'dark' or 'light'
 */
function getSystemTheme() {
    if (typeof window === 'undefined') return 'light';
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
}

/**
 * Gets the initial theme from local storage or falls back to system theme.
 * @returns {string} The initial theme to use.
 */
function getInitialTheme() {
    if (typeof window === 'undefined') return 'light';
    const saved = localStorage.getItem(STORAGE_KEY);
    if (saved === 'light' || saved === 'dark') return saved;
    return getSystemTheme();
}

/**
 * Applies the given theme to the document root element.
 * @param {string} theme The theme to apply ('light' or 'dark').
 */
function applyTheme(theme) {
    document.documentElement.setAttribute('data-theme', theme);
}

const ThemeContext = createContext();

/**
 * Provides theme state and toggling functionality to the application.
 * @param {Object} props
 * @param {React.ReactNode} props.children The child components.
 */
export function ThemeProvider({ children }) {
    const [theme, setThemeState] = useState(getInitialTheme);

    // Effect hook to apply theme changes to the DOM and save to local storage
    useEffect(() => {
        applyTheme(theme);
        localStorage.setItem(STORAGE_KEY, theme);
    }, [theme]);

    /**
     * Explicitly sets the theme.
     * @param {string} next The next theme to set.
     */
    const setTheme = (next) => {
        setThemeState(next === 'dark' ? 'dark' : 'light');
    };

    /**
     * Toggles the current theme between light and dark.
     */
    const toggleTheme = () => {
        setThemeState((prev) => (prev === 'dark' ? 'light' : 'dark'));
    };

    return (
        <ThemeContext.Provider value={{ theme, setTheme, toggleTheme }}>
            {children}
        </ThemeContext.Provider>
    );
}

/**
 * Custom hook to use the ThemeContext.
 * @returns {Object} The current theme context values.
 */
export function useTheme() {
    return useContext(ThemeContext);
}
