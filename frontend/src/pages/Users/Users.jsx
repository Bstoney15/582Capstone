import { useState, useEffect } from "react";
import { Navigate } from "react-router-dom";
import { useMerchant } from "../../contexts/MerchantContext";
import "./Users.css";

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

    const ROLES = ["Owner", "Admin", "Developer"];

    // Fetch current user info on mount
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

    // Fetch users for the selected merchant whenever it changes
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

    if (!requireRole("Admin")) {
        return <Navigate to="/dashboard" replace />;
    }

    // Placeholder – wire up to /api/merchant/add-user when ready
    const handleAddUser = () => {
        console.log("Add user – endpoint not implemented yet", selectedMerchant?.id);
    };

    // edit stuff is also placeholder, edit endpoint not done yet
    const openEditModal = (user) => {
        setEditingUser(user);
        setEditRole(user.role || "");
        setEditError(null);
    };
    const closeEditModal = () => {
        setEditingUser(null);
        setEditError(null);
    };

    const handleSaveEdit = async () => {
        if (!editingUser || !selectedMerchant?.id) return;
        setEditSaving(true);
        setEditError(null);
        try {
            // TODO: replace with real call when POST /api/merchant/edit-user-role is ready
            const res = await fetch("/api/merchant/edit-user-role", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    merchant_id: selectedMerchant.id,
                    user_id: editingUser.user_id,
                    role: editRole,
                }),
            });
            if (!res.ok) throw new Error("Failed to update role");
            closeEditModal();
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

    const handleRemoveFromMerchant = async () => {
        if (!editingUser || !selectedMerchant?.id) return;
        if (!confirm(`Remove ${editingUser.username} from ${selectedMerchant.name}?`)) return;
        setEditSaving(true);
        setEditError(null);
        try {
            // TODO: add backend endpoint to remove user from merchant, then call it here
            console.log("Remove from merchant – endpoint not implemented yet", {
                merchant_id: selectedMerchant.id,
                user_id: editingUser.user_id,
            });
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

            {/* Your info box */}
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
                    <p>You don’t have any merchants yet.</p>
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
        </div>
    );
}
