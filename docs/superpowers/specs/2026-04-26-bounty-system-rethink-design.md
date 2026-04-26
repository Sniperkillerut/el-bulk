# Design Spec: Integrated Bounty & Client Request Ecosystem
**Date:** 2026-04-26
**Topic:** Overhauling the Bounty system to bridge Supply (Offers) and Demand (Requests) through an aggregate model.

## 1. Problem Statement
The current Bounty system feels "clunky and disconnected." 
- Bounties and Client Requests live in separate silos.
- Admin must manually sync data between them.
- Matching incoming Offers to waiting Clients is a manual, name-based process.
- Clients lack a premium way to specify "Any Version" vs. "Specific Variant."

## 2. Proposed Solution: The "Aggregate Bridge" Model
We will transform the **Bounty** into a central Hub that aggregates demand from multiple clients and supply from multiple sellers.

### 2.1 The Core Architecture
- **Bounty (The Hub)**: Represents a specific "Wanted" card identity (Generic or Specific).
- **ClientRequest (The Demand)**: Explicitly links to a `Bounty`.
- **BountyOffer (The Supply)**: Explicitly links to a `Bounty`.

## 3. Data Model Changes

### 3.1 `client_request` Table
- **ADD** `bounty_id` (UUID, FK -> `bounty.id`, nullable).
- **ADD** `match_type` (TEXT, default 'any'): Values `['any', 'exact']`.
- **ADD** `scryfall_id` (TEXT, nullable): To precisely track the desired variant.

### 3.2 `bounty` Table
- **REMOVE** `request_id` (Redundant with the new reverse link).
- **ADD** `is_generic` (BOOLEAN, default false): True if this bounty accepts ANY version of the card.
- **MODIFY** `quantity_needed`: Represents the aggregate sum of linked requests + admin buffer.

## 4. Workflows & Logic

### 4.1 Admin Triage (Demand Side)
When an Admin accepts a `Pending` request:
1.  **Identity Matching**: Search for an `Active` Bounty matching the card identity.
    - If `match_type` is 'any': Find a Bounty where `is_generic` is true.
    - If `match_type` is 'exact': Find a Bounty with matching Set/Foil/Collector Number.
2.  **Linking**: 
    - If match found: Link request to Bounty, increment `bounty.quantity_needed`.
    - If no match: Create a new Bounty, link request, set target quantity.

### 4.2 Integrated Resolution (Supply Side)
When an Admin resolves a `BountyOffer`:
1.  **Fulfillment Modal**: Displays all `Accepted` requests linked to the parent Bounty.
2.  **Compatibility Filtering**: 
    - Specific Offer (e.g. Foil) matches:
        - Clients who wanted "Any" version.
        - Clients who wanted "Exact" Foil.
    - Generic/Regular Offer matches:
        - Clients who wanted "Any" version.
3.  **Allocation**: Admin selects clients to fulfill.
4.  **Auto-Sync**: 
    - Fulfilled requests set to `solved`.
    - Bounty quantity decremented.
    - If bounty qty reaches 0, set `is_active = false`.

## 5. UI/UX: The Premium Client Interface

### 5.1 "Wanted" Page Overhaul
- **Scryfall Search**: Real-time autocomplete for card names.
- **Visual Variant Picker**:
    - Grid of card images for all prints/variants.
    - "ANY VERSION" toggle: Prominent, high-level choice.
    - If "Specific" selected: Click an image to lock in Scryfall metadata.
- **Quantity & Details**: Standard inputs but with better visual feedback.

### 5.2 Deck Importer Integration
- The "Request Missing" button will now trigger the same Premium Picker for cards not found in stock.

## 6. Performance & Scale
- **Grouping**: Admin dashboard will default to the "Bounties" tab as the source of truth, with expandable rows showing linked requests.
- **Lazy Loading**: Scryfall prints will be fetched only when a card is selected in the UI.

## 7. Verification Plan
- **Migration**: Verify `bounty_id` and `match_type` columns.
- **Fulfillment**: Test that resolving an offer correctly decrements the aggregate bounty and solves linked requests.
- **Sync**: Verify that cancelling a request updates the parent bounty quantity.
