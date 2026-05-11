# Finalizing MTG Variant Metadata Integration

- [x] Fix `scryfall.go` compilation errors (missing `FrameEffects` field and literal assignment)
- [x] Fix `bounty_service.go` argument mismatch in `SubmitRequest`
- [x] Fix `ProductEditModal.tsx` duplicate `scryfall_id` property
- [x] Update `models.go` to support `StringArray` for `frame_effects`
- [x] Update `external/scryfall.go` mapping to populate `frame_effects`
- [x] Create database migration `20260510_add_frame_effects.sql`
- [x] Update base schema files (`product.sql`, `deck_card.sql`, `client_request.sql`, `bounty.sql`, `external_scryfall.sql`)
- [x] Verify backend build (Exit code: 0)
- [x] Review `backfill_treatments.go` script for GCloud readiness
- [ ] Run `backfill_treatments.go -sync` on GCloud (Manual Deployment)
