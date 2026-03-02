import { useMerchant } from "../../contexts/MerchantContext";

export default function Dashboard() {
    const { selectedMerchant, isLoading } = useMerchant();

    const stats = [
        { label: "Gross Volume (30d)", value: "126,420 XRP", note: "" },
        { label: "Gross Volume (6m)", value: "712,860 XRP", note: "" },
        { label: "Gross Volume (1y)", value: "1,487,300 XRP", note: "" },
    ];

    const recentActivity = [
        { id: "pay_10021", amount: "4,200 XRP", status: "Settled", dateTime: "2026-03-01 09:58 AM" },
        { id: "pay_10020", amount: "870 XRP", status: "Pending", dateTime: "2026-03-01 09:51 AM" },
        { id: "pay_10019", amount: "1,940 XRP", status: "Settled", dateTime: "2026-03-01 09:39 AM" },
        { id: "pay_10018", amount: "2,560 XRP", status: "Failed", dateTime: "2026-03-01 09:28 AM" },
        { id: "pay_10017", amount: "640 XRP", status: "Settled", dateTime: "2026-03-01 09:07 AM" },
    ];

    const getStatusStyles = (status) => {
        if (status === "Settled") {
            return { color: "#166534", background: "rgba(34,197,94,0.15)", border: "1px solid rgba(34,197,94,0.35)" };
        }
        if (status === "Pending") {
            return { color: "#854d0e", background: "rgba(250,204,21,0.2)", border: "1px solid rgba(250,204,21,0.4)" };
        }
        return { color: "#991b1b", background: "rgba(248,113,113,0.15)", border: "1px solid rgba(248,113,113,0.35)" };
    };

    if (isLoading) {
        return <div style={{ padding: "2rem" }}>Loading dashboard...</div>;
    }

    return (
        <div style={{ padding: "2rem", maxWidth: "1200px", margin: "0 auto" }}>
            <div style={{ marginBottom: "1.25rem" }}>
                <h1 style={{ margin: 0 }}>Dashboard</h1>
                <p style={{ margin: "0.4rem 0 0", opacity: 0.75 }}>
                    {selectedMerchant?.name || "No merchant selected"} performance overview
                </p>
            </div>

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
                        <p style={{ margin: 0, fontSize: "0.9rem", opacity: 0.75 }}>{stat.label}</p>
                        <p style={{ margin: "0.35rem 0 0", fontSize: "1.45rem", fontWeight: 700 }}>{stat.value}</p>
                        {stat.note ? (
                            <p style={{ margin: "0.35rem 0 0", fontSize: "0.85rem", opacity: 0.7 }}>{stat.note}</p>
                        ) : null}
                    </div>
                ))}
            </div>

            <div
                style={{
                    display: "grid",
                    gridTemplateColumns: "minmax(0, 2fr) minmax(0, 1fr)",
                    gap: "1rem",
                    alignItems: "start",
                }}
            >
                <div
                    style={{
                        border: "1px solid rgba(128,128,128,0.25)",
                        borderRadius: "10px",
                        background: "var(--color-base-variant)",
                        padding: "1rem",
                    }}
                >
                    <h2 style={{ margin: "0 0 0.75rem", fontSize: "1.05rem" }}>Recent activity</h2>
                    <div style={{ overflowX: "auto" }}>
                        <table style={{ width: "100%", borderCollapse: "collapse" }}>
                            <thead>
                                <tr>
                                    <th style={{ textAlign: "left", padding: "0.65rem", opacity: 0.7 }}>Payment</th>
                                    <th style={{ textAlign: "left", padding: "0.65rem", opacity: 0.7 }}>Amount</th>
                                    <th style={{ textAlign: "left", padding: "0.65rem", opacity: 0.7 }}>Status</th>
                                    <th style={{ textAlign: "left", padding: "0.65rem", opacity: 0.7 }}>Date/Time</th>
                                </tr>
                            </thead>
                            <tbody>
                                {recentActivity.map((row) => (
                                    <tr key={row.id}>
                                        <td style={{ padding: "0.65rem", borderTop: "1px solid rgba(128,128,128,0.18)" }}>{row.id}</td>
                                        <td style={{ padding: "0.65rem", borderTop: "1px solid rgba(128,128,128,0.18)" }}>{row.amount}</td>
                                        <td style={{ padding: "0.65rem", borderTop: "1px solid rgba(128,128,128,0.18)" }}>
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
                                        <td style={{ padding: "0.65rem", borderTop: "1px solid rgba(128,128,128,0.18)" }}>{row.dateTime}</td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                </div>

                <div style={{ display: "grid", gap: "1rem" }}>
                    <div
                        style={{
                            border: "1px solid rgba(128,128,128,0.25)",
                            borderRadius: "10px",
                            background: "var(--color-base-variant)",
                            padding: "1rem",
                        }}
                    >
                        <h2 style={{ margin: "0 0 0.75rem", fontSize: "1.05rem" }}>Merchant details</h2>
                        <p style={{ margin: "0.25rem 0", fontSize: "0.92rem" }}>
                            <strong>Merchant:</strong> {selectedMerchant?.name || "N/A"}
                        </p>
                        <p style={{ margin: "0.25rem 0", fontSize: "0.92rem" }}>
                            <strong>Merchant ID:</strong> {selectedMerchant?.id || "N/A"}
                        </p>
                        <p style={{ margin: "0.25rem 0", fontSize: "0.92rem" }}>
                            <strong>Your role:</strong> {selectedMerchant?.role || "N/A"}
                        </p>
                    </div>
                </div>
            </div>
        </div>
    );
}
