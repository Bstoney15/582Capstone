import React, { useState } from 'react';
import { useContext } from 'react';
import './ApiDocs.css';

const ApiDocs = () => {
  
  const [activeTab, setActiveTab] = useState('api');
  const [expandedSections, setExpandedSections] = useState({});

  const toggleSection = (sectionId) => {
    setExpandedSections(prev => ({
      ...prev,
      [sectionId]: !prev[sectionId]
    }));
  };

  const CodeBlock = ({ code, language = 'bash' }) => (
    <pre className={`code-block language-${language}`}>
      <code>{code}</code>
    </pre>
  );

  const EndpointCard = ({ method, path, title, description, auth, params, response, examples, expanded, onToggle }) => {
    const methodColors = {
      GET: '#61affe',
      POST: '#49cc90',
      PATCH: '#fca130',
      DELETE: '#f93e3e',
    };

    return (
      <div className="endpoint-card">
        <div className="endpoint-header" onClick={onToggle}>
          <div className="endpoint-method-path">
            <span
              className="endpoint-method"
              style={{ backgroundColor: methodColors[method] }}
            >
              {method}
            </span>
            <span className="endpoint-path">{path}</span>
          </div>
          <div className="endpoint-toggle">
            <span className="endpoint-title">{title}</span>
            <span className="toggle-icon">{expanded ? '▼' : '▶'}</span>
          </div>
        </div>

        {expanded && (
          <div className="endpoint-details">
            <p className="endpoint-description">{description}</p>

            {auth && (
              <div className="endpoint-section">
                <h5>Authentication</h5>
                <p>{auth}</p>
              </div>
            )}

            {params && (
              <div className="endpoint-section">
                <h5>Parameters</h5>
                <table className="params-table">
                  <thead>
                    <tr>
                      <th>Name</th>
                      <th>Type</th>
                      <th>Description</th>
                    </tr>
                  </thead>
                  <tbody>
                    {params.map((param, idx) => (
                      <tr key={idx}>
                        <td className="param-name">{param.name}</td>
                        <td className="param-type">{param.type}</td>
                        <td>{param.description}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}

            {response && (
              <div className="endpoint-section">
                <h5>Response</h5>
                <CodeBlock code={response} language="json" />
              </div>
            )}

            {examples && (
              <div className="endpoint-section">
                <h5>Examples</h5>
                {examples.map((example, idx) => (
                  <div key={idx} className="example-block">
                    <h6>{example.title}</h6>
                    <CodeBlock code={example.code} language={example.language || 'bash'} />
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
      </div>
    );
  };

  return (
    <div className={`api-docs`}>
      <div className="api-docs-container">
        {/* Header */}
        <div className="api-docs-header">
          <h1>API Reference</h1>
          <p>Complete reference for XRPay API and webhook specifications</p>
        </div>

        {/* Tab Navigation */}
        <div className="api-tabs">
          <button
            className={`tab-button ${activeTab === 'api' ? 'active' : ''}`}
            onClick={() => setActiveTab('api')}
          >
            API Reference
          </button>
          <button
            className={`tab-button ${activeTab === 'webhooks' ? 'active' : ''}`}
            onClick={() => setActiveTab('webhooks')}
          >
            Webhooks
          </button>
        </div>

        {/* API Reference Tab */}
        {activeTab === 'api' && (
          <div className="tab-content">
            {/* Table of Contents */}
            <div className="table-of-contents">
              <h3>Table of Contents</h3>
              <ul>
                <li><a href="#auth-overview">Authentication Overview</a></li>
                <li><a href="#api-key-auth">API Key Authentication</a></li>
                <li><a href="#auth-endpoints">Authentication Endpoints</a></li>
                <li><a href="#user-endpoints">User Endpoints</a></li>
                <li><a href="#customer-endpoints">Customer Endpoints</a></li>
                <li><a href="#invoice-endpoints">Invoice Endpoints</a></li>
                <li><a href="#merchant-endpoints">Merchant Endpoints</a></li>
                <li><a href="#admin-endpoints">Admin Endpoints</a></li>
              </ul>
            </div>

            {/* Authentication Overview */}
            <section id="auth-overview" className="doc-section">
              <h2>Authentication Overview</h2>
              <p>XRPay API supports multiple authentication methods depending on the endpoint and use case:</p>
              <ul>
                <li><strong>Session/Cookie Authentication:</strong> For dashboard and web interface requests. Credentials are automatically included in cookies.</li>
                <li><strong>API Key Authentication:</strong> For server-to-server integrations. Required for merchant-scoped operations like customer and invoice management.</li>
                <li><strong>No Authentication:</strong> Public endpoints like invoice checkout, payment verification, and health checks.</li>
              </ul>
            </section>

            {/* API Key Authentication */}
            <section id="api-key-auth" className="doc-section">
              <h2>API Key Authentication</h2>
              <p>API keys are used to authenticate requests from your backend to XRPay's API. All API key requests are scoped to a specific merchant.</p>
              
              <h4>Getting Your API Key</h4>
              <ol>
                <li>Log in to your XRPay dashboard</li>
                <li>Navigate to <strong>API Keys</strong></li>
                <li>Click <strong>Create New API Key</strong> and give it a name</li>
                <li>Copy the generated key (displayed only once)</li>
              </ol>

              <h4>Using Your API Key</h4>
              <p>Include your API key in the request body as <code>api_key</code> field for requests that require merchant API key authentication:</p>
              <CodeBlock
                code={`curl -X POST https://api.xrpay.com/api/merchant/customers \\
  -H "Content-Type: application/json" \\
  -d '{
    "api_key": "sk_1234567890abcdef",
    "first_name": "John",
    "last_name": "Doe",
    "email": "john@example.com"
  }'`}
              />

              <h4>API Key Format</h4>
              <p>API keys are prefixed with <code>sk_</code> followed by 48 hexadecimal characters. Example: <code>sk_1234567890abcdef1234567890abcdef12345678</code></p>
            </section>

            {/* Authentication Endpoints */}
            <section id="auth-endpoints" className="doc-section">
              <h2>Authentication Endpoints</h2>

              <EndpointCard
                method="POST"
                path="/api/user/login"
                title="Login"
                description="Authenticate a user and create a session"
                auth="None (Public)"
                params={[
                  { name: 'email', type: 'string', description: 'User email address' },
                  { name: 'password', type: 'string', description: 'User password' }
                ]}
                response={`{
  "success": true,
  "user_id": "uuid-123",
  "message": "Login successful"
}`}
                examples={[
                  {
                    title: 'cURL',
                    code: `curl -X POST https://api.xrpay.com/api/user/login \\
  -H "Content-Type: application/json" \\
  -d '{
    "email": "user@example.com",
    "password": "secure_password"
  }'`
                  },
                  {
                    title: 'JavaScript',
                    language: 'javascript',
                    code: `const response = await fetch('https://api.xrpay.com/api/user/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    email: 'user@example.com',
    password: 'secure_password'
  }),
  credentials: 'include'
});
const data = await response.json();`
                  }
                ]}
                expanded={expandedSections['login']}
                onToggle={() => toggleSection('login')}
              />

              <EndpointCard
                method="POST"
                path="/api/user/signup"
                title="Signup"
                description="Create a new user account"
                auth="None (Public)"
                params={[
                  { name: 'email', type: 'string', description: 'Unique email address' },
                  { name: 'password', type: 'string', description: 'Account password' },
                  { name: 'first_name', type: 'string', description: 'User first name' },
                  { name: 'last_name', type: 'string', description: 'User last name' }
                ]}
                response={`{
  "success": true,
  "user_id": "uuid-123",
  "message": "Signup successful"
}`}
                examples={[
                  {
                    title: 'cURL',
                    code: `curl -X POST https://api.xrpay.com/api/user/signup \\
  -H "Content-Type: application/json" \\
  -d '{
    "email": "newuser@example.com",
    "password": "secure_password",
    "first_name": "Jane",
    "last_name": "Smith"
  }'`
                  }
                ]}
                expanded={expandedSections['signup']}
                onToggle={() => toggleSection('signup')}
              />

              <EndpointCard
                method="GET"
                path="/api/user/auth"
                title="Check Authentication Status"
                description="Verify if the current session is valid and get the authenticated user's ID"
                auth="Session Cookie"
                response={`{
  "authenticated": true,
  "userId": "uuid-123"
}`}
                examples={[
                  {
                    title: 'cURL',
                    code: `curl -X GET https://api.xrpay.com/api/user/auth \\
  -H "Cookie: session_token=..."`
                  }
                ]}
                expanded={expandedSections['check-auth']}
                onToggle={() => toggleSection('check-auth')}
              />

              <EndpointCard
                method="POST"
                path="/api/user/logout"
                title="Logout"
                description="End the current session"
                auth="Session Cookie"
                response={`{
  "success": true,
  "message": "Logout successful"
}`}
                examples={[
                  {
                    title: 'cURL',
                    code: `curl -X POST https://api.xrpay.com/api/user/logout \\
  -H "Cookie: session_token=..."`
                  }
                ]}
                expanded={expandedSections['logout']}
                onToggle={() => toggleSection('logout')}
              />
            </section>

            {/* User Endpoints */}
            <section id="user-endpoints" className="doc-section">
              <h2>User Endpoints</h2>

              <EndpointCard
                method="GET"
                path="/api/user/info"
                title="Get User Information"
                description="Retrieve the authenticated user's profile information"
                auth="Session Cookie"
                response={`{
  "user_id": "uuid-123",
  "email": "user@example.com",
  "first_name": "John",
  "last_name": "Doe"
}`}
                examples={[
                  {
                    title: 'cURL',
                    code: `curl -X GET https://api.xrpay.com/api/user/info \\
  -H "Cookie: session_token=..."`
                  }
                ]}
                expanded={expandedSections['user-info']}
                onToggle={() => toggleSection('user-info')}
              />

              <EndpointCard
                method="GET"
                path="/api/user/merchants"
                title="Get User's Merchants"
                description="List all merchants the authenticated user is a member of"
                auth="Session Cookie"
                response={`[
  {
    "merchant_id": "uuid-456",
    "merchant_name": "My Store",
    "role": "Admin",
    "business_name": "My Business"
  }
]`}
                examples={[
                  {
                    title: 'cURL',
                    code: `curl -X GET https://api.xrpay.com/api/user/merchants \\
  -H "Cookie: session_token=..."`
                  }
                ]}
                expanded={expandedSections['user-merchants']}
                onToggle={() => toggleSection('user-merchants')}
              />

              <EndpointCard
                method="DELETE"
                path="/api/user/merchants/{merchant_id}"
                title="Leave a Merchant"
                description="Remove yourself from a merchant (except if you're the only owner)"
                auth="Session Cookie"
                params={[
                  { name: 'merchant_id', type: 'string (path)', description: 'ID of the merchant to leave' }
                ]}
                response={`{
  "success": true,
  "message": "Successfully left merchant"
}`}
                examples={[
                  {
                    title: 'cURL',
                    code: `curl -X DELETE https://api.xrpay.com/api/user/merchants/uuid-456 \\
  -H "Cookie: session_token=..."`
                  }
                ]}
                expanded={expandedSections['leave-merchant']}
                onToggle={() => toggleSection('leave-merchant')}
              />
            </section>

            {/* Customer Endpoints */}
            <section id="customer-endpoints" className="doc-section">
              <h2>Customer Endpoints</h2>
              <p><strong>Note:</strong> Customer endpoints require API Key authentication and are scoped to the merchant who owns the API key.</p>

              <EndpointCard
                method="POST"
                path="/api/merchant/customers"
                title="Create Customer"
                description="Create a new customer for the authenticated merchant"
                auth="API Key"
                params={[
                  { name: 'api_key', type: 'string', description: 'Your merchant API key' },
                  { name: 'first_name', type: 'string', description: 'Customer first name' },
                  { name: 'last_name', type: 'string', description: 'Customer last name' },
                  { name: 'email', type: 'string', description: 'Customer email address' }
                ]}
                response={`{
  "customer_id": "uuid-789",
  "first_name": "John",
  "last_name": "Doe",
  "email": "john@example.com",
  "created_at": "2026-04-25T10:30:00Z"
}`}
                examples={[
                  {
                    title: 'cURL',
                    code: `curl -X POST https://api.xrpay.com/api/merchant/customers \\
  -H "Content-Type: application/json" \\
  -d '{
    "api_key": "sk_1234567890abcdef",
    "first_name": "John",
    "last_name": "Doe",
    "email": "john@example.com"
  }'`
                  },
                  {
                    title: 'JavaScript',
                    language: 'javascript',
                    code: `const response = await fetch('https://api.xrpay.com/api/merchant/customers', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    api_key: 'sk_1234567890abcdef',
    first_name: 'John',
    last_name: 'Doe',
    email: 'john@example.com'
  })
});
const customer = await response.json();`
                  }
                ]}
                expanded={expandedSections['create-customer']}
                onToggle={() => toggleSection('create-customer')}
              />

              <EndpointCard
                method="GET"
                path="/api/merchant/customers"
                title="List Customers"
                description="Retrieve all customers for the authenticated merchant"
                auth="API Key"
                params={[
                  { name: 'api_key', type: 'string (query)', description: 'Your merchant API key' },
                  { name: 'search', type: 'string (query)', description: 'Optional search term for first name, last name, or email' }
                ]}
                response={`[
  {
    "customer_id": "uuid-789",
    "first_name": "John",
    "last_name": "Doe",
    "email": "john@example.com",
    "created_at": "2026-04-25T10:30:00Z"
  }
]`}
                examples={[
                  {
                    title: 'cURL',
                    code: `curl -X GET "https://api.xrpay.com/api/merchant/customers?api_key=sk_1234567890abcdef"`
                  },
                  {
                    title: 'With Search',
                    code: `curl -X GET "https://api.xrpay.com/api/merchant/customers?api_key=sk_1234567890abcdef&search=John"`
                  }
                ]}
                expanded={expandedSections['list-customers']}
                onToggle={() => toggleSection('list-customers')}
              />

              <EndpointCard
                method="GET"
                path="/api/merchant/customers/{customer_id}"
                title="Get Customer"
                description="Retrieve details for a specific customer"
                auth="API Key"
                params={[
                  { name: 'api_key', type: 'string (query)', description: 'Your merchant API key' },
                  { name: 'customer_id', type: 'string (path)', description: 'ID of the customer' }
                ]}
                response={`{
  "customer_id": "uuid-789",
  "first_name": "John",
  "last_name": "Doe",
  "email": "john@example.com",
  "created_at": "2026-04-25T10:30:00Z"
}`}
                examples={[
                  {
                    title: 'cURL',
                    code: `curl -X GET "https://api.xrpay.com/api/merchant/customers/uuid-789?api_key=sk_1234567890abcdef"`
                  }
                ]}
                expanded={expandedSections['get-customer']}
                onToggle={() => toggleSection('get-customer')}
              />

              <EndpointCard
                method="PATCH"
                path="/api/merchant/customers/{customer_id}"
                title="Update Customer"
                description="Update customer information"
                auth="API Key"
                params={[
                  { name: 'api_key', type: 'string', description: 'Your merchant API key' },
                  { name: 'customer_id', type: 'string (path)', description: 'ID of the customer' },
                  { name: 'first_name', type: 'string', description: 'Updated first name (optional)' },
                  { name: 'last_name', type: 'string', description: 'Updated last name (optional)' },
                  { name: 'email', type: 'string', description: 'Updated email (optional)' }
                ]}
                response={`{
  "customer_id": "uuid-789",
  "first_name": "Jane",
  "last_name": "Smith",
  "email": "jane@example.com",
  "updated_at": "2026-04-25T11:00:00Z"
}`}
                examples={[
                  {
                    title: 'cURL',
                    code: `curl -X PATCH https://api.xrpay.com/api/merchant/customers/uuid-789 \\
  -H "Content-Type: application/json" \\
  -d '{
    "api_key": "sk_1234567890abcdef",
    "first_name": "Jane",
    "email": "jane@example.com"
  }'`
                  }
                ]}
                expanded={expandedSections['update-customer']}
                onToggle={() => toggleSection('update-customer')}
              />

              <EndpointCard
                method="DELETE"
                path="/api/merchant/customers/{customer_id}"
                title="Delete Customer"
                description="Delete a customer and all associated data"
                auth="API Key"
                params={[
                  { name: 'api_key', type: 'string', description: 'Your merchant API key' },
                  { name: 'customer_id', type: 'string (path)', description: 'ID of the customer to delete' }
                ]}
                response={`{
  "success": true,
  "message": "Customer deleted successfully"
}`}
                examples={[
                  {
                    title: 'cURL',
                    code: `curl -X DELETE https://api.xrpay.com/api/merchant/customers/uuid-789 \\
  -H "Content-Type: application/json" \\
  -d '{"api_key": "sk_1234567890abcdef"}'`
                  }
                ]}
                expanded={expandedSections['delete-customer']}
                onToggle={() => toggleSection('delete-customer')}
              />
            </section>

            {/* Invoice Endpoints */}
            <section id="invoice-endpoints" className="doc-section">
              <h2>Invoice Endpoints</h2>

              <EndpointCard
                method="POST"
                path="/api/invoices"
                title="Create Invoice (Public API)"
                description="Create a new invoice that can be paid via XRP (widget-compatible)"
                auth="API Key"
                params={[
                  { name: 'api_key', type: 'string', description: 'Your merchant API key' },
                  { name: 'amount_xrp', type: 'decimal', description: 'Amount in XRP (required if amount_usd not provided)' },
                  { name: 'amount_usd', type: 'decimal', description: 'Amount in USD (required if amount_xrp not provided)' }
                ]}
                response={`{
  "invoice_id": "uuid-inv-123",
  "amount_xrp": "100.00",
  "amount_usd": "200.00",
  "usd_per_xrp": "2.00",
  "pricing_source": "coinbase"
}`}
                examples={[
                  {
                    title: 'Create with XRP Amount',
                    code: `curl -X POST https://api.xrpay.com/api/invoices \\
  -H "Content-Type: application/json" \\
  -d '{
    "api_key": "sk_1234567890abcdef",
    "amount_xrp": "100"
  }'`
                  },
                  {
                    title: 'Create with USD Amount',
                    code: `curl -X POST https://api.xrpay.com/api/invoices \\
  -H "Content-Type: application/json" \\
  -d '{
    "api_key": "sk_1234567890abcdef",
    "amount_usd": "200"
  }'`
                  }
                ]}
                expanded={expandedSections['create-invoice-public']}
                onToggle={() => toggleSection('create-invoice-public')}
              />

              <EndpointCard
                method="GET"
                path="/api/invoices/{uuid}"
                title="Get Invoice (Public)"
                description="Retrieve invoice details by ID (for checkout page display)"
                auth="None (Public)"
                params={[
                  { name: 'uuid', type: 'string (path)', description: 'Invoice ID' }
                ]}
                response={`{
  "invoice_id": "uuid-inv-123",
  "amount_xrp": "100.00",
  "amount_usd": "200.00",
  "status": "pending",
  "created_at": "2026-04-25T10:30:00Z"
}`}
                examples={[
                  {
                    title: 'cURL',
                    code: `curl -X GET https://api.xrpay.com/api/invoices/uuid-inv-123`
                  }
                ]}
                expanded={expandedSections['get-invoice-public']}
                onToggle={() => toggleSection('get-invoice-public')}
              />

              <EndpointCard
                method="GET"
                path="/api/invoices/{uuid}/events"
                title="Stream Invoice Events (Public)"
                description="Server-sent events stream for real-time invoice status updates"
                auth="None (Public)"
                params={[
                  { name: 'uuid', type: 'string (path)', description: 'Invoice ID' }
                ]}
                response={`event: invoice.status_changed
data: {"invoice_id":"uuid-inv-123","status":"paid","amount_received":"100.00"}

event: invoice.confirmed
data: {"invoice_id":"uuid-inv-123","confirmation_count":5}`}
                examples={[
                  {
                    title: 'JavaScript',
                    language: 'javascript',
                    code: `const eventSource = new EventSource(
  'https://api.xrpay.com/api/invoices/uuid-inv-123/events'
);

eventSource.addEventListener('invoice.status_changed', (e) => {
  const data = JSON.parse(e.data);
  console.log('Invoice status:', data.status);
});

eventSource.addEventListener('invoice.confirmed', (e) => {
  const data = JSON.parse(e.data);
  console.log('Invoice confirmed');
  eventSource.close();
});`
                  }
                ]}
                expanded={expandedSections['stream-invoice-events']}
                onToggle={() => toggleSection('stream-invoice-events')}
              />

              <EndpointCard
                method="POST"
                path="/api/verify"
                title="Verify Invoice Payment"
                description="Verify that an invoice has been paid (public endpoint for widget integration)"
                auth="None (Public)"
                params={[
                  { name: 'invoice_id', type: 'string', description: 'Invoice ID' }
                ]}
                response={`{
  "verified": true,
  "invoice_id": "uuid-inv-123",
  "amount_paid": "100.00",
  "status": "confirmed"
}`}
                examples={[
                  {
                    title: 'cURL',
                    code: `curl -X POST https://api.xrpay.com/api/verify \\
  -H "Content-Type: application/json" \\
  -d '{"invoice_id": "uuid-inv-123"}'`
                  }
                ]}
                expanded={expandedSections['verify-payment']}
                onToggle={() => toggleSection('verify-payment')}
              />

              <EndpointCard
                method="POST"
                path="/api/merchant/invoices"
                title="Create Invoice (Merchant API)"
                description="Create an invoice associated with an optional customer"
                auth="API Key"
                params={[
                  { name: 'api_key', type: 'string', description: 'Your merchant API key' },
                  { name: 'customer_id', type: 'string', description: 'Optional customer ID to associate with invoice' },
                  { name: 'amount_xrp', type: 'decimal', description: 'Amount in XRP (required if amount_usd not provided)' },
                  { name: 'amount_usd', type: 'decimal', description: 'Amount in USD (required if amount_xrp not provided)' }
                ]}
                response={`{
  "invoice_id": "uuid-inv-456",
  "customer_id": "uuid-cust-789",
  "amount_xrp": "50.00",
  "amount_usd": "100.00",
  "status": "pending",
  "created_at": "2026-04-25T10:30:00Z"
}`}
                examples={[
                  {
                    title: 'cURL',
                    code: `curl -X POST https://api.xrpay.com/api/merchant/invoices \\
  -H "Content-Type: application/json" \\
  -d '{
    "api_key": "sk_1234567890abcdef",
    "customer_id": "uuid-cust-789",
    "amount_xrp": "50"
  }'`
                  }
                ]}
                expanded={expandedSections['create-invoice-merchant']}
                onToggle={() => toggleSection('create-invoice-merchant')}
              />

              <EndpointCard
                method="GET"
                path="/api/merchant/invoices"
                title="List Invoices (Merchant API)"
                description="Retrieve all invoices for the authenticated merchant"
                auth="API Key"
                params={[
                  { name: 'api_key', type: 'string (query)', description: 'Your merchant API key' },
                  { name: 'customer_id', type: 'string (query)', description: 'Optional filter by customer ID' }
                ]}
                response={`[
  {
    "invoice_id": "uuid-inv-456",
    "customer_id": "uuid-cust-789",
    "amount_xrp": "50.00",
    "amount_usd": "100.00",
    "status": "pending",
    "created_at": "2026-04-25T10:30:00Z"
  }
]`}
                examples={[
                  {
                    title: 'List All',
                    code: `curl -X GET "https://api.xrpay.com/api/merchant/invoices?api_key=sk_1234567890abcdef"`
                  },
                  {
                    title: 'Filter by Customer',
                    code: `curl -X GET "https://api.xrpay.com/api/merchant/invoices?api_key=sk_1234567890abcdef&customer_id=uuid-cust-789"`
                  }
                ]}
                expanded={expandedSections['list-invoices-merchant']}
                onToggle={() => toggleSection('list-invoices-merchant')}
              />

              <EndpointCard
                method="GET"
                path="/api/merchant/invoices/{invoice_id}"
                title="Get Invoice (Merchant API)"
                description="Retrieve details for a specific invoice"
                auth="API Key"
                params={[
                  { name: 'api_key', type: 'string (query)', description: 'Your merchant API key' },
                  { name: 'invoice_id', type: 'string (path)', description: 'Invoice ID' }
                ]}
                response={`{
  "invoice_id": "uuid-inv-456",
  "customer_id": "uuid-cust-789",
  "amount_xrp": "50.00",
  "amount_usd": "100.00",
  "status": "confirmed",
  "created_at": "2026-04-25T10:30:00Z",
  "paid_at": "2026-04-25T11:00:00Z"
}`}
                examples={[
                  {
                    title: 'cURL',
                    code: `curl -X GET "https://api.xrpay.com/api/merchant/invoices/uuid-inv-456?api_key=sk_1234567890abcdef"`
                  }
                ]}
                expanded={expandedSections['get-invoice-merchant']}
                onToggle={() => toggleSection('get-invoice-merchant')}
              />

              <EndpointCard
                method="DELETE"
                path="/api/merchant/invoices/{invoice_id}"
                title="Delete Invoice (Merchant API)"
                description="Delete an unpaid invoice"
                auth="API Key"
                params={[
                  { name: 'api_key', type: 'string', description: 'Your merchant API key' },
                  { name: 'invoice_id', type: 'string (path)', description: 'Invoice ID to delete' }
                ]}
                response={`{
  "success": true,
  "message": "Invoice deleted successfully"
}`}
                examples={[
                  {
                    title: 'cURL',
                    code: `curl -X DELETE https://api.xrpay.com/api/merchant/invoices/uuid-inv-456 \\
  -H "Content-Type: application/json" \\
  -d '{"api_key": "sk_1234567890abcdef"}'`
                  }
                ]}
                expanded={expandedSections['delete-invoice']}
                onToggle={() => toggleSection('delete-invoice')}
              />

              <EndpointCard
                method="GET"
                path="/api/customer/{customer_id}/invoices"
                title="Get Customer Invoices"
                description="Retrieve all invoices for a specific customer (requires Developer role or above)"
                auth="Session Cookie (Developer+)"
                params={[
                  { name: 'customer_id', type: 'string (path)', description: 'Customer ID' },
                  { name: 'merchant_id', type: 'string (query)', description: 'Merchant ID (required to filter by merchant)' }
                ]}
                response={`[
  {
    "invoice_id": "uuid-inv-456",
    "customer_id": "uuid-cust-789",
    "amount_xrp": "50.00",
    "amount_usd": "100.00",
    "status": "confirmed",
    "created_at": "2026-04-25T10:30:00Z"
  }
]`}
                examples={[
                  {
                    title: 'cURL',
                    code: `curl -X GET "https://api.xrpay.com/api/customer/uuid-cust-789/invoices?merchant_id=uuid-merchant-123" \\
  -H "Cookie: session_token=..."`
                  }
                ]}
                expanded={expandedSections['get-customer-invoices']}
                onToggle={() => toggleSection('get-customer-invoices')}
              />
            </section>

            {/* Merchant Endpoints */}
            <section id="merchant-endpoints" className="doc-section">
              <h2>Merchant Endpoints</h2>
              <p><strong>Note:</strong> Merchant API Key endpoints allow merchants to manage their own API keys. Most other merchant operations require Admin role.</p>

              <EndpointCard
                method="GET"
                path="/api/merchant/api_key"
                title="List API Keys"
                description="Retrieve all API keys for the authenticated merchant"
                auth="Session Cookie (Merchant+)"
                response={`[
  {
    "id": "sk-123456",
    "name": "Production API Key",
    "generated_at": "2026-04-20T10:30:00Z"
  }
]`}
                examples={[
                  {
                    title: 'cURL',
                    code: `curl -X GET https://api.xrpay.com/api/merchant/api_key \\
  -H "Cookie: session_token=..."`
                  }
                ]}
                expanded={expandedSections['list-api-keys']}
                onToggle={() => toggleSection('list-api-keys')}
              />

              <EndpointCard
                method="POST"
                path="/api/merchant/api_key"
                title="Create API Key"
                description="Generate a new API key for the authenticated merchant"
                auth="Session Cookie (Merchant+)"
                params={[
                  { name: 'merchant_id', type: 'string', description: 'Merchant ID' },
                  { name: 'name', type: 'string', description: 'Friendly name for the API key' }
                ]}
                response={`{
  "id": "sk-123456",
  "name": "Production API Key",
  "generated_at": "2026-04-25T10:30:00Z",
  "api_key": "sk_1234567890abcdef1234567890abcdef12345678"
}`}
                examples={[
                  {
                    title: 'cURL',
                    code: `curl -X POST https://api.xrpay.com/api/merchant/api_key \\
  -H "Content-Type: application/json" \\
  -H "Cookie: session_token=..." \\
  -d '{
    "merchant_id": "uuid-merchant-123",
    "name": "Production API Key"
  }'`
                  }
                ]}
                expanded={expandedSections['create-api-key']}
                onToggle={() => toggleSection('create-api-key')}
              />

              <EndpointCard
                method="DELETE"
                path="/api/merchant/api_key/{api_key}"
                title="Delete API Key"
                description="Revoke an API key"
                auth="Session Cookie (Merchant+)"
                params={[
                  { name: 'api_key', type: 'string (path)', description: 'API key ID to delete' }
                ]}
                response={`{
  "success": true,
  "message": "API key deleted successfully"
}`}
                examples={[
                  {
                    title: 'cURL',
                    code: `curl -X DELETE https://api.xrpay.com/api/merchant/api_key/sk-123456 \\
  -H "Cookie: session_token=..."`
                  }
                ]}
                expanded={expandedSections['delete-api-key']}
                onToggle={() => toggleSection('delete-api-key')}
              />

              <EndpointCard
                method="GET"
                path="/api/dashboard"
                title="Get Dashboard Data"
                description="Retrieve dashboard analytics for the authenticated merchant (Developer role or above)"
                auth="Session Cookie (Developer+)"
                response={`{
  "merchant_id": "uuid-merchant-123",
  "total_invoices": 150,
  "total_customers": 45,
  "total_revenue_xrp": "1000.50",
  "total_revenue_usd": "2001.00",
  "pending_invoices": 5,
  "confirmed_invoices": 145
}`}
                examples={[
                  {
                    title: 'cURL',
                    code: `curl -X GET https://api.xrpay.com/api/dashboard \\
  -H "Cookie: session_token=..."`
                  }
                ]}
                expanded={expandedSections['get-dashboard']}
                onToggle={() => toggleSection('get-dashboard')}
              />

              <EndpointCard
                method="GET"
                path="/api/dashboard/search"
                title="Search Invoices"
                description="Search invoices by ID or customer information (Developer role or above)"
                auth="Session Cookie (Developer+)"
                params={[
                  { name: 'query', type: 'string (query)', description: 'Search query (invoice ID, customer name, or email)' }
                ]}
                response={`[
  {
    "invoice_id": "uuid-inv-456",
    "customer_id": "uuid-cust-789",
    "customer_name": "John Doe",
    "amount_xrp": "50.00",
    "status": "confirmed",
    "created_at": "2026-04-25T10:30:00Z"
  }
]`}
                examples={[
                  {
                    title: 'cURL',
                    code: `curl -X GET "https://api.xrpay.com/api/dashboard/search?query=John" \\
  -H "Cookie: session_token=..."`
                  }
                ]}
                expanded={expandedSections['search-invoices']}
                onToggle={() => toggleSection('search-invoices')}
              />
            </section>

            {/* Admin Endpoints */}
            <section id="admin-endpoints" className="doc-section">
              <h2>Admin Endpoints</h2>
              <p><strong>Note:</strong> These endpoints require Admin role and are used for merchant management and configuration.</p>

              <EndpointCard
                method="POST"
                path="/api/merchant/create"
                title="Create Merchant"
                description="Create a new merchant (Admin only)"
                auth="Session Cookie (Admin)"
                params={[
                  { name: 'business_name', type: 'string', description: 'Business legal name' },
                  { name: 'merchant_name', type: 'string', description: 'Display name for the merchant' },
                  { name: 'email', type: 'string', description: 'Business email' }
                ]}
                response={`{
  "merchant_id": "uuid-merchant-123",
  "business_name": "My Business Inc",
  "merchant_name": "My Store",
  "email": "business@example.com",
  "created_at": "2026-04-25T10:30:00Z"
}`}
                examples={[
                  {
                    title: 'cURL',
                    code: `curl -X POST https://api.xrpay.com/api/merchant/create \\
  -H "Content-Type: application/json" \\
  -H "Cookie: session_token=..." \\
  -d '{
    "business_name": "My Business Inc",
    "merchant_name": "My Store",
    "email": "business@example.com"
  }'`
                  }
                ]}
                expanded={expandedSections['create-merchant']}
                onToggle={() => toggleSection('create-merchant')}
              />

              <EndpointCard
                method="POST"
                path="/api/merchant/add-user"
                title="Add User to Merchant"
                description="Invite a user to join a merchant (Admin only)"
                auth="Session Cookie (Admin)"
                params={[
                  { name: 'merchant_id', type: 'string', description: 'Target merchant ID' },
                  { name: 'user_id', type: 'string', description: 'User ID to invite' },
                  { name: 'role', type: 'string', description: 'Role: Owner, Admin, or Developer' }
                ]}
                response={`{
  "success": true,
  "message": "User added successfully"
}`}
                examples={[
                  {
                    title: 'cURL',
                    code: `curl -X POST https://api.xrpay.com/api/merchant/add-user \\
  -H "Content-Type: application/json" \\
  -H "Cookie: session_token=..." \\
  -d '{
    "merchant_id": "uuid-merchant-123",
    "user_id": "uuid-user-456",
    "role": "Developer"
  }'`
                  }
                ]}
                expanded={expandedSections['add-user']}
                onToggle={() => toggleSection('add-user')}
              />

              <EndpointCard
                method="PATCH"
                path="/api/merchant/edit-user-role"
                title="Edit User Role"
                description="Change a user's role within a merchant (Admin only)"
                auth="Session Cookie (Admin)"
                params={[
                  { name: 'merchant_id', type: 'string', description: 'Merchant ID' },
                  { name: 'user_id', type: 'string', description: 'User ID to update' },
                  { name: 'role', type: 'string', description: 'New role: Owner, Admin, or Developer' }
                ]}
                response={`{
  "success": true,
  "message": "User role updated"
}`}
                examples={[
                  {
                    title: 'cURL',
                    code: `curl -X PATCH https://api.xrpay.com/api/merchant/edit-user-role \\
  -H "Content-Type: application/json" \\
  -H "Cookie: session_token=..." \\
  -d '{
    "merchant_id": "uuid-merchant-123",
    "user_id": "uuid-user-456",
    "role": "Admin"
  }'`
                  }
                ]}
                expanded={expandedSections['edit-user-role']}
                onToggle={() => toggleSection('edit-user-role')}
              />

              <EndpointCard
                method="GET"
                path="/api/merchant/get-merchant-users"
                title="List Merchant Users"
                description="Get all users and their roles for a merchant (Admin only)"
                auth="Session Cookie (Admin)"
                params={[
                  { name: 'merchant_id', type: 'string (query)', description: 'Merchant ID' }
                ]}
                response={`[
  {
    "user_id": "uuid-user-123",
    "email": "admin@example.com",
    "first_name": "Jane",
    "last_name": "Smith",
    "role": "Admin"
  }
]`}
                examples={[
                  {
                    title: 'cURL',
                    code: `curl -X GET "https://api.xrpay.com/api/merchant/get-merchant-users?merchant_id=uuid-merchant-123" \\
  -H "Cookie: session_token=..."`
                  }
                ]}
                expanded={expandedSections['list-merchant-users']}
                onToggle={() => toggleSection('list-merchant-users')}
              />

              <EndpointCard
                method="DELETE"
                path="/api/merchant/remove-user"
                title="Remove User from Merchant"
                description="Remove a user from a merchant (Admin only)"
                auth="Session Cookie (Admin)"
                params={[
                  { name: 'merchant_id', type: 'string', description: 'Merchant ID' },
                  { name: 'user_id', type: 'string', description: 'User ID to remove' }
                ]}
                response={`{
  "success": true,
  "message": "User removed successfully"
}`}
                examples={[
                  {
                    title: 'cURL',
                    code: `curl -X DELETE https://api.xrpay.com/api/merchant/remove-user \\
  -H "Content-Type: application/json" \\
  -H "Cookie: session_token=..." \\
  -d '{
    "merchant_id": "uuid-merchant-123",
    "user_id": "uuid-user-456"
  }'`
                  }
                ]}
                expanded={expandedSections['remove-user']}
                onToggle={() => toggleSection('remove-user')}
              />

              <EndpointCard
                method="GET"
                path="/api/merchant/get-wallet"
                title="Get Merchant Wallet"
                description="Retrieve the merchant's XRP wallet address (Admin only)"
                auth="Session Cookie (Admin)"
                params={[
                  { name: 'merchant_id', type: 'string (query)', description: 'Merchant ID' }
                ]}
                response={`{
  "wallet_address": "rN7n7otQDd6FczFgLdhmKkpNvQrV3H9Yc1",
  "merchant_id": "uuid-merchant-123"
}`}
                examples={[
                  {
                    title: 'cURL',
                    code: `curl -X GET "https://api.xrpay.com/api/merchant/get-wallet?merchant_id=uuid-merchant-123" \\
  -H "Cookie: session_token=..."`
                  }
                ]}
                expanded={expandedSections['get-wallet']}
                onToggle={() => toggleSection('get-wallet')}
              />

              <EndpointCard
                method="PATCH"
                path="/api/merchant/set-wallet"
                title="Set Merchant Wallet"
                description="Update the merchant's XRP wallet address (Admin only)"
                auth="Session Cookie (Admin)"
                params={[
                  { name: 'merchant_id', type: 'string', description: 'Merchant ID' },
                  { name: 'wallet_address', type: 'string', description: 'XRP wallet address' }
                ]}
                response={`{
  "success": true,
  "wallet_address": "rN7n7otQDd6FczFgLdhmKkpNvQrV3H9Yc1"
}`}
                examples={[
                  {
                    title: 'cURL',
                    code: `curl -X PATCH https://api.xrpay.com/api/merchant/set-wallet \\
  -H "Content-Type: application/json" \\
  -H "Cookie: session_token=..." \\
  -d '{
    "merchant_id": "uuid-merchant-123",
    "wallet_address": "rN7n7otQDd6FczFgLdhmKkpNvQrV3H9Yc1"
  }'`
                  }
                ]}
                expanded={expandedSections['set-wallet']}
                onToggle={() => toggleSection('set-wallet')}
              />
            </section>

            {/* Health Check */}
            <section id="health-check" className="doc-section">
              <h2>Health Check</h2>

              <EndpointCard
                method="GET"
                path="/api/health"
                title="Health Check"
                description="Simple endpoint to check if the API is online"
                auth="None (Public)"
                response={`{
  "status": "healthy",
  "timestamp": "2026-04-25T10:30:00Z"
}`}
                examples={[
                  {
                    title: 'cURL',
                    code: `curl -X GET https://api.xrpay.com/api/health`
                  }
                ]}
                expanded={expandedSections['health-check']}
                onToggle={() => toggleSection('health-check')}
              />
            </section>
          </div>
        )}

        {/* Webhooks Tab */}
        {activeTab === 'webhooks' && (
          <div className="tab-content">
            {/* Webhook TOC */}
            <div className="table-of-contents">
              <h3>Table of Contents</h3>
              <ul>
                <li><a href="#webhook-overview">Overview</a></li>
                <li><a href="#webhook-events">Event Types</a></li>
                <li><a href="#webhook-payload">Event Payload</a></li>
                <li><a href="#webhook-headers">Headers</a></li>
                <li><a href="#webhook-signatures">Signature Verification</a></li>
                <li><a href="#webhook-retry">Retry Logic</a></li>
                <li><a href="#webhook-examples">Examples</a></li>
              </ul>
            </div>

            <section id="webhook-overview" className="doc-section">
              <h2>Webhook Overview</h2>
              <p>
                XRPay sends real-time event notifications to your registered webhook endpoints when events occur 
                in your merchant account (e.g., invoices are paid, customers are created). Each webhook event is 
                cryptographically signed using HMAC-SHA256 to ensure authenticity.
              </p>
              <p>
                To receive webhooks, you'll need to:
              </p>
              <ol>
                <li>Register a webhook endpoint URL (via the Webhooks section in the dashboard)</li>
                <li>Store your webhook signing secret key securely</li>
                <li>Validate incoming webhook signatures</li>
                <li>Process the event data</li>
                <li>Return a 2xx response to acknowledge receipt</li>
              </ol>
            </section>

            <section id="webhook-events" className="doc-section">
              <h2>Event Types</h2>
              <p>The following events are currently supported:</p>
              <div className="event-list">
                <div className="event-type">
                  <h4>invoice.created</h4>
                  <p>Fired when a new invoice is created via the API or dashboard</p>
                </div>
                <div className="event-type">
                  <h4>invoice.payment_received</h4>
                  <p>Fired when payment is received for an invoice (initial detection)</p>
                </div>
                <div className="event-type">
                  <h4>invoice.confirmed</h4>
                  <p>Fired when an invoice payment is confirmed on the XRP ledger</p>
                </div>
                <div className="event-type">
                  <h4>invoice.expired</h4>
                  <p>Fired when an invoice expires without payment</p>
                </div>
                <div className="event-type">
                  <h4>customer.created</h4>
                  <p>Fired when a new customer is created</p>
                </div>
                <div className="event-type">
                  <h4>customer.updated</h4>
                  <p>Fired when a customer's information is updated</p>
                </div>
                <div className="event-type">
                  <h4>customer.deleted</h4>
                  <p>Fired when a customer is deleted</p>
                </div>
                <div className="event-type">
                  <h4>webhook.test</h4>
                  <p>Test event sent to verify webhook configuration (manual trigger from dashboard)</p>
                </div>
              </div>
            </section>

            <section id="webhook-payload" className="doc-section">
              <h2>Event Payload Structure</h2>
              <p>All webhook events are sent as JSON with the following structure:</p>
              <CodeBlock
                code={`{
  "event_id": "1234567890123456789",
  "event_type": "invoice.confirmed",
  "created_at": "2026-04-25T10:30:00Z",
  "data": {
    "invoice_id": "uuid-inv-123",
    "customer_id": "uuid-cust-456",
    "amount_xrp": "100.00",
    "amount_usd": "200.00",
    "status": "confirmed",
    "paid_at": "2026-04-25T10:28:00Z",
    "confirmations": 5
  }
}`}
                language="json"
              />
              <table className="params-table">
                <thead>
                  <tr>
                    <th>Field</th>
                    <th>Type</th>
                    <th>Description</th>
                  </tr>
                </thead>
                <tbody>
                  <tr>
                    <td className="param-name">event_id</td>
                    <td className="param-type">string</td>
                    <td>Unique identifier for this event (nanosecond timestamp)</td>
                  </tr>
                  <tr>
                    <td className="param-name">event_type</td>
                    <td className="param-type">string</td>
                    <td>Type of event (e.g., "invoice.confirmed")</td>
                  </tr>
                  <tr>
                    <td className="param-name">created_at</td>
                    <td className="param-type">ISO8601 string</td>
                    <td>Timestamp when the event occurred (UTC)</td>
                  </tr>
                  <tr>
                    <td className="param-name">data</td>
                    <td className="param-type">object</td>
                    <td>Event-specific data (see event types below)</td>
                  </tr>
                </tbody>
              </table>
            </section>

            <section id="webhook-headers" className="doc-section">
              <h2>Webhook Request Headers</h2>
              <p>Each webhook request includes the following headers:</p>
              <table className="params-table">
                <thead>
                  <tr>
                    <th>Header</th>
                    <th>Value</th>
                  </tr>
                </thead>
                <tbody>
                  <tr>
                    <td className="param-name">Content-Type</td>
                    <td>application/json</td>
                  </tr>
                  <tr>
                    <td className="param-name">Accept</td>
                    <td>application/json</td>
                  </tr>
                  <tr>
                    <td className="param-name">User-Agent</td>
                    <td>XRPay-Webhook-Dispatcher/1.0</td>
                  </tr>
                  <tr>
                    <td className="param-name">X-Webhook-Event</td>
                    <td>Event type (e.g., "invoice.confirmed")</td>
                  </tr>
                  <tr>
                    <td className="param-name">X-Webhook-Timestamp</td>
                    <td>Unix timestamp when the webhook was sent</td>
                  </tr>
                  <tr>
                    <td className="param-name">X-Webhook-Signature</td>
                    <td>HMAC-SHA256 signature for verification</td>
                  </tr>
                </tbody>
              </table>
            </section>

            <section id="webhook-signatures" className="doc-section">
              <h2>Webhook Signature Verification</h2>
              <p>
                Each webhook request is signed to prove it came from XRPay. You should always verify the signature 
                before processing the webhook data.
              </p>

              <h4>Signature Format</h4>
              <p>The signature is in the <code>X-Webhook-Signature</code> header:</p>
              <CodeBlock code={`t=1618937325,v1=abcdef1234567890...`} />
              <p>
                Where <code>t</code> is the Unix timestamp and <code>v1</code> is the HMAC-SHA256 hash.
              </p>

              <h4>Verification Process</h4>
              <ol>
                <li>Extract the timestamp and signature from the <code>X-Webhook-Signature</code> header</li>
                <li>Construct the signed content as: <code>&lt;timestamp&gt;.&lt;request_body&gt;</code></li>
                <li>Compute HMAC-SHA256 using your webhook secret key</li>
                <li>Compare the computed hash with the signature from the header</li>
              </ol>

              <h4>Node.js Example</h4>
              <CodeBlock
                language="javascript"
                code={`const crypto = require('crypto');
const express = require('express');
const app = express();

app.use(express.raw({ type: 'application/json' }));

// Webhook signing key from your XRPay dashboard
const WEBHOOK_SECRET = 'your_webhook_secret_key';

app.post('/webhooks/xrpay', (req, res) => {
  const signature = req.headers['x-webhook-signature'];
  const timestamp = req.headers['x-webhook-timestamp'];
  const eventType = req.headers['x-webhook-event'];
  
  if (!signature || !timestamp) {
    return res.status(400).json({ error: 'Missing headers' });
  }

  // Reconstruct the signed content
  const body = req.body.toString();
  const signedContent = \`\${timestamp}.\${body}\`;

  // Compute the expected signature
  const expectedSignature = crypto
    .createHmac('sha256', WEBHOOK_SECRET)
    .update(signedContent)
    .digest('hex');

  // Extract the v1 signature from the header
  const headerParts = signature.split(',');
  const v1Signature = headerParts
    .find(part => part.startsWith('v1='))
    ?.substring(3);

  // Compare signatures (constant-time comparison)
  if (!v1Signature || !crypto.timingSafeEqual(
    Buffer.from(v1Signature),
    Buffer.from(expectedSignature)
  )) {
    return res.status(401).json({ error: 'Invalid signature' });
  }

  // Signature is valid, process the webhook
  const payload = JSON.parse(body);
  console.log(\`Received event: \${payload.event_type}\`);
  console.log('Event data:', payload.data);

  // Handle the event
  if (payload.event_type === 'invoice.confirmed') {
    // Process invoice confirmation
  } else if (payload.event_type === 'customer.created') {
    // Process customer creation
  }

  // Acknowledge receipt
  res.status(200).json({ success: true });
});

app.listen(3000, () => {
  console.log('Webhook server listening on port 3000');
});`}
              />

              <h4>Python Example</h4>
              <CodeBlock
                language="python"
                code={`import hmac
import hashlib
from flask import Flask, request

app = Flask(__name__)

# Webhook signing key from your XRPay dashboard
WEBHOOK_SECRET = 'your_webhook_secret_key'

def verify_webhook_signature(timestamp, body, signature_header):
    """Verify the webhook signature."""
    # Construct the signed content
    signed_content = f"{timestamp}.{body}".encode()
    
    # Compute the expected signature
    expected_signature = hmac.new(
        WEBHOOK_SECRET.encode(),
        signed_content,
        hashlib.sha256
    ).hexdigest()
    
    # Extract the v1 signature from the header
    # Format: "t=1618937325,v1=abcd1234..."
    header_parts = signature_header.split(',')
    v1_signature = None
    for part in header_parts:
        if part.startswith('v1='):
            v1_signature = part[3:]
            break
    
    # Compare signatures (constant-time comparison)
    if not v1_signature:
        return False
    
    return hmac.compare_digest(v1_signature, expected_signature)

@app.route('/webhooks/xrpay', methods=['POST'])
def handle_webhook():
    # Extract headers
    signature = request.headers.get('X-Webhook-Signature')
    timestamp = request.headers.get('X-Webhook-Timestamp')
    event_type = request.headers.get('X-Webhook-Event')
    
    if not signature or not timestamp:
        return {'error': 'Missing headers'}, 400
    
    # Get raw body for signature verification
    body = request.get_data(as_text=True)
    
    # Verify signature
    if not verify_webhook_signature(timestamp, body, signature):
        return {'error': 'Invalid signature'}, 401
    
    # Signature is valid, process the webhook
    payload = request.json
    print(f"Received event: {payload['event_type']}")
    print("Event data:", payload['data'])
    
    # Handle the event
    if payload['event_type'] == 'invoice.confirmed':
        # Process invoice confirmation
        pass
    elif payload['event_type'] == 'customer.created':
        # Process customer creation
        pass
    
    # Acknowledge receipt
    return {'success': True}, 200

if __name__ == '__main__':
    app.run(port=3000)`}
              />
            </section>

            <section id="webhook-retry" className="doc-section">
              <h2>Retry Logic</h2>
              <p>
                XRPay will retry failed webhook deliveries using exponential backoff:
              </p>
              <ul>
                <li><strong>Maximum Attempts:</strong> 3 total attempts</li>
                <li><strong>Backoff Strategy:</strong> Exponential (attempt * 500ms)</li>
                <li><strong>Attempt 1:</strong> Immediate</li>
                <li><strong>Attempt 2:</strong> After 500ms</li>
                <li><strong>Attempt 3:</strong> After 1000ms</li>
                <li><strong>Timeout:</strong> 10 seconds per request</li>
              </ul>
              <p>
                A webhook is considered successful if your endpoint returns an HTTP status code in the range 
                2xx (200-299). Any other status or timeout will trigger a retry. After 3 attempts, the webhook 
                delivery is abandoned.
              </p>
            </section>

            <section id="webhook-examples" className="doc-section">
              <h2>Webhook Event Examples</h2>

              <div className="webhook-example">
                <h4>invoice.created</h4>
                <CodeBlock
                  language="json"
                  code={`{
  "event_id": "1234567890123456789",
  "event_type": "invoice.created",
  "created_at": "2026-04-25T10:30:00Z",
  "data": {
    "invoice_id": "uuid-inv-123",
    "customer_id": "uuid-cust-456",
    "amount_xrp": "100.00",
    "amount_usd": "200.00",
    "status": "pending",
    "created_at": "2026-04-25T10:30:00Z"
  }
}`}
                />
              </div>

              <div className="webhook-example">
                <h4>invoice.confirmed</h4>
                <CodeBlock
                  language="json"
                  code={`{
  "event_id": "1234567890123456790",
  "event_type": "invoice.confirmed",
  "created_at": "2026-04-25T10:35:00Z",
  "data": {
    "invoice_id": "uuid-inv-123",
    "customer_id": "uuid-cust-456",
    "amount_xrp": "100.00",
    "amount_usd": "200.00",
    "status": "confirmed",
    "paid_at": "2026-04-25T10:32:00Z",
    "confirmations": 5
  }
}`}
                />
              </div>

              <div className="webhook-example">
                <h4>customer.created</h4>
                <CodeBlock
                  language="json"
                  code={`{
  "event_id": "1234567890123456791",
  "event_type": "customer.created",
  "created_at": "2026-04-25T10:40:00Z",
  "data": {
    "customer_id": "uuid-cust-789",
    "first_name": "Jane",
    "last_name": "Smith",
    "email": "jane@example.com",
    "created_at": "2026-04-25T10:40:00Z"
  }
}`}
                />
              </div>

              <div className="webhook-example">
                <h4>webhook.test</h4>
                <CodeBlock
                  language="json"
                  code={`{
  "event_id": "1234567890123456792",
  "event_type": "webhook.test",
  "created_at": "2026-04-25T10:45:00Z",
  "data": {
    "message": "This is a test webhook event"
  }
}`}
                />
              </div>
            </section>
          </div>
        )}
      </div>
    </div>
  );
};

export default ApiDocs;
