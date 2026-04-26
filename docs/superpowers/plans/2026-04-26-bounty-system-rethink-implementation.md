# Bounty System Rethink: Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Transform Bounties into aggregate hubs that formally link ClientRequests (demand) and BountyOffers (supply), with a premium Scryfall-powered variant picker for the client request form.

**Architecture:** Add `bounty_id` and `match_type` to `client_request`, remove `request_id` from `bounty`, add `is_generic` to `bounty`. Admin acceptance auto-links requests to bounties and aggregates quantity. Offer resolution shows only bounty-linked requests filtered by compatibility.

**Tech Stack:** PostgreSQL migrations, Go (sqlx), Next.js 15 (App Router), TypeScript, Scryfall API

---

## File Map

| File | Change |
|------|--------|
| `backend/db/migrations/20260427_bounty_request_bridge.sql` | NEW — add bounty_id/match_type/scryfall_id to client_request, is_generic to bounty, remove request_id from bounty |
| `backend/db/schema/functions/fn_accept_client_request.sql` | NEW — atomic accept: find/create bounty, link request, increment qty |
| `backend/db/schema/functions/fn_fulfill_bounty_offer.sql` | NEW — atomic fulfill: mark requests solved, decrement bounty qty |
| `backend/models/bounty.go` | MODIFY — remove RequestID, add IsGeneric |
| `backend/models/models.go` | MODIFY — add BountyID, MatchType, ScryfallID to ClientRequest/Input |
| `backend/store/bounty_store.go` | MODIFY — update all SELECTs, add AcceptRequest, FulfillOffer, ListRequestsByBounty |
| `backend/service/bounty_service.go` | MODIFY — add AcceptRequest, FulfillOffer service methods |
| `backend/handlers/bounty_handlers.go` | MODIFY — add AcceptRequest, FulfillOffer endpoints |
| `backend/main.go` | MODIFY — register new routes |
| `frontend/src/lib/types.ts` | MODIFY — ClientRequest add bounty_id/match_type/scryfall_id; Bounty add is_generic, remove request_id |
| `frontend/src/lib/api.ts` | MODIFY — add adminAcceptRequest, adminFulfillOffer API helpers |
| `frontend/src/components/ClientRequestModal.tsx` | MODIFY — replace text inputs with Scryfall search + variant picker |
| `frontend/src/components/admin/BountyOfferResolveModal.tsx` | MODIFY — show only bounty-linked requests, grouped by compatibility |
| `frontend/src/app/admin/bounties/page.tsx` | MODIFY — wire new accept/fulfill flows, remove name-based fallback matching |

---

## Task 1: Database Migration

**Files:**
- Create: `backend/db/migrations/20260427_bounty_request_bridge.sql`

- [ ] **Step 1: Write migration**

```sql
-- 1. Add bounty_id, match_type, scryfall_id to client_request
ALTER TABLE client_request
  ADD COLUMN bounty_id UUID REFERENCES bounty(id) ON DELETE SET NULL,
  ADD COLUMN match_type TEXT NOT NULL DEFAULT 'any' CHECK (match_type IN ('any', 'exact')),
  ADD COLUMN scryfall_id TEXT;

CREATE INDEX idx_client_request_bounty_id ON client_request(bounty_id);

-- 2. Add is_generic to bounty
ALTER TABLE bounty ADD COLUMN is_generic BOOLEAN NOT NULL DEFAULT false;

-- 3. Remove old request_id from bounty (reverse link replaced by bounty_id on client_request)
ALTER TABLE bounty DROP COLUMN IF EXISTS request_id;
```

- [ ] **Step 2: Apply migration**

```bash
docker exec el_bulk_db psql -U elbulk -d elbulk -f /dev/stdin < backend/db/migrations/20260427_bounty_request_bridge.sql
```

Expected: `ALTER TABLE`, `CREATE INDEX`, `ALTER TABLE`, `ALTER TABLE`

- [ ] **Step 3: Verify schema**

```bash
docker exec el_bulk_db psql -U elbulk -d elbulk -c "\d client_request" | grep -E "bounty_id|match_type|scryfall_id"
docker exec el_bulk_db psql -U elbulk -d elbulk -c "\d bounty" | grep -E "is_generic|request_id"
```

Expected: bounty_id/match_type/scryfall_id present in client_request; is_generic present and request_id absent from bounty.

- [ ] **Step 4: Commit**

```bash
git add backend/db/migrations/20260427_bounty_request_bridge.sql
git commit -m "feat(db): add bounty_id/match_type to client_request, is_generic to bounty"
```

---

## Task 2: DB Function — fn_accept_client_request

**Files:**
- Create: `backend/db/schema/functions/fn_accept_client_request.sql`

- [ ] **Step 1: Write function**

