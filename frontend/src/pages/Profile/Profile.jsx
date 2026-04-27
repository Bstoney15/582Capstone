import { useEffect, useState } from "react";
import { fetchUserProfile, updateUserProfile } from "../../services/profileService";
import "./Profile.css";

const EMPTY_FORM = {
    user_username: "",
    user_first_name: "",
    user_last_name: "",
    current_password: "",
};

function buildPayload(form, originalProfile) {
    const payload = {};

    if (form.user_username.trim() !== originalProfile.user_username) {
        payload.user_username = form.user_username.trim();
    }

    if (form.user_first_name.trim() !== originalProfile.user_first_name) {
        payload.user_first_name = form.user_first_name.trim();
    }

    if (form.user_last_name.trim() !== originalProfile.user_last_name) {
        payload.user_last_name = form.user_last_name.trim();
    }

    return payload;
}

export default function Profile() {
    const [profile, setProfile] = useState(null);
    const [form, setForm] = useState(EMPTY_FORM);
    const [isLoading, setIsLoading] = useState(true);
    const [isSaving, setIsSaving] = useState(false);
    const [error, setError] = useState("");
    const [successMessage, setSuccessMessage] = useState("");

    useEffect(() => {
        let isMounted = true;

        async function loadProfile() {
            setIsLoading(true);
            setError("");

            try {
                const data = await fetchUserProfile();
                if (!isMounted) return;
                setProfile(data);
                setForm({
                    user_username: data.user_username ?? "",
                    user_first_name: data.user_first_name ?? "",
                    user_last_name: data.user_last_name ?? "",
                    current_password: "",
                });
            } catch (err) {
                if (!isMounted) return;
                setError(err.message || "Failed to load profile.");
            } finally {
                if (isMounted) {
                    setIsLoading(false);
                }
            }
        }

        loadProfile();

        return () => {
            isMounted = false;
        };
    }, []);

    const hasChanges =
        profile &&
        (form.user_username.trim() !== profile.user_username ||
            form.user_first_name.trim() !== profile.user_first_name ||
            form.user_last_name.trim() !== profile.user_last_name);

    const handleChange = (event) => {
        const { name, value } = event.target;
        setForm((currentForm) => ({
            ...currentForm,
            [name]: value,
        }));
        setSuccessMessage("");
    };

    const handleReset = () => {
        if (!profile) return;

        setForm({
            user_username: profile.user_username,
            user_first_name: profile.user_first_name,
            user_last_name: profile.user_last_name,
            current_password: "",
        });
        setError("");
        setSuccessMessage("");
    };

    const handleSubmit = async (event) => {
        event.preventDefault();
        if (!profile) return;

        const payload = buildPayload(form, profile);
        if (Object.keys(payload).length === 0) {
            setSuccessMessage("No changes to save.");
            setError("");
            return;
        }

        if (!form.current_password) {
            setError("Current password is required to save profile changes.");
            setSuccessMessage("");
            return;
        }

        payload.current_password = form.current_password;

        setIsSaving(true);
        setError("");
        setSuccessMessage("");

        try {
            const updatedProfile = await updateUserProfile(payload);
            setProfile(updatedProfile);
            setForm({
                user_username: updatedProfile.user_username,
                user_first_name: updatedProfile.user_first_name,
                user_last_name: updatedProfile.user_last_name,
                current_password: "",
            });
            setSuccessMessage("Profile updated successfully.");
        } catch (err) {
            setError(err.message || "Failed to save profile.");
        } finally {
            setIsSaving(false);
        }
    };

    if (isLoading) {
        return <div className="profile-page"><p>Loading profile...</p></div>;
    }

    return (
        <div className="profile-page">
            <div className="profile-shell">
                <section className="profile-hero">
                    <p className="profile-eyebrow">Account</p>
                    <h1>Your profile</h1>
                    <p className="profile-subtitle">
                        Update the details tied to your XRPay account. Changes here affect
                        how your name and username appear across the app.
                    </p>
                </section>

                {error && (
                    <div className="profile-banner profile-banner-error" role="alert">
                        {error}
                    </div>
                )}

                {successMessage && (
                    <div className="profile-banner profile-banner-success" role="status">
                        {successMessage}
                    </div>
                )}

                <div className="profile-grid">
                    <section className="profile-card profile-summary-card">
                        <p className="profile-card-label">Profile summary</p>
                        <div className="profile-summary-value">
                            {profile?.user_first_name} {profile?.user_last_name}
                        </div>
                        <div className="profile-summary-username">@{profile?.user_username}</div>

                        <dl className="profile-meta-list">
                            <div>
                                <dt>User ID</dt>
                                <dd>{profile?.user_id}</dd>
                            </div>
                            <div>
                                <dt>Status</dt>
                                <dd>
                                    <span className="profile-status-pill">{profile?.user_status}</span>
                                </dd>
                            </div>
                        </dl>
                    </section>

                    <section className="profile-card">
                        <div className="profile-form-header">
                            <div>
                                <p className="profile-card-label">Editable fields</p>
                                <h2>Profile details</h2>
                            </div>
                        </div>

                        <form className="profile-form" onSubmit={handleSubmit}>
                            <label className="profile-field">
                                <span>Username</span>
                                <input
                                    type="text"
                                    name="user_username"
                                    value={form.user_username}
                                    onChange={handleChange}
                                    autoComplete="username"
                                    disabled={isSaving}
                                />
                            </label>

                            <div className="profile-field-row">
                                <label className="profile-field">
                                    <span>First name</span>
                                    <input
                                        type="text"
                                        name="user_first_name"
                                        value={form.user_first_name}
                                        onChange={handleChange}
                                        autoComplete="given-name"
                                        disabled={isSaving}
                                    />
                                </label>

                                <label className="profile-field">
                                    <span>Last name</span>
                                    <input
                                        type="text"
                                        name="user_last_name"
                                        value={form.user_last_name}
                                        onChange={handleChange}
                                        autoComplete="family-name"
                                        disabled={isSaving}
                                    />
                                </label>
                            </div>

                            <label className="profile-field">
                                <span>Current password</span>
                                <input
                                    type="password"
                                    name="current_password"
                                    value={form.current_password}
                                    onChange={handleChange}
                                    autoComplete="current-password"
                                    disabled={isSaving}
                                />
                            </label>

                            <p className="profile-helper-text">
                                Re-enter your current password before saving changes to your username or name.
                            </p>

                            <div className="profile-actions">
                                <button
                                    type="button"
                                    className="profile-button profile-button-secondary"
                                    onClick={handleReset}
                                    disabled={isSaving || !hasChanges}
                                >
                                    Reset
                                </button>
                                <button
                                    type="submit"
                                    className="profile-button profile-button-primary"
                                    disabled={isSaving || !hasChanges}
                                >
                                    {isSaving ? "Saving..." : "Save changes"}
                                </button>
                            </div>
                        </form>
                    </section>
                </div>
            </div>
        </div>
    );
}
