import { useMerchant } from "../../contexts/MerchantContext";
import { useEffect, useState } from "react";

export default function Dashboard() {
    const { selectedMerchant, isLoading } = useMerchant();

    const [dashboardData, setDashboardData] = useState(null);
    const [dashboardLoading, setDashboardLoading] = useState(false);
    const [dashboardError, setDashboardError] = useState("");
    const [statusFilter, setStatusFilter] = useState("all");

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
                    {selectedMerchant.name} performance overview
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

                <div style={{ marginBottom: "1rem", display: "flex", gap: "0.5rem" }}>
                    <button
                        onClick={() => setStatusFilter("all")}
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
                        All Transactions
                    </button>
                    <button
                        onClick={() => setStatusFilter("completed")}
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
                </div>

                <table style={{ width: "100%", borderCollapse: "collapse" }}>
                    <thead>
                        <tr>
                            <th style={{ textAlign: "left", padding: "0.65rem", opacity: 0.7 }}>
                                Payment
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
                        {filteredActivity?.map((row) => (
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
            </div>
        </div>
    );
}