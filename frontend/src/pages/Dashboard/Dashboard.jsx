import { Link } from "react-router-dom";
import { useMerchant } from "../../contexts/MerchantContext";

export default function Dashboard() {
    const { selectedMerchant, isLoading, requireRole } = useMerchant();

    if (isLoading) {
        return <div className="dashboard-page">Loading dashboard...</div>;
    }

    const hasMerchant = Boolean(selectedMerchant);

    return (
        <div className="dashboard-page">
            <div className="dashboard-header">
                <h1 className="dashboard-title">Dashboard</h1>
                <p className="dashboard-subtitle">Welcome back. Here is a quick overview of your account.</p>
            </div>

            {!hasMerchant ? (
                <section className="dashboard-empty-state">
                    <h2 className="dashboard-section-title">No merchant account linked yet</h2>
                    <p className="dashboard-empty-copy">
                        Your account is active. Once a merchant profile is added, dashboard metrics and tools will appear here.
                    </p>
                    <div className="dashboard-actions">
                        <Link to="/dashboard" className="dashboard-action-btn">Dashboard</Link>
                        <Link to="/profile" className="dashboard-action-btn">View Profile</Link>
                        <Link to="/settings" className="dashboard-action-btn">Account Settings</Link>
                    </div>
                </section>
            ) : (
                <>
                    <section className="dashboard-grid" aria-label="Merchant overview">
                        <article className="dashboard-card">
                            <p className="dashboard-card-label">Merchant</p>
                            <p className="dashboard-card-value">{selectedMerchant.name}</p>
                        </article>
                        <article className="dashboard-card">
                            <p className="dashboard-card-label">Merchant ID</p>
                            <p className="dashboard-card-value">{selectedMerchant.id}</p>
                        </article>
                        <article className="dashboard-card">
                            <p className="dashboard-card-label">Role</p>
                            <p className="dashboard-card-value">{selectedMerchant.role}</p>
                        </article>
                    </section>

                    <section className="dashboard-section">
                        <h2 className="dashboard-section-title">All pages</h2>
                        <div className="dashboard-actions">
                            <Link to="/dashboard" className="dashboard-action-btn">Dashboard</Link>
                            <Link to="/profile" className="dashboard-action-btn">Profile</Link>
                            <Link to="/settings" className="dashboard-action-btn">Settings</Link>
                            <Link to="/webhooks" className="dashboard-action-btn">Manage Webhooks</Link>
                            <Link to="/api-keys" className="dashboard-action-btn">View API Keys</Link>
                            {requireRole("Admin") && (
                                <>
                                    <Link to="/wallet" className="dashboard-action-btn">Open Wallet</Link>
                                    <Link to="/users" className="dashboard-action-btn">Manage Users</Link>
                                </>
                            )}
                        </div>
                    </section>
                </>
            )}
        </div>
    );
}
