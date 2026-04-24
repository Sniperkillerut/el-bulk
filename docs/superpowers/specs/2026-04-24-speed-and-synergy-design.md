# Design Spec: Speed & Synergy Update
**Date:** 2026-04-24
**Topic:** Real-time delivery indicators and smart "shaft" recommendations.

## 1. Problem Statement
Casual players at "El Bulk" need help discovering "overlooked" cards that synergize with their brews, and local Bogotá customers need immediate confirmation of same-day delivery availability.

## 2. Proposed Solution
- **Bogotá Express Indicator:** A live status badge controlled by business hours and a manual database override.
- **Synergy Scout:** A recommendation engine that suggests high-stock, low-price cards matching the user's current deck colors and set preferences.

## 3. Data Model Changes

### 3.1 Settings Table (`setting`)
New keys to be added:
- `delivery_priority_enabled`: Boolean string ('true'/'false').
- `synergy_max_price_cop`: Numeric string (e.g., '2000').

## 4. Components

### 4.1 DeliveryBadge (Frontend)
- **State:** `active` (Green Pulse) | `offline` (Grey Static).
- **Copy:** "AVAILABLE NOW" | "CURRENTLY OFFLINE".

### 4.2 Recommendation API (Backend)
- **Endpoint:** `GET /api/products/recommendations?ids=uuid[]`
- **Query Logic:**
  1. Extract color identities from provided IDs.
  2. Filter products where `price <= synergy_max_price_cop`.
  3. Filter products where `stock > 10`.
  4. Filter products where `color_identity` overlaps with cart colors.
  5. Sort by `random()` to provide variety.

## 5. UI/UX
- **Hero Section:** Badge placed above the main search bar.
- **Cart Drawer:** Suggestions shown at the bottom, just above the subtotal.
- **Checkout:** Final delivery status badge next to the shipping selection.

## 6. Verification
- Manual toggle of the database flag to verify "Offline" state.
- Adding specific cards to cart to verify synergy accuracy.
