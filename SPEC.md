# §G (Goal)
mtg store admin & public catalog. sync prices from scryfall & ck. manage inventory & orders.

# §C (Constraints)
- stack: go/chi, postgres/sqlx, next.js.
- env: windows dev, gcp cloud run prod.
- db: gcp cloud sql (iam auth !).
- security: blacklist mw, rate limit mw (120/min).

# §I (Interfaces)
- api: GET /api/products → paginated results
- api: GET /api/products/{id} → single product
- api: POST /api/admin/prices/refresh → sync prices (sync)
- api: POST /api/admin/currency/sync → sync rates
- api: POST /api/admin/settings → update global cfg
- store: setting table → kv pairs (e.g. blocked_ips)

# §V (Invariants)
- V1: ∀ api req → rate limit check (ratelimit.go)
- V2: ∀ api req → blacklist check (blacklist.go)
- V3: production db ! → cloudsqlconn dialer (db.go)
- V4: mtg prices → scryfall | cardkingdom
- V5: usd_to_cop_rate, eur_to_cop_rate ! in setting table
- V6: heavy tasks (sync, refresh, import) ! → worker pool
- V7: jobs ! → persist status & progress in db
- V8: admins exempt from global rate limits (ratelimit.go)

# §T (Tasks)
| id | status | task | cites |
|----|--------|------|-------|
| T1 | x | impl blacklist mw | V2 |
| T2 | x | add blocked_ips to settings | V5 |
| T3 | x | fix iam connector | V3 |
| T4 | x | impl job table & service | V7 |
| T5 | x | impl worker pool logic | V6 |
| T6 | x | refactor sync/refresh → jobs | V6,V7 |
| T7 | x | add job status api | V7 |
| T8 | x | update admin ui → poll jobs | V7 |
| T9 | x | exempt admins from rate limits | V8 |

# §B (Bugs)
| id | date | cause | fix |
|----|------|-------|-----|
| B1 | 2026-05-14 | iam token timeout | V3 |