```sql
-- Atomically accepts a client_request:
--   1. Find an existing active bounty matching card identity (exact or generic).
--   2. If none, create a new bounty.
--   3. Link the request to the bounty, increment bounty.quantity_needed, set request status = 'accepted'.
CREATE OR REPLACE FUNCTION fn_accept_client_request(
    p_request_id UUID
) RETURNS JSONB AS $$
DECLARE
    v_req         client_request%ROWTYPE;
    v_bounty_id   UUID;
    v_is_generic  BOOLEAN;
BEGIN
    -- Lock and fetch the request
    SELECT * INTO v_req FROM client_request WHERE id = p_request_id FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'request not found';
    END IF;

    v_is_generic := (v_req.match_type = 'any');

    IF v_is_generic THEN
        -- Find existing generic bounty (name + tcg match, is_generic = true)
        SELECT id INTO v_bounty_id
        FROM bounty
        WHERE lower(name) = lower(v_req.card_name)
          AND tcg = v_req.tcg
          AND is_generic = true
          AND is_active = true
        LIMIT 1;
    ELSE
        -- Find existing specific bounty (name + set_name + tcg, is_generic = false)
        SELECT id INTO v_bounty_id
        FROM bounty
        WHERE lower(name) = lower(v_req.card_name)
          AND tcg = v_req.tcg
          AND is_generic = false
          AND is_active = true
          AND (set_name IS NOT DISTINCT FROM v_req.set_name)
        LIMIT 1;
    END IF;

    -- Create bounty if not found
    IF v_bounty_id IS NULL THEN
        INSERT INTO bounty (
            name, tcg, set_name, quantity_needed, is_active, is_generic,
            foil_treatment, card_treatment, language, hide_price, price_source
        ) VALUES (
            v_req.card_name, v_req.tcg, v_req.set_name, v_req.quantity,
            true, v_is_generic,
            'non_foil', 'normal', 'en', false, 'tcgplayer'
        )
        RETURNING id INTO v_bounty_id;
    ELSE
        -- Increment quantity on existing bounty
        UPDATE bounty
        SET quantity_needed = quantity_needed + v_req.quantity,
            updated_at = now()
        WHERE id = v_bounty_id;
    END IF;

    -- Link request to bounty, set accepted
    UPDATE client_request
    SET bounty_id = v_bounty_id,
        status = 'accepted'
    WHERE id = p_request_id;

    RETURN jsonb_build_object(
        'bounty_id', v_bounty_id,
        'request_id', p_request_id,
        'is_generic', v_is_generic
    );
END;
$$ LANGUAGE plpgsql;
```

- [ ] **Step 2: Apply to DB**

```bash
docker exec el_bulk_db psql -U elbulk -d elbulk -f /dev/stdin < backend/db/schema/functions/fn_accept_client_request.sql
```

Expected: `CREATE FUNCTION`

- [ ] **Step 3: Smoke-test**

```bash
# Replace <uuid> with a real pending request ID
docker exec el_bulk_db psql -U elbulk -d elbulk -c "SELECT fn_accept_client_request('<uuid>'::UUID);"
```

Expected: JSON with bounty_id and request_id.

- [ ] **Step 4: Commit**

```bash
git add backend/db/schema/functions/fn_accept_client_request.sql
git commit -m "feat(db): add fn_accept_client_request — links request to bounty atomically"
```

---

## Task 3: DB Function — fn_fulfill_bounty_offer

**Files:**
- Create: `backend/db/schema/functions/fn_fulfill_bounty_offer.sql`

- [ ] **Step 1: Write function**

```sql
-- Atomically fulfills a bounty offer against selected client requests.
--   p_offer_id: the BountyOffer being accepted.
--   p_request_ids: UUID[] of ClientRequests to mark as solved.
CREATE OR REPLACE FUNCTION fn_fulfill_bounty_offer(
    p_offer_id   UUID,
    p_request_ids UUID[]
) RETURNS JSONB AS $$
DECLARE
    v_offer       bounty_offer%ROWTYPE;
    v_fulfilled   INT;
BEGIN
    SELECT * INTO v_offer FROM bounty_offer WHERE id = p_offer_id FOR UPDATE;
    IF NOT FOUND THEN RAISE EXCEPTION 'offer not found'; END IF;

    -- Mark offer as accepted
    UPDATE bounty_offer SET status = 'accepted', updated_at = now() WHERE id = p_offer_id;

    -- Mark selected requests as solved
    UPDATE client_request
    SET status = 'solved'
    WHERE id = ANY(p_request_ids)
      AND bounty_id = v_offer.bounty_id;

    GET DIAGNOSTICS v_fulfilled = ROW_COUNT;

    -- Decrement bounty quantity
    UPDATE bounty
    SET quantity_needed = GREATEST(0, quantity_needed - v_fulfilled),
        is_active = (quantity_needed - v_fulfilled) > 0,
        updated_at = now()
    WHERE id = v_offer.bounty_id;

    RETURN jsonb_build_object(
        'offer_id',   p_offer_id,
        'bounty_id',  v_offer.bounty_id,
        'fulfilled',  v_fulfilled
    );
END;
$$ LANGUAGE plpgsql;
```

