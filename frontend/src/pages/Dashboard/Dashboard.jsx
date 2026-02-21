import { useMerchant } from "../../contexts/MerchantContext";

export default function Dashboard() {
    const { selectedMerchant, isLoading } = useMerchant();

    if (isLoading) {
        return <div style={{ padding: "2rem" }}>Loading dashboard...</div>;
    }

    return (
        <div style={{ padding: "2rem" }}>

            <h1>Dashboard</h1>
            <p>Welcome to your dummy dashboard! This content is protected and only visible to authenticated users.</p>

            <div style={{ marginTop: "1rem", padding: "1rem", background: "rgba(128,128,128,0.1)", borderRadius: "8px" }}>
                <p style={{ margin: "0.25rem 0" }}><strong>Current Merchant:</strong> {selectedMerchant.name}</p>
                <p style={{ margin: "0.25rem 0" }}><strong>Merchant ID:</strong> {selectedMerchant.id}</p>
            </div>
        </div>
    );
}
