// Authors: Ben Stonestreet, Charley Findling, Ryan Grimsley
// Created: 02/14/26
// Description: Home landing page that new users that are not logged in see
import { Link } from "react-router-dom";
import { useTheme } from "../../contexts/ThemeContext";
import {
  SunIcon,
  MoonIcon,
  ArrowsRightLeftIcon,
  BoltIcon,
  ChartBarSquareIcon,
} from "@heroicons/react/24/outline";
import "./Home.css";

export default function Home() {
  const { theme, toggleTheme } = useTheme();

  return (
    <div className="landing">
      {/* Header for landing page */}
      <header className="landing-header">
        <span className="landing-logo">XRPay</span>
        <div className="landing-header-actions">
          {/* log in button */}
          <Link className="landing-nav-link" to="/login">
            Log in
          </Link>
          {/* create account button */}
          <Link className="landing-btn landing-btn-primary" to="/signup">
            Create account
          </Link>
          {/* toggle theme button */}
          {toggleTheme && (
            <button
              type="button"
              className="landing-theme-toggle"
              onClick={toggleTheme}
              title={`Switch to ${theme === "dark" ? "light" : "dark"} mode`}
              aria-label={`Switch to ${theme === "dark" ? "light" : "dark"} mode`}
            >
              {theme === "dark" ? (
                <SunIcon className="landing-theme-icon" aria-hidden />
              ) : (
                <MoonIcon className="landing-theme-icon" aria-hidden />
              )}
            </button>
          )}
        </div>
      </header>
          {/* main parts of home page */}
      <main className="landing-main">
        <section className="landing-hero" aria-labelledby="landing-hero-heading">
          <div className="landing-hero-inner">
            <h1 id="landing-hero-heading" className="landing-hero-title">
              Accept XRP at checkout
            </h1>
            <p className="landing-hero-lead">
              Customers pay in XRP. Your business receives XRP directly in the XRP wallet you
              configure—simple payouts without changing how you think about settlement.
            </p>
            {/* more create acc and login links */}
            <div className="landing-hero-ctas">
              <Link className="landing-btn landing-btn-primary" to="/signup">
                Create account
              </Link>
              <Link className="landing-btn landing-btn-secondary" to="/login">
                Log in
              </Link>
            </div>
          </div>
        </section>
          {/* cards in page */}
        <section className="landing-section" aria-labelledby="features-heading">
          <h2 id="features-heading" className="landing-section-title">
            Built for merchants
          </h2>
          <div className="landing-features">
            <article className="landing-feature-card">
              <div className="landing-feature-icon-wrap" aria-hidden>
                <ArrowsRightLeftIcon className="landing-feature-icon" />
              </div>
              <h3>XRP in, XRP out</h3>
              <p>
                Shoppers pay with XRP. You receive XRP to your own wallet address—no surprise
                conversions in the default flow.
              </p>
            </article>
            <article className="landing-feature-card">
              <div className="landing-feature-icon-wrap" aria-hidden>
                <BoltIcon className="landing-feature-icon" />
              </div>
              <h3>Lightweight checkout widget</h3>
              <p>
                Add the payment widget to your checkout page and start accepting XRP without a heavy
                integration.
              </p>
            </article>
            <article className="landing-feature-card">
              <div className="landing-feature-icon-wrap" aria-hidden>
                <ChartBarSquareIcon className="landing-feature-icon" />
              </div>
              <h3>Dashboard &amp; APIs</h3>
              <p>
                Track transactions, manage API keys, and configure webhooks from one place alongside
                your XRP wallet settings.
              </p>
            </article>
          </div>
        </section>
          {/* "how it works" section */}
        <section className="landing-section" aria-labelledby="how-heading">
          <h2 id="how-heading" className="landing-section-title">
            How it works
          </h2>
          <div className="landing-steps">
            <div className="landing-step">
              <div className="landing-step-num">1</div>
              <p>Embed the XRPay widget on your checkout and connect your merchant XRP wallet.</p>
            </div>
            <div className="landing-step">
              <div className="landing-step-num">2</div>
              <p>Your customer completes payment in XRP through the widget.</p>
            </div>
            <div className="landing-step">
              <div className="landing-step-num">3</div>
              <p>XRP settles to your configured wallet; you monitor activity in the dashboard.</p>
            </div>
          </div>
        </section>
      </main>
          {/* footer */}
      <footer className="landing-footer">
        <div className="landing-footer-inner">
          <p className="landing-footer-meta">XRPay</p>
          <div className="landing-footer-links">
            <Link className="landing-footer-link" to="/login">
              Log in
            </Link>
            <Link className="landing-footer-link" to="/signup">
              Sign up
            </Link>
          </div>
          <p className="landing-footer-dev">
            <span>Developers · </span>
            <Link to="/widget-demo">Widget demo</Link>
          </p>
        </div>
      </footer>
    </div>
  );
}
