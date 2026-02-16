import { useState } from "react";
import { Link } from "react-router-dom";
import { useTheme } from "../hooks/useTheme";

export default function Signup() {
  const [formData, setFormData] = useState({
    firstName: "",
    lastName: "",
    username: "",
    password: "",
    confirmPassword: "",
  });

  const [errors, setErrors] = useState({});

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: value,
    }));
    
    // Clear error when user types
    if (errors[name]) {
      setErrors((prev) => ({
        ...prev,
        [name]: "",
      }));
    }
  };

  const validate = () => {
    let tempErrors = {};
    if (!formData.firstName.trim()) tempErrors.firstName = "First name is required";
    if (!formData.lastName.trim()) tempErrors.lastName = "Last name is required";
    if (!formData.username.trim()) tempErrors.username = "Username is required";
    
    if (!formData.password) {
      tempErrors.password = "Password is required";
    } else if (formData.password.length < 8) {
      tempErrors.password = "Password must be at least 8 characters";
    }

    if (formData.confirmPassword !== formData.password) {
      tempErrors.confirmPassword = "Passwords do not match";
    }

    setErrors(tempErrors);
    return Object.keys(tempErrors).length === 0;
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    if (validate()) {
      console.log("Form submitted successfully", formData);
      // Proceed with submission logic 


    }
  };

  return (
    <div className="auth-page">
      <button className="theme-toggle" onClick={toggleTheme}>
        {theme === "dark" ? "☀️" : "🌙"}
      </button>

      <div className="auth-card">
        <h1 className="auth-title">Create Account</h1>
        <p className="auth-subtitle">Get started with XRPay</p>

        <form className="auth-form" onSubmit={handleSubmit}>
          <div className="form-group">
            <label>First Name</label>
            <input 
              type="text" 
              name="firstName"
              value={formData.firstName}
              onChange={handleChange}
              className={errors.firstName ? "input-error" : ""}
            />
            {errors.firstName && <span className="error-message">{errors.firstName}</span>}
          </div>

          <div className="form-group">
            <label>Last Name</label>
            <input 
              type="text" 
              name="lastName"
              value={formData.lastName}
              onChange={handleChange}
              className={errors.lastName ? "input-error" : ""}
            />
            {errors.lastName && <span className="error-message">{errors.lastName}</span>}
          </div>

          <div className="form-group">
            <label>Username</label>
            <input 
              type="text"
              name="username"
              value={formData.username}
              onChange={handleChange}
              className={errors.username ? "input-error" : ""}
            />
             {errors.username && <span className="error-message">{errors.username}</span>}
          </div>

          <div className="form-group">
            <label>Password</label>
            <input 
              type="password" 
              name="password"
              value={formData.password}
              onChange={handleChange}
              className={errors.password ? "input-error" : ""}
            />
             {errors.password && <span className="error-message">{errors.password}</span>}
          </div>

          <div className="form-group">
            <label>Confirm Password</label>
            <input 
              type="password"
              name="confirmPassword"
              value={formData.confirmPassword}
              onChange={handleChange}
              className={errors.confirmPassword ? "input-error" : ""}
            />
             {errors.confirmPassword && <span className="error-message">{errors.confirmPassword}</span>}
          </div>

          <button className="auth-button" type="submit">
            Create Account
          </button>

          <p className="auth-footer">
            Already have an account? <Link to="/login">Log in</Link>
          </p>
        </form>
      </div>
    </div>
  );
}
