# Design Spec: Thematic Translation Expansion & Hardcoded Text Extraction

## 1. Goal
Complete a full-scale translation of the storefront (3,000+ keys) into 6 languages (DE, EN, ES, FR, IT, PT) using a **TCG/Cardboard (MTG-themed) approach**. Additionally, identify and migrate all hardcoded strings from the frontend into the translation system.

## 2. Approach: Balanced TCG Flavor
We will use a "Modern UI with Soul" approach:
- **Core actions** (Buy, Checkout, Search) remain clear but use thematic flavor in headers, loading states, and auxiliary buttons.
- **Tone**: Casual, respectful, not "cringe," with TCG-related jokes and wordplay.
- **Reference**: Official Magic: The Gathering (MTG) terminology in each respective language.

### Thematic Vocabulary Mapping
| Standard Term | TCG Thematic Term | Usage Context |
| :--- | :--- | :--- |
| **Search** | Scry / Search Library | Search bar, filters. |
| **Inventory** | The Library | Main catalog/grid. |
| **Loading** | Untapping / Shuffling | Initial page load or data fetching. |
| **Deleted/Trash** | The Graveyard / Discard | Cart removals, deleted orders. |
| **Restore** | Unearth / Reanimate | Re-adding items from cart history. |
| **Login** | Draw a Hand | Authentication pages. |
| **Empty State** | One with Nothing | Empty cart or no search results. |
| **Wishlist** | Sideboard | "Saved for later" items. |

## 3. Implementation Phases

### Phase 1: Hardcoded Text Extraction
- **Scan**: Search `frontend/src/**/*.tsx` for text literals that are not wrapped in the `t()` or `i18n` translation hooks.
- **Migrate**: Create new keys in `en.json` and replace hardcoded text with `t('new.key')`.
- **Validation**: Ensure no UI regressions in the frontend.

### Phase 2: Core Tone Refinement (EN/ES)
- Update existing `en.json` and `es.json` files to remove "hacker/futurism" themes.
- Implement the TCG theme for existing keys.

### Phase 3: Mass Expansion (DE, FR, IT, PT)
- **Source**: Use the refined `en.json` as the source of truth.
- **Sub-Agent Workflow**: Dispatch sub-agents per language to translate the 3,000+ keys.
- **Quality Control**: Validate JSON structure and ensure official MTG terms are used correctly.

## 4. Technical Details
- **Location**: `backend/seed/data/translations/*.json`
- **Frontend Hook**: `useTranslation()` from `react-i18next` (or similar library used in project).
- **Sub-Agents**: Will be given specific "flavor" prompts to maintain the casual-but-respectful TCG tone.

## 5. Success Criteria
- [ ] No hardcoded English/Spanish text remains in `frontend/src`.
- [ ] All 6 languages have valid `.json` files with 3,000+ keys.
- [ ] Tone is consistently "Cardboard Themed" across all locales.
- [ ] The app feels premium and respectful to the player base.