- [ ] **Step 2: Apply**

```bash
docker exec el_bulk_db psql -U elbulk -d elbulk -f /dev/stdin < backend/db/schema/functions/fn_fulfill_bounty_offer.sql
```

Expected: `CREATE FUNCTION`

- [ ] **Step 3: Commit**

```bash
git add backend/db/schema/functions/fn_fulfill_bounty_offer.sql
git commit -m "feat(db): add fn_fulfill_bounty_offer — atomic offer resolution"
```


---

## Task 4: Backend Models

**Files:**
- Modify: `backend/models/bounty.go`
- Modify: `backend/models/models.go`

- [ ] **Step 1: Update `backend/models/bounty.go`**

Replace `RequestID *string` with `IsGeneric bool` in both `Bounty` and `BountyInput`. Update `Redact`:

```go
type Bounty struct {
    ID              string        `db:"id"               json:"id"`
    Name            string        `db:"name"             json:"name"`
    TCG             string        `db:"tcg"              json:"tcg"`
    SetName         *string       `db:"set_name"         json:"set_name,omitempty"`
    Condition       *string       `db:"condition"        json:"condition,omitempty"`
    FoilTreatment   FoilTreatment `db:"foil_treatment"   json:"foil_treatment"`
    CardTreatment   CardTreatment `db:"card_treatment"   json:"card_treatment"`
    CollectorNumber *string       `db:"collector_number" json:"collector_number,omitempty"`
    PromoType       *string       `db:"promo_type"       json:"promo_type,omitempty"`
    Language        string        `db:"language"         json:"language"`
    TargetPrice     *float64      `db:"target_price"     json:"target_price,omitempty"`
    HidePrice       bool          `db:"hide_price"       json:"hide_price"`
    QuantityNeeded  int           `db:"quantity_needed"  json:"quantity_needed"`
    IsGeneric       bool          `db:"is_generic"       json:"is_generic"`
    ImageURL        *string       `db:"image_url"        json:"image_url,omitempty"`
    PriceSource     string        `db:"price_source"     json:"price_source,omitempty"`
    PriceReference  *float64      `db:"price_reference"  json:"price_reference,omitempty"`
    IsActive        bool          `db:"is_active"        json:"is_active"`
    CreatedAt       *time.Time    `db:"created_at"       json:"created_at,omitempty"`
    UpdatedAt       *time.Time    `db:"updated_at"       json:"updated_at,omitempty"`
}

func (b *Bounty) Redact(isAdmin bool) {
    if !isAdmin {
        b.PriceSource = ""
        b.PriceReference = nil
        b.CreatedAt = nil
        b.UpdatedAt = nil
        if b.HidePrice {
            b.TargetPrice = nil
        }
    }
}

type BountyInput struct {
    Name            string        `json:"name"`
    TCG             string        `json:"tcg"`
    SetName         *string       `json:"set_name,omitempty"`
    Condition       *string       `json:"condition,omitempty"`
    FoilTreatment   FoilTreatment `json:"foil_treatment"`
    CardTreatment   CardTreatment `json:"card_treatment"`
    CollectorNumber *string       `json:"collector_number,omitempty"`
    PromoType       *string       `json:"promo_type,omitempty"`
    Language        string        `json:"language"`
    TargetPrice     *float64      `json:"target_price,omitempty"`
    HidePrice       bool          `json:"hide_price"`
    QuantityNeeded  int           `json:"quantity_needed"`
    IsGeneric       bool          `json:"is_generic"`
    ImageURL        *string       `json:"image_url,omitempty"`
    PriceSource     string        `json:"price_source"`
    PriceReference  *float64      `json:"price_reference,omitempty"`
    IsActive        *bool         `json:"is_active,omitempty"`
}
```

- [ ] **Step 2: Update `ClientRequest` in `backend/models/models.go`**

Add three fields to `ClientRequest` struct (after `Status`):

```go
BountyID           *string    `db:"bounty_id"            json:"bounty_id,omitempty"`
MatchType          string     `db:"match_type"           json:"match_type"`
ScryfallID         *string    `db:"scryfall_id"          json:"scryfall_id,omitempty"`
```

Add to `ClientRequestInput`:

```go
MatchType  string  `json:"match_type"` // "any" or "exact"
ScryfallID *string `json:"scryfall_id,omitempty"`
```

