# Design Spec: Client Request Cancellation with Reason

**Status**: Draft
**Date**: 2026-04-26
**Topic**: Client Request Management

## Overview
Clients currently have no way to cancel a card request once it has been "Accepted" by the admin. This design introduces a way for clients to mark requests as "No longer needed" at any stage before completion, providing a reason that helps the admin manage their search efforts.

## Proposed Changes

### 1. Database Schema
Update the `client_request` table to support the new status and store the user's feedback.

- **Status**: Add `not_needed` to the allowed statuses (application-level validation or DB check if exists).
- **New Column**: `cancellation_reason` (TEXT, nullable).

```sql
-- Migration
ALTER TABLE client_request ADD COLUMN cancellation_reason TEXT;
```

### 2. Backend (Go)

#### Models
Update `ClientRequest` and relevant input structs in `backend/models/models.go`.

```go
type ClientRequest struct {
    // ... existing fields
    Status             string  `json:"status" db:"status"`
    CancellationReason *string `json:"cancellation_reason,omitempty" db:"cancellation_reason"`
}

type CancelRequestInput struct {
    Reason  string `json:"reason"`
    Details string `json:"details"`
}
```

#### Store & Service
Update `CancelMeRequest` in `backend/store/bounty_store.go` and `backend/service/bounty_service.go` to:
- Accept `reason` and `details`.
- Allow cancellation if status is `pending` or `accepted`.
- Update the `cancellation_reason` field.

### 3. Frontend (Next.js)

#### Components
- **`CancellationModal`**: A new modal component to collect the reason for cancellation.
    - Predefined reasons: "Found it elsewhere", "Price was too high", "No longer interested", "Other".
    - Text area for "Other" or additional context.

#### Profile Page
- Update the `requests` tab in `frontend/src/app/profile/page.tsx`.
- Enable the "Cancel" action for `accepted` requests.
- Change the label to "No longer needed" for `accepted` requests to improve tone.

#### Admin Dashboard
- Update the admin request list to visually distinguish `not_needed` requests (e.g., strike-through or dimmed).
- Show the `cancellation_reason` when hovering or expanding the request row.

## Success Criteria
- Clients can cancel `pending` and `accepted` requests.
- Admin receives context on why the request was stopped.
- The search list remains clean of "dead" requests.
