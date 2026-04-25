# Design Spec: Mass Card Request & Smart Deck Integration
**Date:** 2026-04-25
**Topic:** Mass wanted cards creator, batch processing, and smart deck importer integration.

## 1. Problem Statement
Customers often need to request multiple cards they cannot find in stock. The current system only allows single-card requests. Additionally, the Deck Importer lacks a way to bridge the gap between "No Stock" and "Wanted Cards," leading to lost potential sales and manual follow-ups.

## 2. Proposed Solution
- **Batch Wanted Creator**: A new interface for submitting multiple card requests at once.
- **Smart Deck Resolver**: An enhanced Deck Importer results page that handles partial matches and version alternates, allowing users to request exactly what's missing.
- **Unified Request Logic**: Standardized backend batch processing for efficiency and data integrity.

## 3. Data Model Changes

### 3.1 Client Request Table (`client_request`)
Add columns:
- `quantity`: `INTEGER DEFAULT 1` (to track how many of each card are wanted).
- `tcg`: `TEXT DEFAULT 'mtg'` (to categorize by game).

## 4. Components

### 4.1 Backend
- **Stored Function**: `fn_submit_client_requests_batch` (Handles customer linking and bulk insert).
- **API Endpoint**: `POST /api/client-requests/batch` (Batch submission).

### 4.2 Frontend
- **WantedPage (`/wanted`)**: 
    - Mass input text area (Uses Deck Parser).
    - Review Table (displays parsed Qty, Name, Set, TCG, Details).
    - Batch Submit button.
- **Enhanced DeckImporter**:
    - **Alternate Version Picker**: Shows in-stock alternates when a specific version is missing.
    - **Smart Request Buttons**: Pre-fills missing quantities for partial matches.
    - **Wanted List Sidebar**: Accumulates requests made during the import session.

## 5. UI/UX
- **Conflict Resolution UI**: In the deck importer results, if a specific set is not in stock, show "Available Sets" with a button to swap/add to cart.
- **Logged-in Requirement**: Clearly prompt for login before submitting batch requests.

## 6. Verification
- Verify `client_request` table schema updates.
- Verify batch API creates multiple records correctly.
- Verify Deck Importer correctly identifies missing quantities and versions.
