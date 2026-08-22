# grocr

A terminal UI for scanning barcodes into [Grocy](https://grocy.info/) inventory. Built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea).

This app is meant to serve as a low-friction way to enter grocery item information into your Grocy database.

## Screenshots

**Add product form** — after a barcode scan, fill in expiry, location, quantity, and price:

![Terminal form showing product name, expiry date, location, quantity, and price fields](docs/grocr-1.png)

**Product search** — press `/` to find products by name:

![Terminal search results listing products with stock quantity and location](docs/grocr-2.png)

## Compatibility

Tested against **Grocy 4.6.0**. Other recent versions will likely work, but 4.6.0 is the reference version used during development.

## What it does

Reads UPC barcodes (from a USB scanner or typed manually), looks up product info from Grocy and Open Food Facts, then adds or consumes stock. Products not yet in Grocy are created on the fly with sensible defaults for shelf life and location.

## Features

- **Add / Consume modes** — toggle with `m`
- **Barcode lookup** — checks Grocy first, then Open Food Facts for name, categories, and shelf life
- **Product search** — press `/` to find products by name (useful for items without barcodes, like produce)
- **Edit product name** — press `e` when viewing an existing product
- **Smart defaults** — shelf life estimated from product categories, location remembered per product
- **Scan log** — recent actions shown at the top of the screen
- **Custom keybinds** — every command key can be rebound in a small text file

## Installation

Requires Go 1.21 or later.

```
go install github.com/k-leveller/grocr@latest
```

Or build from source:

```
git clone https://github.com/k-leveller/grocr
cd grocr
go build -o grocr .
```

## Configuration

Set environment variables:

```
GROCY_URL=https://your-grocy-instance.example.com
GROCY_API_KEY=your-api-key
```

Or place a JSON file at `~/.config/grocr/config.json`:

```json
{
  "base_url": "https://your-grocy-instance.example.com",
  "api_key": "your-api-key"
}
```

Environment variables take precedence over the config file.

Keybinds are configured separately, in `~/.config/grocr/keybinds.conf` — see
[Custom keybinds](#custom-keybinds).

### Optional: display name userfield

If you have a custom Grocy userfield you want to use as a short display label for products (shown in the Expiring Soon panel and as an optional field when creating products), set `display_name_userfield` to the field's key:

```json
{
  "base_url": "https://your-grocy-instance.example.com",
  "api_key": "your-api-key",
  "display_name_userfield": "short_name"
}
```

When set, a **Display name** field appears during product creation, and the configured userfield value is shown in the Expiring Soon panel when available. Leave this unset (or omit it) to disable the feature.

## Usage

```
grocr              # normal mode
grocr --consume    # start in consume mode
grocr --test       # UI preview, no Grocy API calls
```

## Default Keybinds

Every key in the first table can be changed — see [Custom keybinds](#custom-keybinds).

| Key | Action |
|-----|--------|
| `q` | Quit / leave a panel |
| `m` | Cycle Add / Consume / Lookup mode |
| `/` | Search products by name |
| `n` | New product (no barcode) |
| `x` | Export stock to CSV |
| `t` | Today's meal plan |
| `P` | Meal plan (next 7 days) |
| `r` | Recipe list |
| `e` | Edit product name |
| `n` | Edit notes (lookup view) |
| `p` | Price history (lookup view) |
| `t` | Transfer stock between locations (lookup view) |
| `?` | Help overlay |
| `k` / `j` | Move up / down |
| `h` / `l` | Go back / drill in |
| `c` | Mark as consumed (expiring-soon panel) |
| `d` | Mark as spoiled (expiring-soon panel) |
| `r` | Reload the current panel |
| `c` | Create a new product (unknown barcode) |
| `l` | Link to an existing product (unknown barcode) |
| `y` | Confirm a yes/no prompt |

The keys below are fixed and cannot be rebound:

| Key | Action |
|-----|--------|
| `Tab` / `Shift-Tab` | Navigate form fields |
| `↑` `↓` `←` `→` | Navigate (always work alongside the bound keys) |
| `Enter` | Advance field / submit |
| `Esc` | Cancel current scan |
| `Ctrl+C` | Quit |

### Custom keybinds

Keybinds live in a plain text file at `~/.config/grocr/keybinds.conf`, written the
first time grocr starts. One binding per line:

```
# Lines starting with # or ; are ignored.
quit            = q
mode            = m
search          = /
meal_plan       = P      ; a trailing comment works too
price_history   = ctrl+p
```

Keys are case-sensitive, so `P` means Shift+P (`shift+p` is accepted and means
the same thing). Special keys can be spelled out: `up`, `down`, `left`, `right`,
`enter`, `esc`, `tab`, `space`, `backspace`, `pgup`, `pgdown`, `home`, `end`,
`delete`, `insert`, `f1`–`f12`, or `ctrl+<key>` / `alt+<key>`.

`Enter`, `Esc`, `Tab`, `Shift+Tab` and `Ctrl+C` are reserved by the UI and cannot
be bound.

The defaults also live in the binary, so anything the file leaves out — or gets
wrong — falls back to its default. Unknown action names, unparseable lines,
reserved keys and misspelled key names (`ctlr+q`) are all skipped, which means a
corrupted file never stops grocr from working. Delete the file to regenerate it
with the defaults.

Hint text and the `?` help overlay always show your bindings, not the defaults.

The bindable action names are:

`quit`, `mode`, `search`, `new`, `export`, `meal_plan_today`, `meal_plan`,
`recipes`, `edit_name`, `notes`, `price_history`, `transfer`, `help`, `up`,
`down`, `left`, `right`, `consume`, `spoil`, `refresh`, `create`, `link`, `yes`
