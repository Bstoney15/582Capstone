// embed.go – embeds the checkout.js widget file into the binary for serving by the backend.
package widgetstatic

// Author: Benjamin Stonestreet
// Created: 2026-03-29

import _ "embed"

// CheckoutJS is the embedded checkout widget payload served by the backend.
//
//go:embed checkout.js
var CheckoutJS []byte
