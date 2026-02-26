import { Link } from "react-router-dom";
import { useMerchant } from "../../contexts/MerchantContext";

export default function Dashboard() {
    const { selectedMerchant, isLoading, requireRole } = useMerchant();

    if (isLoading) {
        return <div className="dashboard-page">Loading dashboard...</div>;
    }

    const hasMerchant = Boolean(selectedMerchant);
    const quickLinks = [
        { to: "/profile", label: "Profile", description: "View your account details" },
        { to: "/settings", label: "Settings", description: "Update preferences and security" },
        { to: "/webhooks", label: "Webhooks", description: "Configure notification endpoints" },
        { to: "/api-keys", label: "API Keys", description: "Create and rotate integration keys" },
    ];

    if (requireRole("Admin")) {
        quickLinks.push(
            { to: "/wallet", label: "Wallet", description: "Manage balances and transfers" },
            { to: "/users", label: "Users", description: "Control team member access" },
        );
    }

    const summaryCards = hasMerchant
        ? [
            { label: "Merchant", value: selectedMerchant.name, hint: "Active profile" },
            { label: "Merchant ID", value: selectedMerchant.id, hint: "Primary identifier" },
            { label: "Access Role", value: selectedMerchant.role, hint: "Current permission tier" },
            { label: "System Status", value: "Healthy", hint: "No incidents detected" },
        ]
        : [
            { label: "Account", value: "Ready", hint: "User account is active" },
            { label: "Merchant", value: "Not Linked", hint: "Connect a merchant to unlock tools" },
            { label: "Access", value: "Standard", hint: "Core navigation available" },
            { label: "Setup", value: "Pending", hint: "Complete merchant onboarding" },
        ];

    return (
        <div className="dashboard-page">
            <div className="dashboard-header">
                <div className="dashboard-title-block">
                    <h1 className="dashboard-title">Dashboard</h1>
                    <p className="dashboard-subtitle">Welcome back. Here is a live overview of your workspace and merchant operations.</p>
                </div>
                <div className="dashboard-chip">Realtime view</div>
            </div>

            <section className="dashboard-summary" aria-label="Summary metrics">
                {summaryCards.map((card) => (
                    <article className="dashboard-card" key={card.label}>
                        <p className="dashboard-card-label">{card.label}</p>
                        <p className="dashboard-card-value">{card.value}</p>
                        <p className="dashboard-card-hint">{card.hint}</p>
                    </article>
                ))}
            </section>

            {!hasMerchant ? (
                <section className="dashboard-empty-state">
                    <h2 className="dashboard-section-title">No merchant account linked yet</h2>
                    <p className="dashboard-empty-copy">
                        Your account is active. Once a merchant profile is added, dashboard metrics and tools will appear here.
                    </p>
                    <div className="dashboard-actions">
                        <Link to="/profile" className="dashboard-action-btn">View Profile</Link>
                        <Link to="/settings" className="dashboard-action-btn">Account Settings</Link>
                    </div>
                </section>
            ) : (
                <section className="dashboard-main-grid" aria-label="Merchant operations overview">
                    <div className="dashboard-panel dashboard-panel-primary">
                        <h2 className="dashboard-section-title">Quick actions</h2>
                        <div className="dashboard-actions-grid">
                            {quickLinks.map((item) => (
                                <Link to={item.to} className="dashboard-action-card" key={item.to}>
                                    <p className="dashboard-action-title">{item.label}</p>
                                    <p className="dashboard-action-copy">{item.description}</p>
                                </Link>
                            ))}
                        </div>
                    </div>

                    <div className="dashboard-panel">
                        <h2 className="dashboard-section-title">Recent activity</h2>
                        <ul className="dashboard-activity-list">
                            <li>
                                <p className="dashboard-activity-title">Merchant session verified</p>
                                <p className="dashboard-activity-meta">Now · Authentication service</p>
                            </li>
                            <li>
                                <p className="dashboard-activity-title">Webhook delivery monitoring enabled</p>
                                <p className="dashboard-activity-meta">Today · Integrations</p>
                            </li>
                            <li>
                                <p className="dashboard-activity-title">API key access reviewed</p>
                                <p className="dashboard-activity-meta">Yesterday · Security</p>
                            </li>
                        </ul>
                    </div>
                </section>
            )}
        </div>
    );
}
