// Dashboard.jsx – merchant dashboard page displaying volume stats, recent activity, and invoice search.
// Author: Ben Stonestreet, Joe Hotze, Ryan Grimsley
// Created: 2026-02-20

import { useMerchant } from "../../contexts/MerchantContext";
import { useEffect, useState } from "react";

/**
 * Dashboard renders the main merchant overview with gross volume statistics,
 * a recent activity table, invoice search, and pagination.
 * @returns {JSX.Element} The rendered Dashboard page.
 */
export default function Dashboard() {
    const { selectedMerchant, isLoading } = useMerchant();

    const [dashboardData, setDashboardData] = useState(null);
    const [dashboardLoading, setDashboardLoading] = useState(false);
    const [dashboardError, setDashboardError] = useState("");
    const [statusFilter, setStatusFilter] = useState("all");

    // Search state
    const [searchTerm, setSearchTerm] = useState("");
    const [searchResults, setSearchResults] = useState(null);
    const [searchLoading, setSearchLoading] = useState(false);
    const [searchError, setSearchError] = useState("");
    const [isSearching, setIsSearching] = useState(false);
    const [currentPage, setCurrentPage] = useState(1);

    // Fetch dashboard data whenever the selected merchant changes.
    useEffect(() => {
        if (!selectedMerchant?.id) {
            setDashboardData(null);
            setDashboardError("");
            setDashboardLoading(false);
            return;
        }

        const fetchDashboard = async () => {
            try {
                setDashboardLoading(true);
                setDashboardError("");

                const merchantID = encodeURIComponent(selectedMerchant.id);
                const res = await fetch(`/api/dashboard?merchant_id=${merchantID}`);

                if (!res.ok) {
                    if (res.status === 403) {
                        throw new Error("You do not have access to this merchant dashboard.");
                    }

                    throw new Error("Failed to fetch dashboard");
                }

                const data = await res.json();
                setDashboardData(data);
            } catch (err) {
                console.error("Dashboard error:", err);
                setDashboardData(null);
                setDashboardError(err instanceof Error ? err.message : "Failed to load dashboard.");
            } finally {
                setDashboardLoading(false);
            }
        };

        fetchDashboard();
    }, [selectedMerchant?.id]);

    /**
     * Submits the invoice search query and stores the matching results.
     * @param {React.FormEvent} e The form submit event.
     */
    const handleSearch = async (e) => {
        e.preventDefault();

        const trimmedSearchTerm = searchTerm.trim();
        if (!trimmedSearchTerm) {
            setSearchError("Please enter an invoice ID to search");
            return;
        }

        try {
            setSearchLoading(true);
            setSearchError("");
            setCurrentPage(1);

            const merchantID = encodeURIComponent(selectedMerchant.id);
            const searchQuery = encodeURIComponent(trimmedSearchTerm);
            const res = await fetch(
                `/api/dashboard/search?merchant_id=${merchantID}&invoice_id=${searchQuery}`
            );

            if (!res.ok) {
                if (res.status === 403) {
                    throw new Error("You do not have access to this merchant.");
                }
                throw new Error("Failed to search invoices");
            }

            const data = await res.json();
            setSearchResults(data || []);
            setIsSearching(true);
        } catch (err) {
            console.error("Search error:", err);
            setSearchError(err instanceof Error ? err.message : "Failed to search invoices.");
            setSearchResults([]);
        } finally {
            setSearchLoading(false);
        }
    };

    /**
     * Clears the active invoice search and restores the default activity view.
     */
    const handleClearSearch = () => {
        setSearchTerm("");
        setSearchResults(null);
        setSearchError("");
        setIsSearching(false);
        setCurrentPage(1);
    };

    const stats = dashboardData
        ? [
              {
                  label: "Gross Volume (30d)",
                  value: `${dashboardData.stats.grossVolume30d.toLocaleString()} XRP`,
              },
              {
                  label: "Gross Volume (6m)",
                  value: `${dashboardData.stats.grossVolume6m.toLocaleString()} XRP`,
              },
              {
                  label: "Gross Volume (1y)",
                  value: `${dashboardData.stats.grossVolume1y.toLocaleString()} XRP`,
              },
          ]
        : [];

    /**
     * Returns inline style overrides for a transaction status badge.
     * @param {string} status The invoice status string.
     * @returns {Object} Style object with color, background, and border.
     */
    const getStatusStyles = (status) => {
        if (status === "Settled") {
            return {
                color: "#166534",
                background: "rgba(34,197,94,0.15)",
                border: "1px solid rgba(34,197,94,0.35)",
            };
        }
        if (status === "Pending") {
            return {
                color: "#854d0e",
                background: "rgba(250,204,21,0.2)",
                border: "1px solid rgba(250,204,21,0.4)",
            };
        }
        return {
            color: "#991b1b",
            background: "rgba(248,113,113,0.15)",
            border: "1px solid rgba(248,113,113,0.35)",
        };
    };

    const filteredActivity = dashboardData?.recentActivity?.filter((row) => {
        if (statusFilter === "all") return true;
        if (statusFilter === "completed") return row.status === "Settled";
        return true;
    });

    // Determine which activity set to display: search results take priority over recent activity.
    const activityToDisplay = isSearching ? searchResults : dashboardData?.recentActivity;

    const filteredActivityForDisplay = activityToDisplay?.filter((row) => {
        if (statusFilter === "all") return true;
        if (statusFilter === "completed") return row.status === "Settled";
        return true;
    }) || [];

    const itemsPerPage = 20;
    const totalPages = Math.ceil(filteredActivityForDisplay.length / itemsPerPage);
    const paginatedActivity = filteredActivityForDisplay.slice(
        (currentPage - 1) * itemsPerPage,
        currentPage * itemsPerPage
    );

    if (isLoading) {
        return <div style={{ padding: "2rem" }}>Loading dashboard...</div>;
    }

    if (dashboardLoading) {
        return <div style={{ padding: "2rem" }}>Loading dashboard data...</div>;
    }

    if (!selectedMerchant) {
        return (
            <div style={{ padding: "2rem", maxWidth: "900px", margin: "0 auto" }}>
                <h1 style={{ margin: 0 }}>Dashboard</h1>
                <p style={{ marginTop: "0.75rem", opacity: 0.8 }}>
                    No merchant selected. Create a merchant or select one from the account dropdown to view dashboard stats.
                </p>
            </div>
        );
    }

    return (
        <div style={{ padding: "2rem", maxWidth: "1200px", margin: "0 auto" }}>
            <div style={{ marginBottom: "1.25rem" }}>
                <h1 style={{ margin: 0 }}>Dashboard</h1>
                <p style={{ margin: "0.4rem 0 0", opacity: 0.75 }}>
                    {selectedMerchant.name} performance overview <br/>
                    (Gross stats include pending transactions)
                </p>
            </div>

            {dashboardError && (
                <div style={{ marginBottom: "1rem", color: "#991b1b" }}>{dashboardError}</div>
            )}

            <div
                style={{
                    display: "grid",
                    gridTemplateColumns: "repeat(auto-fit, minmax(220px, 1fr))",
                    gap: "0.9rem",
                    marginBottom: "1rem",
                }}
            >
                {stats.map((stat) => (
                    <div
                        key={stat.label}
                        style={{
                            border: "1px solid rgba(128,128,128,0.25)",
                            borderRadius: "10px",
                            padding: "1rem",
                            background: "var(--color-base-variant)",
                        }}
                    >
                        <p style={{ margin: 0, fontSize: "0.9rem", opacity: 0.75 }}>
                            {stat.label}
                        </p>
                        <p
                            style={{
                                margin: "0.35rem 0 0",
                                fontSize: "1.45rem",
                                fontWeight: 700,
                            }}
                        >
                            {stat.value}
                        </p>
                    </div>
                ))}
            </div>

            <div
                style={{
                    border: "1px solid rgba(128,128,128,0.25)",
                    borderRadius: "10px",
                    background: "var(--color-base-variant)",
                    padding: "1rem",
                }}
            >
                <h2 style={{ margin: "0 0 0.75rem", fontSize: "1.05rem" }}>
                    Recent activity
                </h2>

                <div style={{ marginBottom: "1rem", display: "flex", gap: "0.5rem", flexWrap: "wrap", alignItems: "center" }}>
                    <button
                        onClick={() => {
                            setStatusFilter("all");
                            setCurrentPage(1);
                        }}
                        style={{
                            padding: "0.5rem 1rem",
                            borderRadius: "6px",
                            border: "1px solid rgba(128,128,128,0.25)",
                            background: statusFilter === "all" ? "var(--color-accent)" : "transparent",
                            color: statusFilter === "all" ? "white" : "inherit",
                            cursor: "pointer",
                            fontWeight: statusFilter === "all" ? 600 : 400,
                        }}
                    >
                        {isSearching ? "All Transactions" : "All Recent Transactions"}
                    </button>
                    <button
                        onClick={() => {
                            setStatusFilter("completed");
                            setCurrentPage(1);
                        }}
                        style={{
                            padding: "0.5rem 1rem",
                            borderRadius: "6px",
                            border: "1px solid rgba(128,128,128,0.25)",
                            background: statusFilter === "completed" ? "var(--color-accent)" : "transparent",
                            color: statusFilter === "completed" ? "white" : "inherit",
                            cursor: "pointer",
                            fontWeight: statusFilter === "completed" ? 600 : 400,
                        }}
                    >
                        Completed
                    </button>

                    {isSearching && (
                        <button
                            onClick={handleClearSearch}
                            style={{
                                padding: "0.5rem 1rem",
                                borderRadius: "6px",
                                border: "1px solid rgba(248,113,113,0.35)",
                                background: "rgba(248,113,113,0.15)",
                                color: "#991b1b",
                                cursor: "pointer",
                                fontWeight: 600,
                                marginLeft: "auto",
                            }}
                        >
                            Clear Search
                        </button>
                    )}
                </div>

                <form
                    onSubmit={handleSearch}
                    style={{ marginBottom: "1rem", display: "flex", gap: "0.5rem" }}
                >
                    <input
                        type="text"
                        placeholder="Search transactions by Invoice ID..."
                        value={searchTerm}
                        onChange={(e) => setSearchTerm(e.target.value)}
                        style={{
                            flex: 1,
                            padding: "0.5rem 0.75rem",
                            borderRadius: "6px",
                            border: "1px solid rgba(128,128,128,0.25)",
                            background: "var(--color-base)",
                            color: "inherit",
                            fontSize: "0.95rem",
                        }}
                        onKeyDown={(e) => {
                            if (e.key === "Enter") {
                                handleSearch(e);
                            }
                        }}
                    />
                    <button
                        type="submit"
                        disabled={searchLoading}
                        style={{
                            padding: "0.5rem 1rem",
                            borderRadius: "6px",
                            border: "1px solid rgba(128,128,128,0.25)",
                            background: "var(--color-accent)",
                            color: "white",
                            cursor: searchLoading ? "not-allowed" : "pointer",
                            fontWeight: 600,
                            opacity: searchLoading ? 0.7 : 1,
                        }}
                    >
                        {searchLoading ? "Searching..." : "Search"}
                    </button>
                </form>

                <table style={{ width: "100%", borderCollapse: "collapse" }}>
                    <thead>
                        <tr>
                            <th style={{ textAlign: "left", padding: "0.65rem", opacity: 0.7 }}>
                                Invoice ID
                            </th>
                            <th style={{ textAlign: "left", padding: "0.65rem", opacity: 0.7 }}>
                                Amount
                            </th>
                            <th style={{ textAlign: "left", padding: "0.65rem", opacity: 0.7 }}>
                                Status
                            </th>
                            <th style={{ textAlign: "left", padding: "0.65rem", opacity: 0.7 }}>
                                Date/Time
                            </th>
                        </tr>
                    </thead>
                    <tbody>
                        {paginatedActivity?.map((row) => (
                            <tr key={row.id}>
                                <td style={{ padding: "0.65rem" }}>{row.id}</td>
                                <td style={{ padding: "0.65rem" }}>
                                    {row.amount.toLocaleString()} XRP
                                </td>
                                <td style={{ padding: "0.65rem" }}>
                                    <span
                                        style={{
                                            ...getStatusStyles(row.status),
                                            display: "inline-block",
                                            padding: "0.15rem 0.5rem",
                                            borderRadius: "999px",
                                            fontSize: "0.78rem",
                                            fontWeight: 600,
                                        }}
                                    >
                                        {row.status}
                                    </span>
                                </td>
                                <td style={{ padding: "0.65rem" }}>{row.dateTime}</td>
                            </tr>
                        ))}
                    </tbody>
                </table>

                {/* Pagination controls and empty / error states */}
                {searchError && (
                    <div style={{ marginTop: "1rem", color: "#991b1b", fontSize: "0.95rem" }}>
                        {searchError}
                    </div>
                )}

                {isSearching && filteredActivityForDisplay.length === 0 && !searchError && (
                    <div style={{ marginTop: "1rem", opacity: 0.7, fontSize: "0.95rem" }}>
                        No invoices found matching "{searchTerm}"
                    </div>
                )}

                {filteredActivityForDisplay.length > itemsPerPage && (
                    <div style={{ marginTop: "1.5rem", display: "flex", gap: "1rem", alignItems: "center", justifyContent: "center" }}>
                        <button
                            onClick={() => setCurrentPage(Math.max(1, currentPage - 1))}
                            disabled={currentPage === 1}
                            style={{
                                padding: "0.5rem 1rem",
                                borderRadius: "6px",
                                border: "1px solid rgba(128,128,128,0.25)",
                                background: currentPage === 1 ? "rgba(128,128,128,0.1)" : "var(--color-accent)",
                                color: currentPage === 1 ? "rgba(128,128,128,0.5)" : "white",
                                cursor: currentPage === 1 ? "not-allowed" : "pointer",
                                fontWeight: 600,
                            }}
                        >
                            Previous
                        </button>
                        <span style={{ opacity: 0.75, fontSize: "0.95rem" }}>
                            Page {currentPage} of {totalPages}
                        </span>
                        <button
                            onClick={() => setCurrentPage(Math.min(totalPages, currentPage + 1))}
                            disabled={currentPage === totalPages}
                            style={{
                                padding: "0.5rem 1rem",
                                borderRadius: "6px",
                                border: "1px solid rgba(128,128,128,0.25)",
                                background: currentPage === totalPages ? "rgba(128,128,128,0.1)" : "var(--color-accent)",
                                color: currentPage === totalPages ? "rgba(128,128,128,0.5)" : "white",
                                cursor: currentPage === totalPages ? "not-allowed" : "pointer",
                                fontWeight: 600,
                            }}
                        >
                            Next
                        </button>
                    </div>
                )}
            </div>
        </div>
    );
}
