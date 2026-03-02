import { Link } from "react-router-dom";
import { useTheme } from "../../contexts/ThemeContext";

export default function Home() {
  const { theme, toggleTheme } = useTheme();

  return (
    <div className="auth-page home-page">
      {toggleTheme && (
        <button
          type="button"
          className="theme-toggle"
          onClick={toggleTheme}
          title={`Switch to ${theme === "dark" ? "light" : "dark"} mode`}
          aria-label={`Switch to ${theme === "dark" ? "light" : "dark"} mode`}
        >
          {theme === "dark" ? "Light" : "Dark"}
        </button>
      )}

      <div className="auth-card home-card">
        <h1 className="auth-title">XRPay</h1>
        <p className="auth-subtitle">Choose an option to continue</p>
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