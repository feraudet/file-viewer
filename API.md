# API Documentation

File Viewer exposes a REST API for rendering and serving files.

## Base URL

```
http://localhost:4120
```

---

## Endpoints

### Render File

Renders a file with format-specific processing.

```
GET /{filepath}
```

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `filepath` | path | Absolute path to the file |

**Supported Formats:**

| Extension | Content-Type | Description |
|-----------|--------------|-------------|
| `.md`, `.markdown` | text/html | Markdown with TOC, syntax highlighting, math, diagrams |
| `.json` | text/html | Interactive tree view with search |
| `.yaml`, `.yml` | text/html | Syntax highlighted YAML |
| `.toml` | text/html | Syntax highlighted TOML |
| `.csv` | text/html | Interactive table with filtering |
| `.html`, `.htm` | text/html | Raw HTML passthrough |
| `.txt`, `.text` | text/html | Preformatted text with search |

**Example:**

```bash
curl http://localhost:4120/Users/me/docs/README.md
```

**Response:** Full HTML page with rendered content.

---

### Alternative File Access

Access files with path context for relative asset resolution.

```
GET /?path={directory}&filename={filename}
```

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `path` | query | Directory containing the file |
| `filename` | query | Name of the file |

**Example:**

```bash
curl "http://localhost:4120/?path=/Users/me/docs&filename=README.md"
```

---

### List Directory

Returns directory contents as JSON.

```
GET /files?dir={directory}
```

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `dir` | query | Absolute path to directory |

**Response:**

```json
{
  "files": [
    {
      "name": "README.md",
      "isDir": false,
      "path": "/Users/me/docs/README.md"
    },
    {
      "name": "images",
      "isDir": true,
      "path": "/Users/me/docs/images"
    }
  ]
}
```

**Notes:**
- Files larger than 5MB are excluded
- Binary files are filtered out
- Hidden files (starting with `.`) are included

---

### Get File Modification Time

Returns the last modification time of a file (used for live reload).

```
GET /mtime/{filepath}
```

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `filepath` | path | Absolute path to the file |

**Response:**

```
1704672000
```

Unix timestamp as plain text.

**Error Response:**

```
error
```

Returned if file doesn't exist or is inaccessible.

---

### Preview Content

Returns only the rendered content without the full HTML wrapper (used for link preview on hover).

```
GET /preview/{filepath}
```

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `filepath` | path | Absolute path to the file |

**Response:** Rendered HTML content fragment.

**Example:**

```bash
curl http://localhost:4120/preview/Users/me/docs/README.md
```

---

### Serve Asset

Serves static assets like images, PDFs, and other binary files.

```
GET /asset?path={filepath}
```

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `path` | query | Absolute path to the asset |

**Content-Type Detection:**

| Extension | Content-Type |
|-----------|--------------|
| `.png` | image/png |
| `.jpg`, `.jpeg` | image/jpeg |
| `.gif` | image/gif |
| `.svg` | image/svg+xml |
| `.webp` | image/webp |
| `.pdf` | application/pdf |
| `.ico` | image/x-icon |
| others | application/octet-stream |

**Example:**

```bash
curl "http://localhost:4120/asset?path=/Users/me/docs/logo.png" --output logo.png
```

---

### CDN Proxy

Proxies and caches CDN resources locally for offline access and performance.

```
GET /cdn/{host}/{path}
```

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `host` | path | CDN hostname (e.g., `cdn.jsdelivr.net`) |
| `path` | path | Resource path on the CDN |

**Example:**

```bash
curl http://localhost:4120/cdn/cdn.jsdelivr.net/npm/prismjs@1.29.0/prism.min.js
```

**Cache Location:**

```
~/.cache/file-viewer/cdn/{host}/{path}
```

**Notes:**
- First request fetches from CDN and caches locally
- Subsequent requests served from cache
- No cache expiration (manual deletion required)

---

## Error Handling

### File Not Found

When a file doesn't exist:

```html
<div class="error">
  <h2>File not found</h2>
  <p>/path/to/missing/file.md</p>
</div>
```

### Invalid JSON

When JSON parsing fails:

```html
<div class="error">Invalid JSON: unexpected end of JSON input</div>
```

---

## Live Reload

The frontend automatically polls `/mtime/{filepath}` every 2 seconds. When the modification time changes, the page reloads automatically.

To disable live reload, the polling can be stopped via browser console:

```javascript
clearInterval(window.liveReloadInterval);
```

---

## CORS

The server does not set CORS headers. All requests should originate from the same localhost origin.

---

## Limits

| Limit | Value |
|-------|-------|
| Max file size for rendering | 5 MB |
| Max recent files tracked | 15 |
| Max split panels | 4 |
| Live reload poll interval | 2 seconds |

---

## Client-Side Storage

The frontend uses `localStorage` for persistence:

| Key | Description |
|-----|-------------|
| `fileViewerTheme` | Selected theme name |
| `fileViewerFavorites` | Bookmarked files/folders |
| `fileViewerRecentFiles` | Recently viewed files |
| `fileViewerPanels` | Panel configuration |
| `fileViewerFavoritesCollapsed` | Favorites section state |
| `fileViewerRecentCollapsed` | Recent section state |
