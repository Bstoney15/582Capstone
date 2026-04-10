import { useState, useEffect } from "react";
import { useMerchant } from "../../contexts/MerchantContext";
import "./Customers.css";

export default function Customers() {
    const { selectedMerchant, isLoading } = useMerchant();

    const [customers, setCustomers] = useState([]);
    const [customersLoading, setCustomersLoading] = useState(false);
    const [customersError, setCustomersError] = useState(null);
    const [search, setSearch] = useState("");

    useEffect(() => {
        if (!selectedMerchant?.id) {
            setCustomers([]);
            setCustomersError(null);
            setSearch("");
            return;
        }
        const fetchCustomers = async () => {
            setCustomersLoading(true);
            setCustomersError(null);
            setSearch("");
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
        return <div className="customers-page">Loading...</div>;
    }

    const filtered = customers.filter((c) =>
        [c.first_name, c.last_name, c.email]
            .join(" ")
            .toLowerCase()
            .includes(search.toLowerCase())
    );

    return (
        <div className="customers-page">
            <h1>Customers</h1>

            {selectedMerchant ? (
                <p className="customers-description">
                    Customers associated with {selectedMerchant.name}.
                </p>
            ) : (
                <p className="customers-description">
                    Select a merchant account to view customers.
                </p>
            )}

            {customersError && (
                <p className="customers-error">{customersError}</p>
            )}

            {selectedMerchant && (
                <input
                    type="text"
                    value={search}
                    onChange={(e) => setSearch(e.target.value)}
                    placeholder="Search by name or email…"
                    className="customers-search-input"
                />
            )}

            {selectedMerchant && customersLoading && <p>Loading customers…</p>}

            {selectedMerchant && !customersLoading && filtered.length === 0 && (
                <p className="customers-empty">
                    {customers.length === 0
                        ? "No customers for this merchant."
                        : "No customers match your search."}
                </p>
            )}

            {selectedMerchant && !customersLoading && filtered.length > 0 && (
                <div className="customers-table-container">
                    <table className="customers-table">
                        <thead>
                            <tr>
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
                </div>
            )}
        </div>
    );
}
