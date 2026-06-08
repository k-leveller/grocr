# Feature Brainstorming

## Current Feature Set (as of 2026-06-08)

- Barcode scan → add or consume stock (with quantity, price, expiry, location)
- Three scan modes: add / consume / lookup (toggle with `m`)
- Product search by name (`/`)
- Manual new product entry (`Ctrl+N`)
- Edit product name (`e`)
- Expiring-soon side panel (7-day window)
- OpenFoodFacts integration (auto-fill name, shelf life)
- Shelf life estimation from product name/category
- Stock info display (amount on hand, last price)
- Add/consume log with syslog output

---

## Ideas

### Expiry & Waste

**Mark as spoiled from expiry panel**
The expiring-soon panel already lists items about to go bad. Add a keypress (e.g. `d`) to consume them as spoiled directly from the panel, which maps to Grocy's `spoiled=true` consume flag. High value, small lift.

**Configurable expiry-panel threshold**
Hard-coded at 7 days. Expose as a config-file setting so it can be widened to 14 or 30 days without a rebuild.

**Expiry urgency tiers**
Color-code expiring items: expired (red), 0–3 days (orange), 4–7 days (yellow), 8–14 days (dim). Makes the panel scannable at a glance.

---

### Shopping List

**Shopping list mode**
New scan mode (fourth `m` toggle) that looks up scanned barcodes against the Grocy shopping list and checks them off. Useful when unpacking groceries — scan each item, it removes it from the list and adds it to stock. A combined "buy and stock" flow for a solo household.

**Add to shopping list**
From the display/lookup view, press a key (e.g. `s`) to add the current product to the Grocy shopping list. Useful when you notice you're running low while scanning.

**Low-stock shopping-list auto-add**
After a consume action drops stock to zero, offer a prompt: "Add to shopping list? [y/N]". Maps to Grocy's existing shopping list API.

---

### Inventory Management

**Stock transfer between locations**
From the display view, allow moving N units from one location to another (fridge → freezer, pantry → fridge). Grocy supports `inventory` transactions with a `location_id`. Useful for meal-prep freezing flows.

**Mark as opened/in-use**
Grocy tracks "open" stock. Add a key to mark a unit as opened, which affects shelf life for things like yogurt or sauce.

**Stock-taking mode**
Systematic mode: cycle through all products one by one, display current expected count, let you enter the actual count. Posts a Grocy inventory correction. Good for monthly pantry checks.

---

### Recipes & Meal Planning

**Recipe list view**
New screen (accessible from idle, e.g. `r`) that fetches Grocy recipes and shows which ones are fulfillable with current stock. Read-only to start. Very useful for deciding what to cook.

**Recipe consume**
From the recipe view, press Enter to consume all ingredients for a recipe at once. Grocy has `/api/recipes/{id}/fulfillment` and consume endpoints.

**Today's meal plan**
Show a compact read-only view of the Grocy meal plan for today/this week. Lets you check the plan without leaving the terminal.

---

### Price & Spending

**Price history for a product**
From the lookup/display view, press `p` to see the last N price entries for that product. Grocy's `/api/stock/transactions` filtered by product gives purchase history.

**Per-unit price display alongside total**
Already shows last price. Show price/unit (e.g. $/oz) in the form hint when the product has a known quantity unit and size. Requires storing unit size metadata.

---

### Data Quality

**Spoiled count display**
In the product display/lookup view, show "X spoiled in last 30d" next to stock info. Helps identify products consistently going bad before use.

**Duplicate product warning**
When adding a new product manually, check for similar names (fuzzy match against `allProducts`) and warn before creating. Reduces junk data for a solo user who adds items infrequently.

**Reassign barcode**
When a scanned UPC is unknown, offer not just "create new" but also "link to existing product" (search picker). Useful when a brand repackages under a new UPC.

---

### UI & Navigation

**Product notes view/edit**
Grocy products have a `description` field. Show it in the lookup view and allow inline editing. Good for storing notes like "buy at Trader Joe's" or "use within 3 days of opening".

**Location filter in search**
Filter the `/` search results by location (fridge, pantry, freezer). Lets you quickly see what's in one place.

**Vim-style navigation in search**
`j`/`k` for up/down in the search list (already have arrow keys; add letter aliases for keyboard-only flow).

**History navigation in idle input**
Up/down in the idle input box cycles through recently scanned UPCs, like shell history. Useful when re-adding the same product.

---

### Integration & Output

**Desktop notification on startup for expiring items**
On launch, if any items are expired or expiring within 24h, send a system notification (via `notify-send` or macOS `osascript`). The app already queries expiring-soon at startup.

**Export stock snapshot**
Command (e.g. `ctrl+e`) to write current stock to a markdown or CSV file in `~/grocy-export-YYYYMMDD.csv`. Useful for a quick overview without opening the Grocy web UI.

**Receipt parsing**
Experimental: paste a store receipt (as text) and batch-add items. Would require a simple line-by-line parser + confirmation flow. High effort, high reward for weekly grocery runs.

---

## Rough Priority Order

1. Mark as spoiled from expiry panel — small, immediately useful
2. Add to shopping list (from display view) — one API call, very handy
3. Low-stock shopping-list prompt after consume — small UX win
4. Configurable expiry threshold — trivial config change
5. Recipe list view (read-only) — moderate effort, high value for meal planning
6. Expiry urgency color tiers — cosmetic but makes panel much more useful
7. Shopping list scan mode — medium effort, useful for unpacking groceries
8. Stock transfer between locations — moderate effort
9. Reassign barcode to existing product — small lift on the add flow
10. Recipe consume — depends on recipe list view being done first
