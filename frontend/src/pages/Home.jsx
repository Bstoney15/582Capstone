import { Link } from "react-router-dom";

export default function Home({ theme, toggleTheme }) {
  return (
    <div className="auth-page">
      {toggleTheme && (
        <button
          type="button"
          className="theme-toggle"
          onClick={toggleTheme}
          title={`Switch to ${theme === "dark" ? "light" : "dark"} mode`}
          aria-label={`Switch to ${theme === "dark" ? "light" : "dark"} mode`}
        >
          {theme === "dark" ? "☀️" : "🌙"}
        </button>
      )}

      <div className="auth-card home-card">
        <h1 className="auth-title">XRPay</h1>
        <p className="auth-subtitle">Accept XRP. Receive USDC.</p>

        <div className="home-content">
          <div className="home-feature">
            <span className="home-icon">💱</span>
            <p>Customers pay in XRP. You receive USDC automatically—no crypto to manage.</p>
          </div>

          <div className="home-feature">
            <span className="home-icon">⚡</span>
            <p>Drop a lightweight widget on your checkout page and start accepting payments in minutes.</p>
          </div>

          <div className="home-feature">
            <span className="home-icon">📊</span>
            <p>Track transactions, manage API keys, and configure webhooks from one simple dashboard.</p>
          </div>
        </div>

        <div className="home-actions">
          <Link className="home-button home-button-primary" to="/signup">
            Create Account
          </Link>
          <Link className="home-button home-button-secondary" to="/login">
            Log In
          </Link>
        </div>
      </div>
    </div>
  );
}