- [ ] **Step 3: Add `FulfillOfferInput` to `backend/models/models.go`**

```go
type FulfillOfferInput struct {
    RequestIDs []string `json:"request_ids"`
}
```

- [ ] **Step 4: Verify build**

```bash
docker exec el_bulk_backend go build ./...
```

Expected: no output (clean build).

- [ ] **Step 5: Commit**

```bash
git add backend/models/bounty.go backend/models/models.go
git commit -m "feat(models): replace request_id with is_generic on Bounty; add BountyID/MatchType to ClientRequest"
```

---

## Task 5: Store Layer

**Files:**
- Modify: `backend/store/bounty_store.go`

- [ ] **Step 1: Update all column lists to replace `request_id` with `is_generic`**

In `ListBounties`, `CreateBounty`, `UpdateBounty` — replace every occurrence of `request_id` with `is_generic` in both the SELECT/INSERT/UPDATE column lists and the Go argument bindings.

`ListBounties` SELECT line becomes:
```sql
b.id, b.name, b.tcg, b.set_name, b.condition, b.foil_treatment, b.card_treatment,
b.collector_number, b.promo_type, b.language, b.target_price, b.hide_price,
b.quantity_needed, b.is_generic, b.image_url, b.price_source, b.price_reference,
b.is_active, b.created_at, b.updated_at
```

`CreateBounty` INSERT columns: replace `request_id` → `is_generic`, argument `input.RequestID` → `input.IsGeneric`.

`UpdateBounty` SET clause: replace `request_id = $13` → `is_generic = $13`, argument `input.RequestID` → `input.IsGeneric`.

- [ ] **Step 2: Update `ListRequests` and `UpdateRequestStatus` to include new columns**

```go
// ListRequests query — add bounty_id, match_type, scryfall_id
query := `
    SELECT id, customer_id, customer_name, customer_contact, card_name, set_name,
           details, quantity, tcg, status, cancellation_reason, bounty_id, match_type, scryfall_id, created_at
    FROM client_request
    ORDER BY created_at DESC
`
// Same addition for ListMeRequests and UpdateRequestStatus RETURNING clause
```

- [ ] **Step 3: Add `AcceptRequest` to store**

```go
func (s *BountyStore) AcceptRequest(ctx context.Context, requestID string) (map[string]interface{}, error) {
    var result []byte
    err := s.DB.GetContext(ctx, &result, "SELECT fn_accept_client_request($1::UUID)", requestID)
    if err != nil {
        return nil, err
    }
    var out map[string]interface{}
    if err := json.Unmarshal(result, &out); err != nil {
        return nil, err
    }
    return out, nil
}
```

- [ ] **Step 4: Add `FulfillOffer` to store**

```go
func (s *BountyStore) FulfillOffer(ctx context.Context, offerID string, requestIDs []string) (map[string]interface{}, error) {
    var result []byte
    err := s.DB.GetContext(ctx, &result, "SELECT fn_fulfill_bounty_offer($1::UUID, $2::UUID[])", offerID, pq.Array(requestIDs))
    if err != nil {
        return nil, err
    }
    var out map[string]interface{}
    if err := json.Unmarshal(result, &out); err != nil {
        return nil, err
    }
    return out, nil
}
```

Note: `pq.Array` requires `"github.com/lib/pq"` in imports.

- [ ] **Step 5: Add `ListRequestsByBounty` to store**

```go
func (s *BountyStore) ListRequestsByBounty(ctx context.Context, bountyID string) ([]models.ClientRequest, error) {
    requests := []models.ClientRequest{}
    query := `
        SELECT id, customer_id, customer_name, customer_contact, card_name, set_name,
               details, quantity, tcg, status, cancellation_reason, bounty_id, match_type, scryfall_id, created_at
        FROM client_request
        WHERE bounty_id = $1 AND status IN ('accepted', 'pending')
        ORDER BY created_at ASC
    `
    err := s.DB.SelectContext(ctx, &requests, query, bountyID)
    return requests, err
}
```

- [ ] **Step 6: Build check**

```bash
docker exec el_bulk_backend go build ./...
```

- [ ] **Step 7: Commit**

```bash
git add backend/store/bounty_store.go
git commit -m "feat(store): wire AcceptRequest, FulfillOffer, ListRequestsByBounty; update column lists"
```

---

## Task 6: Service + Handler + Routes

**Files:**
- Modify: `backend/service/bounty_service.go`
- Modify: `backend/handlers/bounty_handlers.go`
- Modify: `backend/main.go`

- [ ] **Step 1: Add service methods to `bounty_service.go`**

