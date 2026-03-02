import { useState } from "react";
import { Link } from "react-router-dom";

export default function Kyb() {
  const [form, setForm] = useState({
    // merchant
    merchant_name: "",

    // merchant_business_profile
    merchant_business_profile_dba_name: "",
    merchant_business_profile_registration_number: "",
    merchant_business_profile_tax_id: "",
    merchant_business_profile_website_url: "",
    merchant_business_profile_incoporation_date: "", // matches model tag (note spelling)
    merchant_business_profile_legal_structure: "llc",
    merchant_business_profile_mcc_code: "",
    merchant_business_profile_phone_number: "",
    merchant_business_profile_email: "",

    // merchant_address
    merchant_address_line_1: "",
    merchant_address_line_2: "",
    merchant_address_city: "",
    merchant_address_state: "",
    merchant_address_postal_code: "",
    merchant_address_verified: false, // typically backend-controlled

    // merchant_owner
    merchant_owner_first_name: "",
    merchant_owner_last_name: "",
    merchant_owner_phone_number: "",
    merchant_owner_email: "",
    merchant_owner_dob: "", // send as ISO date string (YYYY-MM-DD)
    merchant_owner_stake: "" // decimal as string
  });

  const [error, setError] = useState("");
  const [successMessage, setSuccessMessage] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  function handleChange(e) {
    const { name, value, type, checked } = e.target;

    setForm((prev) => ({
      ...prev,
      [name]: type === "checkbox" ? checked : value
    }));
  }

  function handleSubmit(e) {
    e.preventDefault();

    // Minimal required checks (mirror your model "not null" intent)
    const requiredFields = [
      "merchant_name",

      "merchant_business_profile_dba_name",
      "merchant_business_profile_registration_number",
      "merchant_business_profile_tax_id",
      "merchant_business_profile_website_url",
      "merchant_business_profile_incoporation_date",
      "merchant_business_profile_legal_structure",
      "merchant_business_profile_mcc_code",
      "merchant_business_profile_phone_number",
      "merchant_business_profile_email",

      "merchant_address_line_1",
      "merchant_address_city",
      "merchant_address_state",
      "merchant_address_postal_code",

      "merchant_owner_first_name",
      "merchant_owner_last_name",
      "merchant_owner_phone_number",
      "merchant_owner_email",
      "merchant_owner_dob",
      "merchant_owner_stake"
    ];

    for (const key of requiredFields) {
      if (!String(form[key] ?? "").trim()) {
        setSuccessMessage("");
        setError("Please complete all required fields.");
        return;
      }
    }

    setError("");
    setSuccessMessage("");

    // If you want postal code as a number to match int:
    const payload = {
      ...form,
      merchant_address_postal_code: Number(form.merchant_address_postal_code)
    };

    setIsSubmitting(true);
    fetch("/api/merchant/create", {
      method: "POST",
      headers: {
        "Content-Type": "application/json"
      },
      body: JSON.stringify(payload)
    })
      .then(async (response) => {
        if (!response.ok) {
          const text = await response.text();
          throw new Error(text || "Failed to create merchant");
        }
        return response.json();
      })
      .then((data) => {
        setSuccessMessage(`Merchant created successfully (ID: ${data.merchant_id})`);
      })
      .catch((err) => {
        setError(err.message || "Failed to create merchant");
      })
      .finally(() => {
        setIsSubmitting(false);
      });
  }

  return (
    <div className="auth-page">
      <div className="auth-card">
        <h1 className="auth-title">Merchant KYB</h1>
        <p className="auth-subtitle">Enter your business details to get started with XRPay.</p>

        <form className="auth-form" onSubmit={handleSubmit}>
          <SectionTitle>Merchant</SectionTitle>

          <div className="form-group">
            <label htmlFor="merchant_name">Merchant name</label>
            <input
              id="merchant_name"
              name="merchant_name"
              type="text"
              value={form.merchant_name}
              onChange={handleChange}
              required
            />
          </div>

          <SectionTitle>Business profile</SectionTitle>

          <div className="form-group">
            <label htmlFor="merchant_business_profile_dba_name">DBA name</label>
            <input
              id="merchant_business_profile_dba_name"
              name="merchant_business_profile_dba_name"
              type="text"
              value={form.merchant_business_profile_dba_name}
              onChange={handleChange}
              required
            />
          </div>

          <div className="form-group">
            <label htmlFor="merchant_business_profile_registration_number">
              Registration number
            </label>
            <input
              id="merchant_business_profile_registration_number"
              name="merchant_business_profile_registration_number"
              type="text"
              value={form.merchant_business_profile_registration_number}
              onChange={handleChange}
              required
            />
          </div>

          <div className="form-group">
            <label htmlFor="merchant_business_profile_tax_id">Tax ID</label>
            <input
              id="merchant_business_profile_tax_id"
              name="merchant_business_profile_tax_id"
              type="text"
              value={form.merchant_business_profile_tax_id}
              onChange={handleChange}
              required
              placeholder="EIN / Tax ID"
            />
          </div>

          <div className="form-group">
            <label htmlFor="merchant_business_profile_website_url">Website URL</label>
            <input
              id="merchant_business_profile_website_url"
              name="merchant_business_profile_website_url"
              type="url"
              value={form.merchant_business_profile_website_url}
              onChange={handleChange}
              required
              placeholder="https://example.com"
            />
          </div>

          <div className="form-group">
            <label htmlFor="merchant_business_profile_incoporation_date">
              Incorporation date
            </label>
            <input
              id="merchant_business_profile_incoporation_date"
              name="merchant_business_profile_incoporation_date"
              type="date"
              value={form.merchant_business_profile_incoporation_date}
              onChange={handleChange}
              required
            />
          </div>

          <div className="form-group">
            <label htmlFor="merchant_business_profile_legal_structure">
              Legal structure
            </label>
            <select
              id="merchant_business_profile_legal_structure"
              name="merchant_business_profile_legal_structure"
              value={form.merchant_business_profile_legal_structure}
              onChange={handleChange}
              required
            >
              <option value="llc">LLC</option>
              <option value="corporation">Corporation</option>
              <option value="partnership">Partnership</option>
              <option value="sole_prop">Sole Proprietorship</option>
              <option value="nonprofit">Nonprofit</option>
              <option value="other">Other</option>
            </select>
          </div>

          <div className="form-group">
            <label htmlFor="merchant_business_profile_mcc_code">MCC code</label>
            <input
              id="merchant_business_profile_mcc_code"
              name="merchant_business_profile_mcc_code"
              type="text"
              value={form.merchant_business_profile_mcc_code}
              onChange={handleChange}
              required
              placeholder="e.g., 5812"
            />
          </div>

          <div className="form-group">
            <label htmlFor="merchant_business_profile_phone_number">Phone number</label>
            <input
              id="merchant_business_profile_phone_number"
              name="merchant_business_profile_phone_number"
              type="tel"
              value={form.merchant_business_profile_phone_number}
              onChange={handleChange}
              required
            />
          </div>

          <div className="form-group">
            <label htmlFor="merchant_business_profile_email">Business email</label>
            <input
              id="merchant_business_profile_email"
              name="merchant_business_profile_email"
              type="email"
              value={form.merchant_business_profile_email}
              onChange={handleChange}
              required
            />
          </div>

          <SectionTitle>Business address</SectionTitle>

          <div className="form-group">
            <label htmlFor="merchant_address_line_1">Address line 1</label>
            <input
              id="merchant_address_line_1"
              name="merchant_address_line_1"
              type="text"
              value={form.merchant_address_line_1}
              onChange={handleChange}
              required
            />
          </div>

          <div className="form-group">
            <label htmlFor="merchant_address_line_2">Address line 2</label>
            <input
              id="merchant_address_line_2"
              name="merchant_address_line_2"
              type="text"
              value={form.merchant_address_line_2}
              onChange={handleChange}
              required
            />
          </div>

          <div className="form-group">
            <label htmlFor="merchant_address_city">City</label>
            <input
              id="merchant_address_city"
              name="merchant_address_city"
              type="text"
              value={form.merchant_address_city}
              onChange={handleChange}
              required
            />
          </div>

          <div className="form-group">
            <label htmlFor="merchant_address_state">State/Region</label>
            <input
              id="merchant_address_state"
              name="merchant_address_state"
              type="text"
              value={form.merchant_address_state}
              onChange={handleChange}
              required
            />
          </div>

          <div className="form-group">
            <label htmlFor="merchant_address_postal_code">Postal code</label>
            <input
              id="merchant_address_postal_code"
              name="merchant_address_postal_code"
              type="number"
              value={form.merchant_address_postal_code}
              onChange={handleChange}
              required
            />
          </div>

          <SectionTitle>Owner</SectionTitle>

          <div className="form-group">
            <label htmlFor="merchant_owner_first_name">First name</label>
            <input
              id="merchant_owner_first_name"
              name="merchant_owner_first_name"
              type="text"
              value={form.merchant_owner_first_name}
              onChange={handleChange}
              required
            />
          </div>

          <div className="form-group">
            <label htmlFor="merchant_owner_last_name">Last name</label>
            <input
              id="merchant_owner_last_name"
              name="merchant_owner_last_name"
              type="text"
              value={form.merchant_owner_last_name}
              onChange={handleChange}
              required
            />
          </div>

          <div className="form-group">
            <label htmlFor="merchant_owner_phone_number">Phone number</label>
            <input
              id="merchant_owner_phone_number"
              name="merchant_owner_phone_number"
              type="tel"
              value={form.merchant_owner_phone_number}
              onChange={handleChange}
              required
            />
          </div>

          <div className="form-group">
            <label htmlFor="merchant_owner_email">Owner email</label>
            <input
              id="merchant_owner_email"
              name="merchant_owner_email"
              type="email"
              value={form.merchant_owner_email}
              onChange={handleChange}
              required
            />
          </div>

          <div className="form-group">
            <label htmlFor="merchant_owner_dob">Date of birth</label>
            <input
              id="merchant_owner_dob"
              name="merchant_owner_dob"
              type="date"
              value={form.merchant_owner_dob}
              onChange={handleChange}
              required
            />
          </div>

          <div className="form-group">
            <label htmlFor="merchant_owner_stake">Ownership stake (%)</label>
            <input
              id="merchant_owner_stake"
              name="merchant_owner_stake"
              type="number"
              step="0.0001"
              value={form.merchant_owner_stake}
              onChange={handleChange}
              required
              placeholder="e.g., 25.0000"
            />
          </div>

          {error && <p style={{ color: "red", margin: 0 }}>{error}</p>}
          {successMessage && <p style={{ color: "green", margin: 0 }}>{successMessage}</p>}

          <button className="auth-button" type="submit" disabled={isSubmitting}>
            {isSubmitting ? "Submitting..." : "Create Merchant"}
          </button>

          <p className="auth-footer">
            Back to <Link to="/dashboard">Home</Link>
          </p>
        </form>
      </div>
    </div>
  );
}

function SectionTitle({ children }) {
  return (
    <h2 style={{ margin: "0.5rem 0 0", fontSize: "1rem", opacity: 0.9 }}>
      {children}
    </h2>
  );
}