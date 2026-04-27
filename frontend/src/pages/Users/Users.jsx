// Users.jsx – page for viewing, adding, editing, and removing users of a merchant organization.
// Author: Ryan Grimsley
// Created: 2026-02-20

import { useState, useEffect } from "react";
import { Navigate } from "react-router-dom";
import { useMerchant } from "../../contexts/MerchantContext";
import "./Users.css";

/**
 * Users renders the merchant user management page. Admins and Owners can view,
 * add, edit roles for, and remove users from the selected merchant.
 * @returns {JSX.Element} The rendered Users page or a redirect for unauthorized callers.
 */
export default function Users() {
    const { merchants, selectedMerchant, setSelectedMerchant, requireRole, isLoading } = useMerchant();

    const [userInfo, setUserInfo] = useState(null);
    const [userInfoLoading, setUserInfoLoading] = useState(true);
    const [userInfoError, setUserInfoError] = useState(null);

    const [merchantUsers, setMerchantUsers] = useState([]);
    const [merchantUsersLoading, setMerchantUsersLoading] = useState(false);
    const [merchantUsersError, setMerchantUsersError] = useState(null);

    const [editingUser, setEditingUser] = useState(null);
    const [editRole, setEditRole] = useState("");
    const [editSaving, setEditSaving] = useState(false);
    const [editError, setEditError] = useState(null);

    const [addingUser, setAddingUser] = useState(false);
    const [addUserUsername, setAddUserUsername] = useState("");
    const [addUserRole, setAddUserRole] = useState("");
    const [addError, setAddError] = useState(null);
    const [addSaving, setAddSaving] = useState(false);

    const ROLES = ["Owner", "Admin", "Developer"];

    // Fetch the current user's own info on mount so the editor ID is available for mutations.
    useEffect(() => {
        const fetchUserInfo = async () => {
            setUserInfoLoading(true);
            setUserInfoError(null);
            try {
                const res = await fetch("/api/user/info");
                if (!res.ok) throw new Error("Failed to load your info");
                const data = await res.json();
                setUserInfo(data);
            } catch (err) {
                setUserInfoError(err.message);
            } finally {
                setUserInfoLoading(false);
            }
        };
        fetchUserInfo();
    }, []);

    // Fetch users for the selected merchant whenever it changes.
    useEffect(() => {
        if (!selectedMerchant?.id) {
            setMerchantUsers([]);
            setMerchantUsersError(null);
            return;
        }

        const fetchMerchantUsers = async () => {
            setMerchantUsersLoading(true);
            setMerchantUsersError(null);
            try {
                const res = await fetch(
                    `/api/merchant/get-merchant-users?merchant_id=${encodeURIComponent(selectedMerchant.id)}`
                );
                if (!res.ok) throw new Error("Failed to load users for this merchant");
                const data = await res.json();
                setMerchantUsers(Array.isArray(data) ? data : []);
            } catch (err) {
                setMerchantUsersError(err.message);
                setMerchantUsers([]);
            } finally {
                setMerchantUsersLoading(false);
            }
        };
        fetchMerchantUsers();
    }, [selectedMerchant?.id]);

    if (isLoading) {
        return (
            <div className="users-loading">
                Loading...
            </div>
        );
    }

    // Redirect non-admin users back to the dashboard.
    if (!requireRole("Admin")) {
        return <Navigate to="/dashboard" replace />;
    }

    /**
     * Opens the add-user modal if a merchant is selected.
     */
    const handleAddUser = () => {
        if (!selectedMerchant?.id) return;
        openAddModal();
    };

    const openAddModal = () => {
        setAddingUser(true);
        setAddUserUsername("");
        setAddUserRole(ROLES[0] || "");
        setAddError(null);
    };

    const closeAddModal = () => {
        setAddingUser(false);
        setAddUserUsername("");
        setAddUserRole("");
        setAddError(null);
    };

    /**
     * Submits the add-user form, creating a new role entry for the specified username.
     */
    const handleSaveAddUser = async () => {
        if (!selectedMerchant?.id) return;
        const trimmedUsername = addUserUsername.trim();
        if (!trimmedUsername || !addUserRole) {
            setAddError("Username and role are required.");
            return;
        }
        setAddSaving(true);
        setAddError(null);
        try {
            const res = await fetch("/api/merchant/add-user", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    merchant_id: selectedMerchant.id,
                    user_username: trimmedUsername,
                    role: addUserRole,
                }),
            });
            if (!res.ok) {
                const text = await res.text();
                throw new Error(text || "Failed to add user");
            }

            // Refresh the merchant users list after a successful add.
            try {
                const usersRes = await fetch(
                    `/api/merchant/get-merchant-users?merchant_id=${encodeURIComponent(selectedMerchant.id)}`
                );
                if (usersRes.ok) {
                    const usersData = await usersRes.json();
                    setMerchantUsers(Array.isArray(usersData) ? usersData : []);
                }
            } catch {
                // ignore refresh errors; user was still created
            }

            closeAddModal();
        } catch (err) {
            setAddError(err.message || "Failed to add user");
        } finally {
            setAddSaving(false);
        }
    };

    /**
     * Opens the edit modal for the given user.
     * @param {Object} user The user record to edit.
     */
    const openEditModal = (user) => {
        setEditingUser(user);
        setEditRole(user.role || "");
        setEditError(null);
    };

    const closeEditModal = () => {
        setEditingUser(null);
        setEditError(null);
    };

    /**
     * Saves the role change for the user being edited.
     */
    const handleSaveEdit = async () => {
        if (!editingUser || !selectedMerchant?.id) return;
        setEditSaving(true);
        setEditError(null);
        try {
            const res = await fetch("/api/merchant/edit-user-role", {
                method: "PATCH",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    merchant_id: selectedMerchant.id,
                    user_id: editingUser.user_id,
                    role: editRole,
                    editor_id: userInfo.user_id,
                }),
            });
            if (!res.ok) throw new Error("Failed to update role");
            closeEditModal();
            // Optimistically update the local list with the new role.
            setMerchantUsers((prev) =>
                prev.map((u) =>
                    u.user_id === editingUser.user_id ? { ...u, role: editRole } : u
                )
            );
        } catch (err) {
            setEditError(err.message);
        } finally {
            setEditSaving(false);
        }
    };

    /**
     * Removes the user being edited from the merchant after confirming with the user.
     */
    const handleRemoveFromMerchant = async () => {
        if (!editingUser || !selectedMerchant?.id) return;
        if (!confirm(`Remove ${editingUser.username} from ${selectedMerchant.name}?`)) return;
        setEditSaving(true);
        setEditError(null);
        try {
            const res = await fetch("/api/merchant/remove-user", {
                method: "DELETE",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    merchant_id: selectedMerchant.id,
                    user_id: editingUser.user_id,
                    editor_id: userInfo.user_id,
                })
            })
            if (!res.ok) throw new Error("Failed to remove user");
            closeEditModal();
            setMerchantUsers((prev) => prev.filter((u) => u.user_id !== editingUser.user_id));
        } catch (err) {
            setEditError(err.message);
        } finally {
            setEditSaving(false);
        }
    };

    return (
        <div style={{ padding: "2rem", maxWidth: 900, margin: "0 auto" }}>
            <h1 style={{ marginBottom: "1.5rem" }}>Users</h1>

            {/* Current user info summary */}
            <section className="users-info-box"
            >
                <h2 >
                    Your info
                </h2>
                {userInfoLoading && <p>Loading...</p>}
                {userInfoError && (
                    <p className="error-message">{userInfoError}</p>
                )}
                {!userInfoLoading && !userInfoError && userInfo && (
                    <dl className="users-info-dl">
                        <dt >Username</dt>
                        <dd >{userInfo.user_username}</dd>
                        <dt >First name</dt>
                        <dd >{userInfo.user_first_name}</dd>
                        <dt >Last name</dt>
                        <dd >{userInfo.user_last_name}</dd>
                        <dt >Status</dt>
                        <dd>{userInfo.user_status}</dd>
                    </dl>
                )}
            </section>

            {/* Merchant users for currently selected merchant */}
            <section>
                {merchants.length === 0 ? (
                    <p>You don't have any merchants yet.</p>
                ) : !selectedMerchant ? (
                    <p>No merchant selected.</p>
                ) : (
                    <>
                        <div
                            className="users-toolbar"
                        >
                            <button
                                type="button"
                                className="users-btn"
                                onClick={handleAddUser}
                            >
                                Add user
                            </button>
                        </div>
                        <div className="users-info-box">
                            <h2>
                                Merchant users for {selectedMerchant.name}
                                {selectedMerchant.role && ` (${selectedMerchant.role})`}
                            </h2>
                            {merchantUsersLoading && <p>Loading users...</p>}
                            {merchantUsersError && (
                                <p className="error-message">{merchantUsersError}</p>
                            )}
                            {!merchantUsersLoading && !merchantUsersError && (
                                <>
                                    {merchantUsers.length === 0 ? (
                                        <p>No users for this merchant.</p>
                                    ) : (
                                        <table style={{ width: "100%", borderCollapse: "collapse" }}>
                                            <thead>
                                                <tr className="users-table">
                                                    <th>Username</th>
                                                    <th>First name</th>
                                                    <th>Last name</th>
                                                    <th>Status</th>
                                                    <th>Role</th>
                                                    <th />
                                                </tr>
                                            </thead>
                                            <tbody>
                                                {merchantUsers.map((u) => (
                                                    <tr key={u.user_id}>
                                                        <td>{u.username}</td>
                                                        <td>{u.first_name}</td>
                                                        <td>{u.last_name}</td>
                                                        <td>{u.user_status}</td>
                                                        <td>{u.role}</td>
                                                        <td>
                                                            <button
                                                                type="button"
                                                                className="users-btn-edit"
                                                                onClick={() => openEditModal(u)}
                                                            >
                                                                Edit
                                                            </button>
                                                        </td>
                                                    </tr>
                                                ))}
                                            </tbody>
                                        </table>
                                    )}
                                </>
                            )}
                        </div>
                    </>
                )}
            </section>

            {/* Edit user modal */}
            {editingUser && (
                <div className="users-modal-backdrop" onClick={closeEditModal}>
                    <div className="users-modal" onClick={(e) => e.stopPropagation()}>
                        <h3>Edit user</h3>
                        <p className="users-modal-subtitle">
                            {editingUser.username}
                            {editingUser.first_name || editingUser.last_name
                                ? ` (${[editingUser.first_name, editingUser.last_name].filter(Boolean).join(" ")})`
                                : ""}
                        </p>
                        <div className="form-group">
                            <label>Role</label>
                            <select value={editRole} onChange={(e) => setEditRole(e.target.value)}>
                                {ROLES.map((r) => (
                                    <option key={r} value={r}>
                                        {r}
                                    </option>
                                ))}
                            </select>
                        </div>
                        {editError && <p className="users-modal-error">{editError}</p>}
                        <div className="users-modal-actions">
                            <button
                                type="button"
                                className="users-btn-save"
                                onClick={handleSaveEdit}
                                disabled={editSaving}
                            >
                                {editSaving ? "Saving…" : "Save"}
                            </button>
                            <button
                                type="button"
                                className="users-btn-remove"
                                onClick={handleRemoveFromMerchant}
                                disabled={editSaving}
                            >
                                Remove from merchant
                            </button>
                            <button type="button" className="users-btn-cancel" onClick={closeEditModal}>
                                Cancel
                            </button>
                        </div>
                    </div>
                </div>
            )}

            {/* Add user modal */}
            {addingUser && (
                <div className="users-modal-backdrop" onClick={closeAddModal}>
                    <div className="users-modal" onClick={(e) => e.stopPropagation()}>
                        <h3>Add user</h3>
                        <div className="form-group">
                            <label>Username</label>
                            <input
                                type="text"
                                value={addUserUsername}
                                onChange={(e) => setAddUserUsername(e.target.value)}
                                placeholder="Enter username"
                            />
                        </div>
                        <div className="form-group">
                            <label>Role</label>
                            <select
                                value={addUserRole}
                                onChange={(e) => setAddUserRole(e.target.value)}
                            >
                                <option value="">Select a role</option>
                                {ROLES.map((r) => (
                                    <option key={r} value={r}>
                                        {r}
                                    </option>
                                ))}
                            </select>
                        </div>
                        {addError && <p className="users-modal-error">{addError}</p>}
                        <div className="users-modal-actions">
                            <button
                                type="button"
                                className="users-btn-save"
                                onClick={handleSaveAddUser}
                                disabled={addSaving || !addUserUsername.trim() || !addUserRole}
                            >
                                {addSaving ? "Adding…" : "Add user"}
                            </button>
                            <button
                                type="button"
                                className="users-btn-cancel"
                                onClick={closeAddModal}
                                disabled={addSaving}
                            >
                                Cancel
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}