```go
func (s *BountyService) AcceptRequest(ctx context.Context, requestID string) (map[string]interface{}, error) {
    return s.Store.AcceptRequest(ctx, requestID)
}

func (s *BountyService) FulfillOffer(ctx context.Context, offerID string, requestIDs []string) (map[string]interface{}, error) {
    if len(requestIDs) == 0 {
        return nil, fmt.Errorf("at least one request_id is required")
    }
    return s.Store.FulfillOffer(ctx, offerID, requestIDs)
}

func (s *BountyService) ListRequestsByBounty(ctx context.Context, bountyID string) ([]models.ClientRequest, error) {
    requests, err := s.Store.ListRequestsByBounty(ctx, bountyID)
    if err != nil {
        return nil, err
    }
    for i := range requests {
        requests[i].CustomerContact = *crypto.DecryptSafe(&requests[i].CustomerContact)
    }
    return requests, nil
}
```

- [ ] **Step 2: Add handlers to `bounty_handlers.go`**

```go
func (h *BountyHandler) AcceptRequest(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    result, err := h.Service.AcceptRequest(r.Context(), id)
    if err != nil {
        render.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    render.Success(w, result)
}

func (h *BountyHandler) FulfillOffer(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    var input models.FulfillOfferInput
    if err := decodeJSON(r, &input); err != nil {
        render.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    result, err := h.Service.FulfillOffer(r.Context(), id, input.RequestIDs)
    if err != nil {
        render.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    render.Success(w, result)
}

func (h *BountyHandler) ListRequestsByBounty(w http.ResponseWriter, r *http.Request) {
    bountyID := chi.URLParam(r, "id")
    requests, err := h.Service.ListRequestsByBounty(r.Context(), bountyID)
    if err != nil {
        render.Error(w, "Failed to fetch requests", http.StatusInternalServerError)
        return
    }
    render.Success(w, requests)
}
```

- [ ] **Step 3: Register routes in `backend/main.go`**

Find the admin bounties route block and add:

```go
r.Post("/api/admin/client-requests/{id}/accept", bh.AcceptRequest)
r.Post("/api/admin/bounties/offers/{id}/fulfill", bh.FulfillOffer)
r.Get("/api/admin/bounties/{id}/requests", bh.ListRequestsByBounty)
```

- [ ] **Step 4: Build and verify**

```bash
docker exec el_bulk_backend go build ./...
```

- [ ] **Step 5: Commit**

```bash
git add backend/service/bounty_service.go backend/handlers/bounty_handlers.go backend/main.go
git commit -m "feat(api): add accept-request, fulfill-offer, list-requests-by-bounty endpoints"
```


---

## Task 7: Frontend Types & API Helpers

**Files:**
- Modify: `frontend/src/lib/types.ts`
- Modify: `frontend/src/lib/api.ts`

- [ ] **Step 1: Update `Bounty` interface in `types.ts`**

Remove `request_id?: string`, add `is_generic: boolean`.

- [ ] **Step 2: Update `BountyInput` in `types.ts`**

Remove `request_id?: string`, add `is_generic: boolean`.

- [ ] **Step 3: Update `ClientRequest` in `types.ts`**

Add:
```typescript
bounty_id?: string;
match_type: 'any' | 'exact';
scryfall_id?: string;
```

- [ ] **Step 4: Update `ClientRequestInput` in `types.ts`**

Add:
```typescript
match_type?: 'any' | 'exact';
scryfall_id?: string;
```

- [ ] **Step 5: Add API helpers to `api.ts`**

```typescript
export async function adminAcceptClientRequest(id: string): Promise<{ bounty_id: string; request_id: string; is_generic: boolean }> {
  return apiFetch(`/api/admin/client-requests/${id}/accept`, { method: 'POST' });
}

export async function adminFulfillBountyOffer(offerId: string, requestIds: string[]): Promise<{ offer_id: string; bounty_id: string; fulfilled: number }> {
  return apiFetch(`/api/admin/bounties/offers/${offerId}/fulfill`, {
    method: 'POST',
    body: JSON.stringify({ request_ids: requestIds }),
  });
}

export async function adminFetchRequestsByBounty(bountyId: string): Promise<import('./types').ClientRequest[]> {
  return apiFetch(`/api/admin/bounties/${bountyId}/requests`, { cache: 'no-store' });
}
```

- [ ] **Step 6: Commit**

```bash
git add frontend/src/lib/types.ts frontend/src/lib/api.ts
git commit -m "feat(frontend): update types and API helpers for bounty bridge model"
```

---

## Task 8: Admin Bounties Page — Wire New Accept Flow

**Files:**
- Modify: `frontend/src/app/admin/bounties/page.tsx`

- [ ] **Step 1: Replace `handleAcceptRequest`**

Remove the old implementation that called `adminUpdateClientRequestStatus` and pre-filled bounty modal. Replace with:

