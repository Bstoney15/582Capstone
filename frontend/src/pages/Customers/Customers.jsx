import { useState, useEffect } from "react";
import { useMerchant } from "../../contexts/MerchantContext";
import "./Customers.css";

export default function Customers() {
    const { merchants, selectedMerchant, isLoading } = useMerchant();

    const [customers, setCustomers] = useState([]);
    const [customersLoading, setCustomersLoading] = useState(false);
    const [customersError, setCustomersError] = useState(null);
    const [search, setSearch] = useState("");

    useEffect(() => {
        if (!selectedMerchant?.id) {
            setCustomers([]);
            setCustomersError(null);
            return;
        }
        const fetchCustomers = async () => {
            setCustomersLoading(true);
            setCustomersError(null);
            try {
                const res = await fetch(
                    `/api/merchant/customers?merchant_id=${encodeURIComponent(selectedMerchant.id)}`
                );
                if (!res.ok) throw new Error("Failed to load customers for this merchant");
                const data = await res.json();
                setCustomers(Array.isArray(data) ? data : []);
            } catch (err) {
                setCustomersError(err.message);
                setCustomers([]);
            } finally {
                setCustomersLoading(false);
            }
        };
        fetchCustomers();
    }, [selectedMerchant?.id]);

    if (isLoading) {
        return <div className="users-loading">Loading...</div>;
    }

    const filtered = customers.filter((c) =>
        [c.first_name, c.last_name, c.email]
            .join(" ")
            .toLowerCase()
            .includes(search.toLowerCase())
    );

    return (
        <div style={{ padding: "2rem", maxWidth: 900, margin: "0 auto" }}>
            <h1 style={{ marginBottom: "1.5rem" }}>Customers</h1>

            {merchants.length === 0 ? (
                <p>You don't have any merchants yet.</p>
            ) : !selectedMerchant ? (
                <p>No merchant selected.</p>
            ) : (
                <section>
                    <div className="users-toolbar">
                        <input
                            type="text"
                            className="customers-search-input"
                            placeholder="Search by name or email…"
                            value={search}
                            onChange={(e) => setSearch(e.target.value)}
                        />
                    </div>
                    <div className="users-info-box">
                        <h2>
                            Customers for {selectedMerchant.name}
                        </h2>
                        {customersLoading && <p>Loading customers...</p>}
                        {customersError && (
                            <p className="error-message">{customersError}</p>
                        )}
                        {!customersLoading && !customersError && (
                            <>
                                {filtered.length === 0 ? (
                                    <p>{customers.length === 0 ? "No customers for this merchant." : "No customers match your search."}</p>
                                ) : (
                                    <table style={{ width: "100%", borderCollapse: "collapse" }}>
                                        <thead>
                                            <tr className="users-table">
                                                <th>First Name</th>
                                                <th>Last Name</th>
                                                <th>Email</th>
                                            </tr>
                                        </thead>
                                        <tbody>
                                            {filtered.map((c) => (
                                                <tr key={c.customer_id}>
                                                    <td>{c.first_name}</td>
                                                    <td>{c.last_name}</td>
                                                    <td>{c.email}</td>
                                                </tr>
                                            ))}
                                        </tbody>
                                    </table>
                                )}
                            </>
                        )}
                    </div>
                </section>
            )}
        </div>
    );
}
