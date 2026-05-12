# Design Spec: Nightly Currency Refresh & Price Materialization

**Date**: 2026-05-12  
**Status**: Approved  
**Topic**: Automated exchange rate synchronization and materialized price indexing.

## 1. Objective
Automate the nightly update of USD->COP and EUR->COP exchange rates to ensure that materialized card prices (`price_cop`) used for sorting and filtering remain accurate without manual intervention.

## 2. Architecture

### 2.1 External Integration (`backend/external/currency.go`)
- **Source**: [ExchangeRate-API](https://www.exchangerate-api.com/) (v6).
- **Authentication**: `EXCHANGERATE_API_KEY` environment variable.
- **Functionality**: `FetchRates()` will return the latest conversion rates for COP relative to USD and EUR.

### 2.2 Service Layer (`backend/service/refresh_service.go`)
- **New Method**: `SyncCurrencyRates(ctx context.Context)`.
- **Workflow**:
    1. Call `external.FetchRates()`.
    2. Update `usd_to_cop_rate` and `eur_to_cop_rate` in the `settings` table via `SettingsService`.
    3. Trigger `store.RecalculateAllPrices(usd, eur, ck)` to update the `price_cop` column database-wide.

### 2.3 Data Layer (`backend/store/refresh_store.go`)
- **New Method**: `RecalculateAllPrices(ctx, usd, eur, ck float64)`.
- **Logic**: Performs a bulk `UPDATE product` using a `CASE` statement to recompute `price_cop` based on `price_source` and `price_reference`.

### 2.4 Scheduler (`backend/handlers/refresh.go`)
- **Update**: `StartMidnightScheduler` will execute `SyncCurrencyRates` as the first step of the midnight routine, ensuring card prices fetched immediately after use the newest rates.

## 3. Data Flow
1. **00:00**: Scheduler triggers.
2. **Sync**: API → Service → Settings DB.
3. **Materialize**: Service → RefreshStore → Product Table (`price_cop`).
4. **Price Update**: Card price fetch (Scryfall/CK) → Bulk Update (using new rates).

## 4. Error Handling
- If the API fails, the system logs the error and continues with existing rates to avoid blocking the price refresh.
- A "Last Synced" timestamp will be added to the settings to monitor health.

## 5. Performance
- The database-wide update (~20k+ rows) will be run in a single transaction.
- Indexing on `price_cop` ensures that sorting performance is unaffected by the update frequency.