```typescript
const handleAcceptRequest = async (req: ClientRequest) => {
  try {
    await adminAcceptClientRequest(req.id);
    handleRefresh();
  } catch {
    alert(t('pages.admin.bounties.error_accept', 'Failed to accept request.'));
  }
};
```

- [ ] **Step 2: Remove name-based fallback warning badge**

Replace the entire IIFE `(() => { if (b.request_id) ... })()` badge logic with a simpler check using `is_generic`:

```tsx
{b.is_generic && (
  <span className="badge bg-blue-100 text-blue-600 text-[8px] font-mono-stack border border-blue-200">
    {t('pages.admin.bounties.generic_badge', 'ANY VERSION')}
  </span>
)}
```

- [ ] **Step 3: Add "No Longer Needed" warning using bounty-linked requests**

Since requests now properly link to bounties, load linked requests per bounty via the new endpoint when a bounty row is expanded. Add state:

```typescript
const [linkedRequests, setLinkedRequests] = useState<Record<string, ClientRequest[]>>({});

const loadLinkedRequests = async (bountyId: string) => {
  if (linkedRequests[bountyId]) return;
  const reqs = await adminFetchRequestsByBounty(bountyId);
  setLinkedRequests(prev => ({ ...prev, [bountyId]: reqs }));
};
```

- [ ] **Step 4: Show not_needed warning badge using linked data**

```tsx
{(linkedRequests[b.id] || []).some(r => r.status === 'not_needed') && (
  <span className="badge bg-red-100 text-red-600 text-[8px] font-bold animate-pulse border border-red-200">
    ⚠️ {t('pages.admin.bounties.request_cancelled_warning', 'CLIENT NO LONGER NEEDS THIS')}
  </span>
)}
```

- [ ] **Step 5: Commit**

```bash
git add frontend/src/app/admin/bounties/page.tsx
git commit -m "feat(admin): wire adminAcceptClientRequest; use bounty-linked requests for warnings"
```

---

## Task 9: Admin Resolve Modal — Linked Request Fulfillment

**Files:**
- Modify: `frontend/src/components/admin/BountyOfferResolveModal.tsx`
- Modify: `frontend/src/app/admin/bounties/page.tsx`

- [ ] **Step 1: Update `BountyOfferResolveModal` props**

Change `requests: ClientRequest[]` (all requests) to `linkedRequests: ClientRequest[]` (only bounty-linked).

Remove the name-based filter:
```typescript
// DELETE this line:
const relatedRequests = requests.filter(r => r.card_name.toLowerCase()...);
// REPLACE with:
const relatedRequests = linkedRequests;
```

- [ ] **Step 2: Group requests by compatibility in modal**

```tsx
const exactMatches = relatedRequests.filter(r => r.match_type === 'exact');
const genericMatches = relatedRequests.filter(r => r.match_type === 'any');
```

Render two sections in the modal UI:

```tsx
{exactMatches.length > 0 && (
  <div>
    <p className="text-[10px] font-mono-stack uppercase text-emerald-600 mb-2">🎯 Direct Matches</p>
    {exactMatches.map(r => <RequestCheckbox key={r.id} request={r} ... />)}
  </div>
)}
{genericMatches.length > 0 && (
  <div className="mt-3">
    <p className="text-[10px] font-mono-stack uppercase text-blue-600 mb-2">🤝 Any Version</p>
    {genericMatches.map(r => <RequestCheckbox key={r.id} request={r} ... />)}
  </div>
)}
```

- [ ] **Step 3: Update `onAccept` callback signature in page.tsx**

Replace the `handleResolveOffer` logic that used `adminUpdateBountyOfferStatus` + manual request loop with:

```typescript
const handleFulfillOffer = async (requestIds: string[]) => {
  if (!resolvingOffer) return;
  try {
    await adminFulfillBountyOffer(resolvingOffer.offer.id, requestIds);
    handleRefresh();
  } catch {
    alert(t('pages.admin.bounties.error_fulfill', 'Failed to fulfill offer.'));
  } finally {
    setResolvingOffer(null);
  }
};
```

- [ ] **Step 4: Load linked requests when opening resolve modal**

```typescript
const handleOpenResolveModal = async (offer: BountyOffer, bounty: Bounty) => {
  setResolvingOffer({ offer, bounty });
  const reqs = await adminFetchRequestsByBounty(bounty.id);
  setLinkedRequests(prev => ({ ...prev, [bounty.id]: reqs }));
};
```

