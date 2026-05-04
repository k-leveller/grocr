# grocy-scanner

A terminal UI for scanning barcodes into [Grocy](https://grocy.info/) inventory. Built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## What it does

Reads UPC barcodes (from a USB scanner or typed manually), looks up product info from Grocy and Open Food Facts, then adds or consumes stock. Products not yet in Grocy are created on the fly with sensible defaults for shelf life and location.

## Features

- **Add / Consume modes** — toggle with `m`
- **Barcode lookup** — checks Grocy first, then Open Food Facts for name, categories, and shelf life
- **Product search** — press `/` to find products by name (useful for items without barcodes, like produce)
- **Edit product name** — press `e` when viewing an existing product
- **Smart defaults** — shelf life estimated from product categories, location remembered per product
- **Scan log** — recent actions shown at the top of the screen

## Usage

```
grocy-scanner              # normal mode
grocy-scanner --consume    # start in consume mode
grocy-scanner --test       # UI preview, no Grocy API calls
```

## Configuration

Set environment variables:

```
GROCY_URL=https://your-grocy-instance.example.com
GROCY_API_KEY=your-api-key
```

Or place a JSON file at `~/.config/grocy-scanner/config.json`:

```json
{
  "base_url": "https://your-grocy-instance.example.com",
  "api_key": "your-api-key"
}
```

## Keybindings

| Key | Action |
|-----|--------|
| `q` | Quit |
| `m` | Toggle Add/Consume mode |
| `/` | Search products by name |
| `e` | Edit product name |
| `?` | Help overlay |
| `Tab` / `Shift-Tab` | Navigate form fields |
| `Enter` | Advance field / submit |
| `Esc` | Cancel current scan |

## Building

```
go build -o grocy-scanner .
```
