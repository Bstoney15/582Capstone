package widgetstatic

import _ "embed"

// CheckoutJS is the embedded checkout widget payload served by the backend.
//
//go:embed checkout.js
var CheckoutJS []byte
