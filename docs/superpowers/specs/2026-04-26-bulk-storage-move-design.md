# Design Doc: Bulk Storage Relocation

Allow the admin to move multiple products between storage locations through a consolidated, interactive modal.

## Goal
Improve inventory management efficiency by providing a batch interface for relocating items, especially during warehouse reorganization or restocks.

## User Experience

### 1. Trigger
In the **Admin Inventory Dashboard**, when one or more products are selected via checkboxes, a new action button **"📦 RELOCATE"** appears in the Bulk Action Toolbar.

### 2. The Relocation Modal
A modal dialog titled **"Batch Storage Relocation"** opens.

#### Table Columns (Based on provided reference image):
- **IMAGE**: Small thumbnail of the product.
- **Name**: Product title.
- **Set**: TCG set name (e.g., "Modern Masters").
- **Code**: Collector number.
- **Foil**: Foil treatment label.
- **Variant**: Card treatment (Showcase, Borderless, etc.).
- **Source Location**: The current box/shelf name.
- **Current Stock**: The quantity currently in that location.
  - Interactive: `[ - ] 10 [ + ]`
  - Clicking `-`: Decrements stock, increments "To Move".
  - Clicking `+`: Increments stock (up to original), decrements "To Move" (down to 0).
- **To Move**: Read-only counter showing the delta to be moved.

#### Target Selection (Option A: Global Target):
- A dropdown at the top/side labeled **"TARGET LOCATION (NEW BOX)"**.
- This selection applies to all rows in the current batch.

#### Actions:
- **[Cancel]**: Closes the modal without changes.
- **[Confirm Relocation]**: Sends the batch move request to the server. Disabled if no items are selected to move or no target is chosen.

## Technical Architecture

### Frontend (React/Next.js)
- **Component**: `BulkMoveStorageModal.tsx`
- **State Management**:
  - `selectedProducts`: Array of full product objects (including `stored_in` data).
  - `moves`: A map or array tracking `{ product_id, storage_id, quantity_to_move }`.
  - `targetStorageId`: The ID of the destination box.
- **API Call**: `adminBulkMoveStorage(targetStorageId, moves)`

### Backend (Go/Fiber)
- **Endpoint**: `POST /api/admin/products/bulk-move-storage`
- **Request Body**:
  ```json
  {
    "target_storage_id": "uuid",
    "moves": [
      { "product_id": "uuid", "from_storage_id": "uuid", "quantity": 5 },
      ...
    ]
  }
  ```
- **Store Layer**:
  - `BulkMoveStorage(ctx, targetId, moves)`
  - Uses a database transaction (`tx`).
  - For each move:
    1. Update `product_storage` for `(product_id, from_storage_id)`: `quantity = quantity - move_quantity`.
    2. If new quantity is 0, delete the record.
    3. Update `product_storage` for `(product_id, target_id)`: `quantity = quantity + move_quantity` (Upsert).
    4. Audit log the individual move.

## Data Model Changes
No schema changes required. Uses existing `product_storage` and `storage_location` tables.

## Verification Plan
1. **Unit Tests**:
   - Verify `BulkMoveStorage` correctly handles quantities (no negative stock).
   - Verify record deletion when source stock hits zero.
2. **Integration Tests**:
   - Mock API call from frontend with a batch of 3 products (one with multiple source locations).
3. **Manual Verification**:
   - Select multiple cards.
   - Open modal.
   - "Extract" items from various boxes.
   - Select a new target box.
   - Confirm and verify the stock moved correctly in the main table.