Pass `linkedRequests[resolvingOffer.bounty.id] ?? []` as `linkedRequests` to the modal.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/admin/BountyOfferResolveModal.tsx frontend/src/app/admin/bounties/page.tsx
git commit -m "feat(admin): linked-request fulfillment modal with exact/generic grouping"
```

---

## Task 10: Premium Client Request Form — Scryfall Variant Picker

**Files:**
- Create: `frontend/src/components/ScryfallVariantPicker.tsx`
- Modify: `frontend/src/components/ClientRequestModal.tsx`

- [ ] **Step 1: Create `ScryfallVariantPicker.tsx`**

```tsx
'use client';
import { useState, useEffect } from 'react';

interface ScryfallPrint {
  id: string;
  set_name: string;
  set: string;
  collector_number: string;
  image_uris?: { small: string; normal: string };
  card_faces?: { image_uris?: { small: string } }[];
  foil: boolean;
  nonfoil: boolean;
  finishes: string[];
}

interface Props {
  cardName: string;
  onSelect: (print: ScryfallPrint | null) => void; // null = "Any Version"
  selectedId?: string;
}

export default function ScryfallVariantPicker({ cardName, onSelect, selectedId }: Props) {
  const [prints, setPrints] = useState<ScryfallPrint[]>([]);
  const [loading, setLoading] = useState(false);
  const [anyVersion, setAnyVersion] = useState(true);

  useEffect(() => {
    if (!cardName) return;
    setLoading(true);
    fetch(`https://api.scryfall.com/cards/search?q=!"${encodeURIComponent(cardName)}"&unique=prints&order=released`)
      .then(r => r.json())
      .then(data => setPrints(data.data || []))
      .catch(() => setPrints([]))
      .finally(() => setLoading(false));
  }, [cardName]);

  const getImage = (p: ScryfallPrint) =>
    p.image_uris?.small ?? p.card_faces?.[0]?.image_uris?.small ?? '/placeholder-card.png';

  return (
    <div className="space-y-3">
      {/* Any Version Toggle */}
      <label className="flex items-center gap-3 cursor-pointer group">
        <div
          onClick={() => { setAnyVersion(!anyVersion); if (!anyVersion) onSelect(null); }}
          className={`w-10 h-6 rounded-full transition-colors relative ${anyVersion ? 'bg-gold' : 'bg-kraft-dark/30'}`}
        >
          <div className={`w-4 h-4 bg-white rounded-full absolute top-1 transition-all shadow ${anyVersion ? 'left-5' : 'left-1'}`} />
        </div>
        <span className="text-xs font-mono-stack uppercase text-text-muted group-hover:text-ink-deep">
          Any Version — I'll take whatever you find
        </span>
      </label>

      {/* Variant Grid */}
      {!anyVersion && (
        <div>
          {loading && <p className="text-xs text-text-muted animate-pulse">Searching prints...</p>}
          <div className="grid grid-cols-4 gap-2 max-h-64 overflow-y-auto">
            {prints.map(p => (
              <button
                key={p.id}
                type="button"
                onClick={() => onSelect(p)}
                className={`rounded overflow-hidden border-2 transition-all hover:scale-105 ${
                  selectedId === p.id ? 'border-gold shadow-lg shadow-gold/20' : 'border-transparent hover:border-gold/40'
                }`}
                title={`${p.set_name} #${p.collector_number}`}
              >
                <img src={getImage(p)} alt={p.set_name} className="w-full object-cover" loading="lazy" />
                <p className="text-[8px] font-mono-stack text-text-muted px-1 truncate">{p.set}</p>
              </button>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
```

- [ ] **Step 2: Update `ClientRequestModal.tsx`**

Add state for selected variant:

```typescript
const [selectedPrint, setSelectedPrint] = useState<ScryfallPrint | null>(null);
const [matchType, setMatchType] = useState<'any' | 'exact'>('any');
```

Replace the plain `card_name` text input with a two-step flow:
1. Text input for card name (existing).
2. Below it, render `<ScryfallVariantPicker>` when `form.card_name.length > 2`.

```tsx
{form.card_name.length > 2 && (
  <ScryfallVariantPicker
    cardName={form.card_name}
    selectedId={selectedPrint?.id}
    onSelect={(print) => {
      setSelectedPrint(print);
      setMatchType(print ? 'exact' : 'any');
      if (print) {
        setFieldValue('set_name', print.set_name);
      }
    }}
  />
)}
```

Update `onSubmit` to include new fields:

```typescript
const onSubmit = async (data: Record<string, string>) => {
  await createClientRequest({
    ...data,
    match_type: matchType,
    scryfall_id: selectedPrint?.id,
    set_name: selectedPrint?.set_name || data.set_name,
  } as import('@/lib/types').ClientRequestInput);
  onSuccess();
};
```

- [ ] **Step 3: Verify in browser**

Navigate to the Bounties public page, click "Request a card," type a card name (e.g. "Mondrak"), and verify the variant picker appears after 2+ characters.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/ScryfallVariantPicker.tsx frontend/src/components/ClientRequestModal.tsx
git commit -m "feat(ui): Scryfall variant picker with Any/Exact toggle in client request form"
```

---

## Task 11: Update DB Function fn_submit_client_request

**Files:**
- Modify: `backend/db/schema/functions/fn_submit_client_request.sql`

- [ ] **Step 1: Add match_type and scryfall_id params**

```sql
CREATE OR REPLACE FUNCTION fn_submit_client_request(
    p_customer_name TEXT,
    p_customer_contact TEXT,
    p_card_name TEXT,
    p_set_name TEXT DEFAULT NULL,
    p_details TEXT DEFAULT NULL,
    p_quantity INTEGER DEFAULT 1,
    p_tcg TEXT DEFAULT 'mtg',
    p_customer_id UUID DEFAULT NULL,
    p_match_type TEXT DEFAULT 'any',
    p_scryfall_id TEXT DEFAULT NULL
) RETURNS JSONB AS $$
-- ... (existing customer lookup/create logic unchanged) ...
    -- Update INSERT to include new fields:
    INSERT INTO client_request (
        customer_id, customer_name, customer_contact, card_name, set_name,
        details, quantity, tcg, status, match_type, scryfall_id
    ) VALUES (
        v_customer_id, p_customer_name, p_customer_contact, p_card_name, p_set_name,
        p_details, p_quantity, p_tcg, 'pending', p_match_type, p_scryfall_id
    ) RETURNING id, created_at INTO v_request_id, v_created_at;

    RETURN jsonb_build_object(
        'id', v_request_id,
        'customer_id', v_customer_id,
        'customer_name', p_customer_name,
        'customer_contact', p_customer_contact,
        'card_name', p_card_name,
        'set_name', p_set_name,
        'details', p_details,
        'quantity', p_quantity,
        'tcg', p_tcg,
        'status', 'pending',
        'match_type', p_match_type,
        'scryfall_id', p_scryfall_id,
        'created_at', v_created_at
    );
-- ... END $$ LANGUAGE plpgsql;
```

- [ ] **Step 2: Update `BountyStore.SubmitRequest` in Go to pass new args**

```go
func (s *BountyStore) SubmitRequest(ctx context.Context, customerName, customerContact, cardName string,
    setName, details *string, quantity int, tcg string, userID *string,
    matchType string, scryfallID *string) ([]byte, error) {
    var result []byte
    err := s.DB.GetContext(ctx, &result,
        "SELECT fn_submit_client_request($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)",
        customerName, customerContact, cardName, setName, details, quantity, tcg, userID, matchType, scryfallID)
    return result, err
}
```

- [ ] **Step 3: Update `BountyService.SubmitRequest` to pass new args from `ClientRequestInput`**

```go
return s.Store.SubmitRequest(ctx,
    input.CustomerName, input.CustomerContact, input.CardName,
    input.SetName, input.Details, input.Quantity, input.TCG, userID,
    input.MatchType, input.ScryfallID)
```

- [ ] **Step 4: Apply updated function to DB**

```bash
docker exec el_bulk_db psql -U elbulk -d elbulk -f /dev/stdin < backend/db/schema/functions/fn_submit_client_request.sql
```

- [ ] **Step 5: Build and commit**

```bash
docker exec el_bulk_backend go build ./...
git add backend/db/schema/functions/fn_submit_client_request.sql backend/store/bounty_store.go backend/service/bounty_service.go
git commit -m "feat: pass match_type and scryfall_id through request submission pipeline"
```

---

## Task 12: Final Verification

- [ ] **Step 1: Test full accept flow**

1. As a client, submit a request for "Mondrak" with "Any Version."
2. As admin, go to Client Requests tab — click Accept.
3. Verify: a Bounty is created (or existing one linked), quantity incremented.

- [ ] **Step 2: Test specific variant flow**

1. Submit a second request for "Mondrak" with "Exact" — pick a specific Foil print.
2. Accept it as admin — verify it creates a *separate* specific bounty.

- [ ] **Step 3: Test offer fulfillment**

1. Submit a BountyOffer against the generic Mondrak bounty.
2. Open Resolve modal — verify both requests appear (grouped by exact/any).
3. Select one — verify request marked `solved`, bounty qty decremented.

- [ ] **Step 4: Test cancellation sync**

1. Cancel a linked request.
2. Verify the parent bounty's `quantity_needed` decrements correctly.
   ```bash
   docker exec el_bulk_db psql -U elbulk -d elbulk -c \
     "SELECT quantity_needed FROM bounty WHERE id = '<bounty_id>';"
   ```

- [ ] **Step 5: Final commit**

```bash
git add -A
git commit -m "feat: bounty aggregate bridge model — complete integration"
```

