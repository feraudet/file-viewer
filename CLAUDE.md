# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

File Viewer is a local HTTP server (Go) that renders Markdown, JSON, HTML, and text files with a modern web interface. It's designed for iTerm2's browser pane integration via `~/bin/open-in-iterm-browser` (Cmd+click on files).

## Build & Run

```bash
go build -o file-viewer
./file-viewer
```

Server runs on `http://localhost:4120`

## Endpoints

- `GET /{filepath}` - Render file (Markdown, JSON, HTML, text)
- `GET /?path=X&filename=Y` - Alternative file access with path context
- `GET /` - Shows Claude Code CHANGELOG (fetched from GitHub)
- `GET /asset?path=/absolute/path` - Serve static assets (images, SVG, PDF)
- `GET /mtime/{filepath}` - Return file modification time (for live reload)

## Architecture

Single-file Go application (`main.go`) with embedded HTML/CSS/JS template:

**Key functions:**
- `handler()` - Main HTTP router (line 225)
- `renderFile()` - Dispatch to format-specific renderers (line 336)
- `renderMarkdown()` - Custom Markdown parser with TOC generation (line 377)
- `renderJSON()` - Interactive tree view with expand/collapse (line 725)
- `buildHTML()` - Generate full HTML page with all dependencies (line 852)

**Markdown renderer features:**
- Inline processing: bold, italic, code, links, images, math, emoji
- Block elements: headers with anchors, lists (ordered/unordered/task), tables, blockquotes, code blocks
- Special blocks: Mermaid diagrams, KaTeX math (`$$...$$` and `$...$`)
- Auto-generated TOC for documents with 3+ headers

**Frontend dependencies (CDN):**
- Prism.js for syntax highlighting
- KaTeX for math rendering
- Mermaid for diagrams

## Supported Formats

| Extension | Handler |
|-----------|---------|
| .md, .markdown | Markdown with TOC, syntax highlighting, math, diagrams |
| .json | Interactive tree with search, expand/collapse |
| .yaml, .yml | YAML with syntax highlighting |
| .toml | TOML with syntax highlighting |
| .csv | Interactive table with filter |
| .html, .htm | Raw HTML passthrough |
| .txt, .text | Preformatted text with search |
