# Bulk Storage Relocation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement a batch interface for moving products between storage locations with precise quantity control.

**Architecture:**
- **Backend**: Add `BulkMoveStorage` endpoint. Uses a transaction to decrement source stock and increment target stock (upserting or creating as needed).
- **Frontend**: Add "RELOCATE" button to bulk toolbar. Create `BulkMoveStorageModal` which allows users to "extract" stock from multiple source locations into a single target box.

**Tech Stack:** Go (Fiber), PostgreSQL, TypeScript (Next.js, Tailwind CSS)

---

### Task 1: Backend Models
**Files:**
- Modify: `backend/models/models.go`

- [ ] **Step 1: Add request models for bulk storage move**
```go
type BulkMoveStorageRequest struct {
	TargetStorageID string            `json:"target_storage_id"`
	Moves           []MoveStorageItem `json:"moves"`
}

type MoveStorageItem struct {
	ProductID       string `json:"product_id"`
	FromStorageID   string `json:"from_storage_id"`
	Quantity        int    `json:"quantity"`
}
```

- [ ] **Step 2: Commit**
```bash
git add backend/models/models.go
git commit -m "feat(backend): add bulk storage move request models"
```

---

### Task 2: Store Layer Logic
**Files:**
- Modify: `backend/store/product_store.go`

- [ ] **Step 1: Implement BulkMoveStorage method in ProductStore**
```go
func (s *ProductStore) BulkMoveStorage(ctx context.Context, req models.BulkMoveStorageRequest) error {
	return s.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		for _, move := range req.Moves {
			if move.Quantity <= 0 {
				continue
			}

			// 1. Decrement from source
			_, err := tx.ExecContext(ctx, `
				UPDATE product_storage 
				SET quantity = quantity - $1 
				WHERE product_id = $2 AND storage_id = $3
			`, move.Quantity, move.ProductID, move.FromStorageID)
			if err != nil {
				return fmt.Errorf("failed to decrement source: %w", err)
			}

			// 2. Cleanup source if 0
			_, err = tx.ExecContext(ctx, `
				DELETE FROM product_storage 
				WHERE product_id = $1 AND storage_id = $2 AND quantity = 0
			`, move.ProductID, move.FromStorageID)
			if err != nil {
				return fmt.Errorf("failed to cleanup source: %w", err)
			}

			// 3. Increment target (Upsert)
			_, err = tx.ExecContext(ctx, `
				INSERT INTO product_storage (product_id, storage_id, quantity)
				VALUES ($1, $2, $3)
				ON CONFLICT (product_id, storage_id) 
				DO UPDATE SET quantity = product_storage.quantity + EXCLUDED.quantity
			`, move.ProductID, req.TargetStorageID, move.Quantity)
			if err != nil {
				return fmt.Errorf("failed to increment target: %w", err)
			}
		}
		return nil
	})
}
```
*Note: Ensure `WithTransaction` helper exists in `BaseStore` or implement it manually using `s.DB.BeginTxx`.*

- [ ] **Step 2: Commit**
```bash
git add backend/store/product_store.go
git commit -m "feat(backend): implement BulkMoveStorage in store"
```

---

### Task 3: Service Layer Implementation
**Files:**
- Modify: `backend/service/product_service.go`

- [ ] **Step 1: Add BulkMoveStorage to ProductService**
```go
func (s *ProductService) BulkMoveStorage(ctx context.Context, req models.BulkMoveStorageRequest) error {
	if req.TargetStorageID == "" {
		return errors.New("target storage ID is required")
	}
	if len(req.Moves) == 0 {
		return nil
	}

	err := s.Store.BulkMoveStorage(ctx, req)
	if err != nil {
		return err
	}

	// Audit Log the batch action
	s.Audit.Log(ctx, "bulk_move_storage", "product", "batch", models.JSONB{
		"target_id": req.TargetStorageID,
		"move_count": len(req.Moves),
	})

	return nil
}
```

- [ ] **Step 2: Commit**
```bash
git add backend/service/product_service.go
git commit -m "feat(backend): add BulkMoveStorage to service layer"
```

---

### Task 4: API Handler and Routes
**Files:**
- Modify: `backend/handlers/products.go`
- Modify: `backend/main.go`

- [ ] **Step 1: Implement BulkMoveStorage handler**
```go
func (h *ProductHandler) BulkMoveStorage(w http.ResponseWriter, r *http.Request) {
	var req models.BulkMoveStorageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RenderError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.Service.BulkMoveStorage(r.Context(), req); err != nil {
		logger.ErrorCtx(r.Context(), "BulkMoveStorage failed: %v", err)
		httputil.RenderError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	httputil.RenderSuccess(w, map[string]string{"message": "Relocation complete"})
}
```

- [ ] **Step 2: Register the route in main.go**
```go
r.Post("/products/bulk-move-storage", productHandler.BulkMoveStorage)
```
*(Add near other bulk product routes in the admin group)*

- [ ] **Step 3: Commit**
```bash
git add backend/handlers/products.go backend/main.go
git commit -m "feat(backend): add bulk-move-storage API endpoint"
```

---

### Task 5: Frontend API and Types
**Files:**
- Modify: `frontend/src/lib/api.ts`
- Modify: `frontend/src/lib/types.ts`

- [ ] **Step 1: Add types to types.ts**
```typescript
export interface BulkMoveStorageRequest {
  target_storage_id: string;
  moves: {
    product_id: string;
    from_storage_id: string;
    quantity: number;
  }[];
}
```

- [ ] **Step 2: Add API function to api.ts**
```typescript
export async function adminBulkMoveStorage(req: BulkMoveStorageRequest): Promise<{ message: string }> {
  return apiFetch<{ message: string }>('/api/admin/products/bulk-move-storage', {
    method: 'POST',
    body: JSON.stringify(req),
  });
}
```

- [ ] **Step 3: Commit**
```bash
git add frontend/src/lib/api.ts frontend/src/lib/types.ts
git commit -m "feat(frontend): add bulk move API client"
```

---

### Task 6: BulkMoveStorageModal Component
**Files:**
- Create: `frontend/src/app/admin/dashboard/BulkMoveStorageModal.tsx`

- [ ] **Step 1: Create the modal component with the interactive relocation table**
(Implementation should follow the "extract" logic: `-` button decreases stock and increases `toMove`)

- [ ] **Step 2: Commit**
```bash
git add frontend/src/app/admin/dashboard/BulkMoveStorageModal.tsx
git commit -m "feat(frontend): create BulkMoveStorageModal"
```

---

### Task 7: Integration in Admin Dashboard
**Files:**
- Modify: `frontend/src/app/admin/dashboard/page.tsx`

- [ ] **Step 1: Add "RELOCATE" button to bulk toolbar and state to open the modal**

- [ ] **Step 2: Commit**
```bash
git add frontend/src/app/admin/dashboard/page.tsx
git commit -m "feat(frontend): integrate bulk relocate in dashboard"
```

---

### Task 8: Verification
- [ ] **Step 1: Verify relocation logic**
- [ ] **Step 2: Run backend tests if any**
- [ ] **Step 3: Final Commit**
