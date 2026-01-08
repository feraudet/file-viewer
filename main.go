package main

import (
	"bytes"
	"compress/zlib"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Emoji map for common emojis
var emojiMap = map[string]string{
	"smile":           "😊",
	"grinning":        "😀",
	"laughing":        "😆",
	"joy":             "😂",
	"wink":            "😉",
	"heart":           "❤️",
	"heart_eyes":      "😍",
	"thumbsup":        "👍",
	"+1":              "👍",
	"thumbsdown":      "👎",
	"-1":              "👎",
	"clap":            "👏",
	"fire":            "🔥",
	"rocket":          "🚀",
	"star":            "⭐",
	"warning":         "⚠️",
	"check":           "✅",
	"x":               "❌",
	"question":        "❓",
	"exclamation":     "❗",
	"bulb":            "💡",
	"memo":            "📝",
	"book":            "📖",
	"link":            "🔗",
	"eyes":            "👀",
	"thinking":        "🤔",
	"tada":            "🎉",
	"party":           "🥳",
	"wave":            "👋",
	"muscle":          "💪",
	"coffee":          "☕",
	"bug":             "🐛",
	"wrench":          "🔧",
	"hammer":          "🔨",
	"gear":            "⚙️",
	"lock":            "🔒",
	"key":             "🔑",
	"sparkles":        "✨",
	"zap":             "⚡",
	"boom":            "💥",
	"100":             "💯",
	"ok":              "👌",
	"pray":            "🙏",
	"raised_hands":    "🙌",
	"point_right":     "👉",
	"point_left":      "👈",
	"point_up":        "👆",
	"point_down":      "👇",
	"arrow_right":     "➡️",
	"arrow_left":      "⬅️",
	"arrow_up":        "⬆️",
	"arrow_down":      "⬇️",
	"white_check_mark": "✅",
	"heavy_check_mark": "✔️",
	"green_circle":    "🟢",
	"red_circle":      "🔴",
	"blue_circle":     "🔵",
	"yellow_circle":   "🟡",
	"package":         "📦",
	"folder":          "📁",
	"file":            "📄",
	"computer":        "💻",
	"keyboard":        "⌨️",
	"mouse":           "🖱️",
	"robot":           "🤖",
	"alien":           "👽",
	"skull":           "💀",
	"ghost":           "👻",
	"poop":            "💩",
	"sunglasses":      "😎",
	"nerd":            "🤓",
	"cry":             "😢",
	"sob":             "😭",
	"angry":           "😠",
	"rage":            "🤬",
	"confused":        "😕",
	"worried":         "😟",
	"relieved":        "😌",
	"sleeping":        "😴",
	"zzz":             "💤",
	"clock":           "🕐",
	"hourglass":       "⏳",
	"stopwatch":       "⏱️",
	"calendar":        "📅",
	"pin":             "📌",
	"paperclip":       "📎",
	"scissors":        "✂️",
	"pencil":          "✏️",
	"mag":             "🔍",
	"bell":            "🔔",
	"speaker":         "🔊",
	"mute":            "🔇",
	"email":           "📧",
	"phone":           "📱",
	"battery":         "🔋",
	"signal":          "📶",
	"wifi":            "📡",
	"cloud":           "☁️",
	"sun":             "☀️",
	"moon":            "🌙",
	"rainbow":         "🌈",
	"umbrella":        "☔",
	"snow":            "❄️",
	"tree":            "🌳",
	"flower":          "🌸",
	"apple":           "🍎",
	"pizza":           "🍕",
	"beer":            "🍺",
	"wine":            "🍷",
	"cake":            "🎂",
	"gift":            "🎁",
	"trophy":          "🏆",
	"medal":           "🏅",
	"crown":           "👑",
	"gem":             "💎",
	"money":           "💰",
	"dollar":          "💵",
	"chart":           "📊",
	"graph":           "📈",
	"construction":    "🚧",
	"car":             "🚗",
	"airplane":        "✈️",
	"ship":            "🚢",
	"house":           "🏠",
	"office":          "🏢",
	"hospital":        "🏥",
	"school":          "🏫",
}

const PORT = 4120

var httpClient = &http.Client{Timeout: 10 * time.Second}

// Generate a URL-friendly slug from header text
func slugify(text string) string {
	// Remove HTML tags
	tagRe := regexp.MustCompile(`<[^>]+>`)
	text = tagRe.ReplaceAllString(text, "")
	// Decode HTML entities
	text = html.UnescapeString(text)
	// Convert to lowercase
	text = strings.ToLower(text)
	// Replace spaces and non-alphanumeric with hyphens
	re := regexp.MustCompile(`[^a-z0-9]+`)
	text = re.ReplaceAllString(text, "-")
	// Trim hyphens
	text = strings.Trim(text, "-")
	// Generate hash for uniqueness if needed
	if text == "" {
		hash := md5.Sum([]byte(text))
		text = "heading-" + hex.EncodeToString(hash[:4])
	}
	return text
}

// Replace :emoji: with actual emoji characters
func replaceEmojis(text string) string {
	emojiRe := regexp.MustCompile(`:([a-z0-9_+-]+):`)
	return emojiRe.ReplaceAllStringFunc(text, func(m string) string {
		name := m[1 : len(m)-1]
		if emoji, ok := emojiMap[name]; ok {
			return emoji
		}
		return m // Keep original if not found
	})
}

// Header struct for TOC generation
type Header struct {
	Level  int
	Text   string
	Anchor string
}

// FileEntry represents a file or directory for the sidebar
type FileEntry struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	IsDir    bool   `json:"isDir"`
	Size     int64  `json:"size"`
	Ext      string `json:"ext"`
	Viewable bool   `json:"viewable"`
}

// Maximum file size for viewing (5MB)
const MaxViewableSize = 5 * 1024 * 1024

// Binary/non-viewable extensions
var binaryExtensions = map[string]bool{
	"": true, ".exe": true, ".bin": true, ".so": true, ".dylib": true, ".dll": true,
	".o": true, ".a": true, ".lib": true, ".obj": true,
	".zip": true, ".tar": true, ".gz": true, ".bz2": true, ".xz": true, ".7z": true, ".rar": true,
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".bmp": true, ".ico": true, ".webp": true, ".tiff": true,
	".mp3": true, ".wav": true, ".flac": true, ".aac": true, ".ogg": true, ".m4a": true,
	".mp4": true, ".avi": true, ".mkv": true, ".mov": true, ".wmv": true, ".flv": true, ".webm": true,
	".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true, ".ppt": true, ".pptx": true,
	".ttf": true, ".otf": true, ".woff": true, ".woff2": true, ".eot": true,
	".class": true, ".pyc": true, ".pyo": true, ".wasm": true,
	".db": true, ".sqlite": true, ".sqlite3": true,
}

func fetchURL(url string) (string, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// getCacheDir returns the cache directory for CDN resources
func getCacheDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "/tmp/file-viewer-cache"
	}
	return filepath.Join(homeDir, ".cache", "file-viewer", "cdn")
}

// encodePlantUML encodes PlantUML content for the PlantUML server API
func encodePlantUML(content string) string {
	// Compress using zlib/deflate
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	w.Write([]byte(content))
	w.Close()

	// Encode using PlantUML's custom base64-like encoding
	compressed := buf.Bytes()
	return encodePlantUMLBytes(compressed)
}

// PlantUML uses a custom base64-like encoding
func encodePlantUMLBytes(data []byte) string {
	const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-_"
	var result strings.Builder

	for i := 0; i < len(data); i += 3 {
		b1 := data[i]
		b2 := byte(0)
		b3 := byte(0)
		if i+1 < len(data) {
			b2 = data[i+1]
		}
		if i+2 < len(data) {
			b3 = data[i+2]
		}

		c1 := b1 >> 2
		c2 := ((b1 & 0x3) << 4) | (b2 >> 4)
		c3 := ((b2 & 0xF) << 2) | (b3 >> 6)
		c4 := b3 & 0x3F

		result.WriteByte(alphabet[c1])
		result.WriteByte(alphabet[c2])
		if i+1 < len(data) {
			result.WriteByte(alphabet[c3])
		}
		if i+2 < len(data) {
			result.WriteByte(alphabet[c4])
		}
	}

	return result.String()
}

// listDirectory returns a sorted list of files and directories
func listDirectory(dirPath string) ([]FileEntry, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	var files []FileEntry
	for _, entry := range entries {
		// Skip hidden files (starting with .)
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		fullPath := filepath.Join(dirPath, entry.Name())
		ext := ""
		if !entry.IsDir() {
			ext = strings.ToLower(filepath.Ext(entry.Name()))
		}

		// Determine if file is viewable (not binary and not too large)
		viewable := true
		if !entry.IsDir() {
			if binaryExtensions[ext] {
				viewable = false
			} else if info.Size() > MaxViewableSize {
				viewable = false
			}
		}

		files = append(files, FileEntry{
			Name:     entry.Name(),
			Path:     fullPath,
			IsDir:    entry.IsDir(),
			Size:     info.Size(),
			Ext:      ext,
			Viewable: viewable,
		})
	}

	// Sort: directories first, then alphabetically
	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir != files[j].IsDir {
			return files[i].IsDir
		}
		return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
	})

	return files, nil
}

func main() {
	http.HandleFunc("/", handler)

	fmt.Printf("File Viewer running on http://localhost:%d\n", PORT)
	fmt.Printf("Usage: http://localhost:%d/path/to/file\n", PORT)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", PORT), nil); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	// Log request
	fmt.Printf("[%s] %s %s\n", time.Now().Format("15:04:05"), r.Method, r.URL.String())

	urlPath := filepath.Clean(r.URL.Path)

	// Asset endpoint - serve static files (images, etc.)
	if urlPath == "/asset" {
		assetPath := r.URL.Query().Get("path")
		if assetPath == "" {
			http.Error(w, "Missing path parameter", http.StatusBadRequest)
			return
		}
		// Clean and validate path
		assetPath = filepath.Clean(assetPath)

		// Check file exists
		info, err := os.Stat(assetPath)
		if err != nil || info.IsDir() {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		// Determine content type
		ext := strings.ToLower(filepath.Ext(assetPath))
		contentTypes := map[string]string{
			".svg":  "image/svg+xml",
			".png":  "image/png",
			".jpg":  "image/jpeg",
			".jpeg": "image/jpeg",
			".gif":  "image/gif",
			".webp": "image/webp",
			".ico":  "image/x-icon",
			".pdf":  "application/pdf",
		}
		if ct, ok := contentTypes[ext]; ok {
			w.Header().Set("Content-Type", ct)
		}

		http.ServeFile(w, r, assetPath)
		return
	}

	// CDN Cache endpoint - proxy and cache CDN resources locally
	if strings.HasPrefix(urlPath, "/cdn/") {
		cdnPath := urlPath[5:] // Remove "/cdn/" prefix
		// Reconstruct CDN URL from path
		// Format: /cdn/cdnjs.cloudflare.com/... or /cdn/cdn.jsdelivr.net/...
		parts := strings.SplitN(cdnPath, "/", 2)
		if len(parts) < 2 {
			http.Error(w, "Invalid CDN path", http.StatusBadRequest)
			return
		}
		cdnHost := parts[0]
		cdnResource := parts[1]
		cdnURL := "https://" + cdnHost + "/" + cdnResource

		// Get cache directory
		cacheDir := getCacheDir()
		cachePath := filepath.Join(cacheDir, cdnPath)

		// Check if cached
		if data, err := os.ReadFile(cachePath); err == nil {
			// Serve from cache
			ext := strings.ToLower(filepath.Ext(cachePath))
			contentTypes := map[string]string{
				".css": "text/css; charset=utf-8",
				".js":  "application/javascript; charset=utf-8",
			}
			if ct, ok := contentTypes[ext]; ok {
				w.Header().Set("Content-Type", ct)
			}
			w.Header().Set("X-Cache", "HIT")
			w.Write(data)
			return
		}

		// Fetch from CDN
		resp, err := http.Get(cdnURL)
		if err != nil {
			http.Error(w, "Failed to fetch from CDN", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			http.Error(w, fmt.Sprintf("CDN returned %d", resp.StatusCode), resp.StatusCode)
			return
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "Failed to read CDN response", http.StatusBadGateway)
			return
		}

		// Cache the response
		if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err == nil {
			os.WriteFile(cachePath, data, 0644)
		}

		// Serve response
		ext := strings.ToLower(filepath.Ext(cachePath))
		contentTypes := map[string]string{
			".css": "text/css; charset=utf-8",
			".js":  "application/javascript; charset=utf-8",
		}
		if ct, ok := contentTypes[ext]; ok {
			w.Header().Set("Content-Type", ct)
		}
		w.Header().Set("X-Cache", "MISS")
		w.Write(data)
		return
	}

	// Files API endpoint - list directory contents for sidebar
	if urlPath == "/files" {
		dirPath := r.URL.Query().Get("dir")
		if dirPath == "" {
			dirPath = "/"
		}
		dirPath = filepath.Clean(dirPath)

		// Validate it's a directory
		info, err := os.Stat(dirPath)
		if err != nil || !info.IsDir() {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid directory"})
			return
		}

		files, err := listDirectory(dirPath)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"dir":    dirPath,
			"parent": filepath.Dir(dirPath),
			"files":  files,
		})
		return
	}

	// Live reload endpoint
	if strings.HasPrefix(urlPath, "/mtime/") {
		filePath := urlPath[7:]
		info, err := os.Stat(filePath)
		if err != nil {
			w.Write([]byte("0"))
			return
		}
		w.Write([]byte(fmt.Sprintf("%d", info.ModTime().UnixNano())))
		return
	}

	// Preview endpoint - return rendered content only (for link preview)
	if strings.HasPrefix(urlPath, "/preview/") {
		filePath := urlPath[9:]
		content, _ := renderFile(filePath)
		// Truncate to first ~500 chars of text content for preview
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(content))
		return
	}

	// Check for path and filename query parameters
	queryPath := r.URL.Query().Get("path")
	queryFilename := r.URL.Query().Get("filename")

	var filePath string
	if queryFilename != "" {
		if filepath.IsAbs(queryFilename) {
			// Filename is absolute, use it directly
			filePath = queryFilename
		} else if queryPath != "" {
			// Filename is relative, combine with path
			filePath = filepath.Join(queryPath, queryFilename)
		} else {
			// Relative filename without path context
			filePath = queryFilename
		}
	} else if queryPath != "" {
		// Just path parameter
		filePath = queryPath
	} else if urlPath != "/" {
		// Use URL path directly
		filePath = urlPath
	}

	// Root path without query params - show Claude Code CHANGELOG
	if filePath == "" {
		content, err := fetchURL("https://raw.githubusercontent.com/anthropics/claude-code/refs/heads/main/CHANGELOG.md")
		if err != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(fmt.Sprintf(`<html><body><h1>Error</h1><p>%s</p></body></html>`, err.Error())))
			return
		}
		htmlPage := buildHTML("Claude Code changelog", "Claude Code", renderMarkdown(content, ""), "markdown")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(htmlPage))
		return
	}

	// Render file
	content, contentClass := renderFile(filePath)
	htmlPage := buildHTML(filepath.Base(filePath), filePath, content, contentClass)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(htmlPage))
}

func usagePage() string {
	return `<!DOCTYPE html>
<html><head><title>File Viewer</title></head>
<body style="font-family: sans-serif; max-width: 600px; margin: 50px auto; padding: 20px;">
<h1>File Viewer</h1>
<p>Usage: <code>http://localhost:4120/path/to/file</code></p>
<p>Supported formats: .md, .json, .txt, .html</p>
</body></html>`
}

func renderFile(filePath string) (string, string) {
	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Sprintf(`<p style="color: red;">File not found: %s</p>`, html.EscapeString(filePath)), ""
	}
	if info.IsDir() {
		return fmt.Sprintf(`<p style="color: red;">Not a file: %s</p>`, html.EscapeString(filePath)), ""
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Sprintf(`<p style="color: red;">Error reading file: %s</p>`, html.EscapeString(err.Error())), ""
	}

	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".md", ".markdown":
		return renderMarkdown(string(content), filepath.Dir(filePath)), "markdown"
	case ".json":
		return renderJSON(string(content)), "json"
	case ".yaml", ".yml":
		return renderYAML(string(content)), "yaml"
	case ".toml":
		return renderTOML(string(content)), "toml"
	case ".csv":
		return renderCSV(string(content)), "csv"
	case ".html", ".htm":
		return string(content), "html"
	case ".txt", ".text", "":
		return renderText(string(content)), "text"
	default:
		return fmt.Sprintf(`<div class="text">%s</div>`, html.EscapeString(string(content))), "text"
	}
}

func renderText(content string) string {
	toolbar := `<div class="search-toolbar">
    <input type="text" id="content-search" placeholder="Rechercher..." oninput="textSearch(this.value, 'searchable-content')" />
    <button onclick="prevMatch()">◀</button>
    <button onclick="nextMatch()">▶</button>
    <span id="search-count" class="search-count"></span>
</div>`
	initScript := `<script>document.addEventListener("DOMContentLoaded", function() { initSearch("searchable-content"); });</script>`
	return fmt.Sprintf(`%s<div id="searchable-content" class="text">%s</div>%s`, toolbar, html.EscapeString(content), initScript)
}

func renderMarkdown(content string, baseDir string) string {
	lines := strings.Split(content, "\n")
	var result strings.Builder
	var headers []Header
	inCodeBlock := false
	codeLang := ""
	var codeLines []string
	inUL := false
	inOL := false
	inBlockquote := false
	inTable := false
	tableHeaderDone := false
	inMathBlock := false
	var mathLines []string
	codeBlockID := 0
	footnotes := make(map[string]string)
	footnoteOrder := []string{}

	processInline := func(text string) string {
		// Inline math $...$ (protect first, before other processing)
		mathRe := regexp.MustCompile(`\$([^$\n]+)\$`)
		text = mathRe.ReplaceAllString(text, `<span class="math-inline">$$$1$$</span>`)

		// Code (protect early)
		codeRe := regexp.MustCompile("`([^`]+)`")
		text = codeRe.ReplaceAllStringFunc(text, func(m string) string {
			inner := m[1 : len(m)-1]
			return "<code>" + html.EscapeString(inner) + "</code>"
		})

		// Images (with lightbox support) - resolve relative paths
		imgRe := regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)
		text = imgRe.ReplaceAllStringFunc(text, func(m string) string {
			matches := imgRe.FindStringSubmatch(m)
			if len(matches) != 3 {
				return m
			}
			alt := matches[1]
			src := matches[2]
			// Resolve relative paths
			if baseDir != "" && !strings.HasPrefix(src, "http://") && !strings.HasPrefix(src, "https://") && !strings.HasPrefix(src, "/") {
				src = "/asset?path=" + filepath.Join(baseDir, src)
			} else if baseDir != "" && strings.HasPrefix(src, "/") && !strings.HasPrefix(src, "/asset") {
				// Absolute path on filesystem
				src = "/asset?path=" + src
			}
			return fmt.Sprintf(`<img src="%s" alt="%s" class="lightbox-img" onclick="openLightbox(this.src, this.alt)" style="max-width:100%%; cursor: zoom-in;">`, src, html.EscapeString(alt))
		})

		// Links
		linkRe := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
		text = linkRe.ReplaceAllString(text, `<a href="$2">$1</a>`)

		// Highlight ==text==
		highlightRe := regexp.MustCompile(`==(.+?)==`)
		text = highlightRe.ReplaceAllString(text, `<mark>$1</mark>`)

		// Bold
		boldRe := regexp.MustCompile(`\*\*(.+?)\*\*`)
		text = boldRe.ReplaceAllString(text, `<strong>$1</strong>`)
		boldRe2 := regexp.MustCompile(`__(.+?)__`)
		text = boldRe2.ReplaceAllString(text, `<strong>$1</strong>`)

		// Italic
		italicRe := regexp.MustCompile(`\*(.+?)\*`)
		text = italicRe.ReplaceAllString(text, `<em>$1</em>`)

		// Strikethrough
		strikeRe := regexp.MustCompile(`~~(.+?)~~`)
		text = strikeRe.ReplaceAllString(text, `<del>$1</del>`)

		// Emoji :name:
		text = replaceEmojis(text)

		// Footnote references [^id]
		footnoteRefRe := regexp.MustCompile(`\[\^([^\]]+)\]`)
		text = footnoteRefRe.ReplaceAllStringFunc(text, func(m string) string {
			id := m[2 : len(m)-1] // Extract id from [^id]
			return fmt.Sprintf(`<sup class="footnote-ref"><a href="#fn-%s" id="fnref-%s">[%s]</a></sup>`, id, id, id)
		})

		return text
	}

	closeLists := func() {
		if inUL {
			result.WriteString("</ul>\n")
			inUL = false
		}
		if inOL {
			result.WriteString("</ol>\n")
			inOL = false
		}
	}

	isHTMLBlock := func(line string) bool {
		trimmed := strings.TrimSpace(line)
		htmlBlockRe := regexp.MustCompile(`(?i)^<(/?)(div|p|table|tr|td|th|thead|tbody|ul|ol|li|h[1-6]|pre|blockquote|hr|br|img|a|span|section|article|header|footer|nav|aside|figure|figcaption|details|summary|style|script)(\s|>|/>)`)
		return htmlBlockRe.MatchString(trimmed)
	}

	for i, line := range lines {
		_ = i

		// Math blocks $$...$$
		if strings.TrimSpace(line) == "$$" {
			if inMathBlock {
				mathContent := strings.Join(mathLines, "\n")
				result.WriteString(fmt.Sprintf("<div class=\"math-block\">$$%s$$</div>\n", html.EscapeString(mathContent)))
				inMathBlock = false
				mathLines = nil
			} else {
				closeLists()
				inMathBlock = true
			}
			continue
		}

		if inMathBlock {
			mathLines = append(mathLines, line)
			continue
		}

		// Code blocks
		if strings.HasPrefix(line, "```") {
			if inCodeBlock {
				codeContent := strings.Join(codeLines, "\n")
				if codeLang == "mermaid" {
					// Mermaid diagram
					result.WriteString(fmt.Sprintf("<div class=\"mermaid\">%s</div>\n", html.EscapeString(codeContent)))
				} else if codeLang == "plantuml" || codeLang == "puml" {
					// PlantUML diagram - render via PlantUML server
					encoded := encodePlantUML(codeContent)
					result.WriteString(fmt.Sprintf("<div class=\"plantuml\"><img src=\"https://www.plantuml.com/plantuml/svg/%s\" alt=\"PlantUML diagram\" loading=\"lazy\"></div>\n", encoded))
				} else {
					// Regular code block with copy button and line numbers
					escapedCode := html.EscapeString(codeContent)
					result.WriteString(escapedCode)
					result.WriteString("</code></pre>")
					result.WriteString(fmt.Sprintf(`<button class="copy-btn" onclick="copyCode('code-%d')">📋 Copy</button>`, codeBlockID))
					result.WriteString("</div>\n")
					codeBlockID++
				}
				inCodeBlock = false
				codeLang = ""
				codeLines = nil
			} else {
				closeLists()
				codeLang = strings.TrimSpace(line[3:])
				if codeLang != "mermaid" && codeLang != "plantuml" && codeLang != "puml" {
					langClass := ""
					if codeLang != "" {
						langClass = fmt.Sprintf(` class="language-%s"`, codeLang)
					}
					result.WriteString(fmt.Sprintf("<div class=\"code-block\"><pre class=\"line-numbers\"><code id=\"code-%d\"%s>", codeBlockID, langClass))
				}
				inCodeBlock = true
			}
			continue
		}

		if inCodeBlock {
			codeLines = append(codeLines, line)
			continue
		}

		// Horizontal rule
		hrRe := regexp.MustCompile(`^(-{3,}|\*{3,}|_{3,})$`)
		if hrRe.MatchString(strings.TrimSpace(line)) {
			closeLists()
			result.WriteString("<hr>\n")
			continue
		}

		// Footnote definitions [^id]: content
		footnoteDefRe := regexp.MustCompile(`^\[\^([^\]]+)\]:\s*(.+)$`)
		if m := footnoteDefRe.FindStringSubmatch(line); m != nil {
			id := m[1]
			content := m[2]
			if _, exists := footnotes[id]; !exists {
				footnoteOrder = append(footnoteOrder, id)
			}
			footnotes[id] = content
			continue
		}

		// Headers with anchors
		headerRe := regexp.MustCompile(`^(#{1,6})\s+(.+)$`)
		if m := headerRe.FindStringSubmatch(line); m != nil {
			closeLists()
			level := len(m[1])
			rawText := m[2]
			content := processInline(rawText)
			anchor := slugify(rawText)

			// Track header for TOC
			headers = append(headers, Header{Level: level, Text: rawText, Anchor: anchor})

			// Header with anchor link
			result.WriteString(fmt.Sprintf("<h%d id=\"%s\" class=\"header-anchor\">%s <a href=\"#%s\" class=\"anchor-link\">#</a></h%d>\n", level, anchor, content, anchor, level))
			continue
		}

		// Tables
		if strings.Contains(line, "|") && strings.HasPrefix(strings.TrimSpace(line), "|") {
			cells := strings.Split(strings.Trim(strings.TrimSpace(line), "|"), "|")
			for i := range cells {
				cells[i] = strings.TrimSpace(cells[i])
			}

			// Separator row
			isSeparator := true
			sepRe := regexp.MustCompile(`^:?-+:?$`)
			for _, c := range cells {
				if !sepRe.MatchString(c) {
					isSeparator = false
					break
				}
			}
			if isSeparator {
				tableHeaderDone = true
				continue
			}

			if !inTable {
				closeLists()
				result.WriteString("<table>\n")
				inTable = true
				tableHeaderDone = false
			}

			tag := "th"
			if tableHeaderDone {
				tag = "td"
			}
			result.WriteString("<tr>")
			for _, c := range cells {
				result.WriteString(fmt.Sprintf("<%s>%s</%s>", tag, processInline(c), tag))
			}
			result.WriteString("</tr>\n")
			continue
		} else if inTable {
			result.WriteString("</table>\n")
			inTable = false
			tableHeaderDone = false
		}

		// Blockquotes
		if strings.HasPrefix(line, "> ") {
			if !inBlockquote {
				closeLists()
				result.WriteString("<blockquote>")
				inBlockquote = true
			}
			result.WriteString(processInline(line[2:]))
			continue
		} else if inBlockquote && strings.TrimSpace(line) == "" {
			result.WriteString("</blockquote>\n")
			inBlockquote = false
		}

		// Task lists (checkboxes)
		taskRe := regexp.MustCompile(`^(\s*)[-*+]\s+\[([ xX])\]\s+(.+)$`)
		if m := taskRe.FindStringSubmatch(line); m != nil {
			if inOL {
				result.WriteString("</ol>\n")
				inOL = false
			}
			if !inUL {
				result.WriteString("<ul class=\"task-list\">\n")
				inUL = true
			}
			checked := strings.ToLower(m[2]) == "x"
			checkbox := `<span class="checkbox unchecked">☐</span>`
			if checked {
				checkbox = `<span class="checkbox checked">✓</span>`
			}
			result.WriteString(fmt.Sprintf("<li class=\"task-item\">%s %s</li>\n", checkbox, processInline(m[3])))
			continue
		}

		// Unordered lists
		ulRe := regexp.MustCompile(`^(\s*)[-*+]\s+(.+)$`)
		if m := ulRe.FindStringSubmatch(line); m != nil {
			if inOL {
				result.WriteString("</ol>\n")
				inOL = false
			}
			if !inUL {
				result.WriteString("<ul>\n")
				inUL = true
			}
			result.WriteString(fmt.Sprintf("<li>%s</li>\n", processInline(m[2])))
			continue
		}

		// Ordered lists
		olRe := regexp.MustCompile(`^(\s*)\d+\.\s+(.+)$`)
		if m := olRe.FindStringSubmatch(line); m != nil {
			if inUL {
				result.WriteString("</ul>\n")
				inUL = false
			}
			if !inOL {
				result.WriteString("<ol>\n")
				inOL = true
			}
			result.WriteString(fmt.Sprintf("<li>%s</li>\n", processInline(m[2])))
			continue
		}

		// Empty line
		if strings.TrimSpace(line) == "" {
			closeLists()
			continue
		}

		// HTML block
		if isHTMLBlock(line) {
			closeLists()
			result.WriteString(line + "\n")
			continue
		}

		// Regular paragraph
		closeLists()
		result.WriteString(fmt.Sprintf("<p>%s</p>\n", processInline(line)))
	}

	// Close open elements
	closeLists()
	if inCodeBlock {
		codeContent := strings.Join(codeLines, "\n")
		if codeLang == "mermaid" {
			result.WriteString(fmt.Sprintf("<div class=\"mermaid\">%s</div>\n", html.EscapeString(codeContent)))
		} else if codeLang == "plantuml" || codeLang == "puml" {
			encoded := encodePlantUML(codeContent)
			result.WriteString(fmt.Sprintf("<div class=\"plantuml\"><img src=\"https://www.plantuml.com/plantuml/svg/%s\" alt=\"PlantUML diagram\" loading=\"lazy\"></div>\n", encoded))
		} else {
			result.WriteString(html.EscapeString(codeContent))
			result.WriteString("</code></pre></div>\n")
		}
	}
	if inMathBlock {
		mathContent := strings.Join(mathLines, "\n")
		result.WriteString(fmt.Sprintf("<div class=\"math-block\">$$%s$$</div>\n", html.EscapeString(mathContent)))
	}
	if inBlockquote {
		result.WriteString("</blockquote>\n")
	}
	if inTable {
		result.WriteString("</table>\n")
	}

	// Render footnotes section if any exist
	if len(footnoteOrder) > 0 {
		result.WriteString("<hr class=\"footnotes-sep\">\n")
		result.WriteString("<section class=\"footnotes\">\n")
		result.WriteString("<h4>Notes</h4>\n")
		result.WriteString("<ol class=\"footnotes-list\">\n")
		for _, id := range footnoteOrder {
			content := footnotes[id]
			result.WriteString(fmt.Sprintf("<li id=\"fn-%s\" class=\"footnote-item\">%s <a href=\"#fnref-%s\" class=\"footnote-backref\">↩</a></li>\n", id, processInline(content), id))
		}
		result.WriteString("</ol>\n")
		result.WriteString("</section>\n")
	}

	// Generate TOC if there are enough headers
	var toc strings.Builder
	if len(headers) >= 3 {
		// Collapse TOC if more than 10 entries
		openAttr := " open"
		if len(headers) > 10 {
			openAttr = ""
		}
		toc.WriteString(fmt.Sprintf("<details class=\"toc\"%s>\n", openAttr))
		toc.WriteString(fmt.Sprintf("<summary>📑 Table of Contents (%d)</summary>\n", len(headers)))
		toc.WriteString("<nav class=\"toc-nav\">\n")
		for _, h := range headers {
			indent := (h.Level - 1) * 16
			toc.WriteString(fmt.Sprintf("<a href=\"#%s\" style=\"padding-left: %dpx;\">%s</a>\n", h.Anchor, indent, replaceEmojis(h.Text)))
		}
		toc.WriteString("</nav>\n")
		toc.WriteString("</details>\n")
	}

	return toc.String() + result.String()
}

func renderJSON(content string) string {
	var parsed interface{}
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return fmt.Sprintf(`<span class="error">Invalid JSON</span><pre>%s</pre>`, html.EscapeString(content))
	}

	treeHTML := renderJSONTree(parsed)

	return fmt.Sprintf(`<div class="json-toolbar">
    <input type="text" id="json-search" placeholder="Rechercher..." oninput="searchJson(this.value)" />
    <button onclick="expandAll()">Expand All</button>
    <button onclick="collapseAll()">Collapse All</button>
</div>
<div class="json-tree"><ul><li>%s</li></ul></div>
<script>
function expandAll() {
    document.querySelectorAll('.json-tree li.json-collapsed').forEach(function(li) {
        li.classList.remove('json-collapsed');
        var toggle = li.querySelector('.json-toggle');
        if (toggle) toggle.textContent = '▼';
    });
}
function collapseAll() {
    document.querySelectorAll('.json-tree li').forEach(function(li) {
        if (li.querySelector('ul')) {
            li.classList.add('json-collapsed');
            var toggle = li.querySelector('.json-toggle');
            if (toggle) toggle.textContent = '▶';
        }
    });
}
function searchJson(query) {
    document.querySelectorAll('.json-highlight').forEach(function(el) {
        var text = el.textContent;
        el.replaceWith(document.createTextNode(text));
    });
    if (!query) return;
    var spans = document.querySelectorAll('.json-key, .json-string, .json-number');
    spans.forEach(function(el) {
        if (el.textContent.toLowerCase().indexOf(query.toLowerCase()) !== -1) {
            var parent = el.closest('li');
            while (parent) {
                parent.classList.remove('json-collapsed');
                var toggle = parent.querySelector('.json-toggle');
                if (toggle) toggle.textContent = '▼';
                parent = parent.parentElement ? parent.parentElement.closest('li') : null;
            }
            var re = new RegExp('(' + query.replace(/[.*+?^${}()|[\]\\]/g, '\\$&') + ')', 'gi');
            var parts = el.textContent.split(re);
            el.textContent = '';
            parts.forEach(function(part) {
                if (part.toLowerCase() === query.toLowerCase()) {
                    var mark = document.createElement('span');
                    mark.className = 'json-highlight';
                    mark.textContent = part;
                    el.appendChild(mark);
                } else {
                    el.appendChild(document.createTextNode(part));
                }
            });
        }
    });
}
</script>`, treeHTML)
}

func renderJSONTree(obj interface{}) string {
	var result strings.Builder

	switch v := obj.(type) {
	case map[string]interface{}:
		if len(v) == 0 {
			return `<span class="json-bracket">{}</span>`
		}
		result.WriteString(`<span class="json-toggle" onclick="this.parentElement.classList.toggle('json-collapsed');this.textContent=this.textContent==='▼'?'▶':'▼'">▼</span>`)
		result.WriteString(`<span class="json-bracket">{</span>`)
		result.WriteString(fmt.Sprintf(`<span class="json-preview">%d items...</span>`, len(v)))
		result.WriteString("<ul>")
		i := 0
		for key, value := range v {
			comma := ","
			if i == len(v)-1 {
				comma = ""
			}
			result.WriteString(fmt.Sprintf(`<li><span class="json-key">"%s"</span>: %s%s</li>`, html.EscapeString(key), renderJSONTree(value), comma))
			i++
		}
		result.WriteString("</ul>")
		result.WriteString(`<span class="json-bracket">}</span>`)

	case []interface{}:
		if len(v) == 0 {
			return `<span class="json-bracket">[]</span>`
		}
		result.WriteString(`<span class="json-toggle" onclick="this.parentElement.classList.toggle('json-collapsed');this.textContent=this.textContent==='▼'?'▶':'▼'">▼</span>`)
		result.WriteString(`<span class="json-bracket">[</span>`)
		result.WriteString(fmt.Sprintf(`<span class="json-preview">%d items...</span>`, len(v)))
		result.WriteString("<ul>")
		for i, value := range v {
			comma := ","
			if i == len(v)-1 {
				comma = ""
			}
			result.WriteString(fmt.Sprintf(`<li>%s%s</li>`, renderJSONTree(value), comma))
		}
		result.WriteString("</ul>")
		result.WriteString(`<span class="json-bracket">]</span>`)

	case string:
		return fmt.Sprintf(`<span class="json-string">"%s"</span>`, html.EscapeString(v))

	case float64:
		if v == float64(int64(v)) {
			return fmt.Sprintf(`<span class="json-number">%d</span>`, int64(v))
		}
		return fmt.Sprintf(`<span class="json-number">%v</span>`, v)

	case bool:
		return fmt.Sprintf(`<span class="json-boolean">%v</span>`, v)

	case nil:
		return `<span class="json-null">null</span>`
	}

	return result.String()
}

func renderYAML(content string) string {
	// Display YAML with syntax highlighting using Prism
	toolbar := `<div class="yaml-toolbar">
    <button onclick="copyYAML()" title="Copy YAML">📋 Copy</button>
</div>`
	escaped := html.EscapeString(content)
	return fmt.Sprintf(`%s<pre class="line-numbers"><code class="language-yaml" id="yaml-content">%s</code></pre>
<script>
function copyYAML() {
    const content = document.getElementById('yaml-content').textContent;
    navigator.clipboard.writeText(content);
}
</script>`, toolbar, escaped)
}

func renderTOML(content string) string {
	// Display TOML with syntax highlighting using Prism
	toolbar := `<div class="toml-toolbar">
    <button onclick="copyTOML()" title="Copy TOML">📋 Copy</button>
</div>`
	escaped := html.EscapeString(content)
	return fmt.Sprintf(`%s<pre class="line-numbers"><code class="language-toml" id="toml-content">%s</code></pre>
<script>
function copyTOML() {
    const content = document.getElementById('toml-content').textContent;
    navigator.clipboard.writeText(content);
}
</script>`, toolbar, escaped)
}

func renderCSV(content string) string {
	lines := strings.Split(strings.TrimSpace(content), "\n")
	if len(lines) == 0 {
		return `<p>Empty CSV file</p>`
	}

	var result strings.Builder
	result.WriteString(`<div class="csv-toolbar">
    <input type="text" id="csv-search" placeholder="Filter rows..." oninput="filterCSV(this.value)" />
    <span id="csv-count"></span>
</div>
<div class="csv-container">
<table class="csv-table" id="csv-table">
<thead><tr>`)

	// Parse header
	headers := parseCSVLine(lines[0])
	for _, h := range headers {
		result.WriteString(fmt.Sprintf(`<th>%s</th>`, html.EscapeString(h)))
	}
	result.WriteString(`</tr></thead><tbody>`)

	// Parse rows
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "" {
			continue
		}
		cells := parseCSVLine(lines[i])
		result.WriteString(`<tr>`)
		for j := 0; j < len(headers); j++ {
			val := ""
			if j < len(cells) {
				val = cells[j]
			}
			result.WriteString(fmt.Sprintf(`<td>%s</td>`, html.EscapeString(val)))
		}
		result.WriteString(`</tr>`)
	}

	result.WriteString(`</tbody></table></div>
<script>
function filterCSV(query) {
    const table = document.getElementById('csv-table');
    const rows = table.querySelectorAll('tbody tr');
    let count = 0;
    const q = query.toLowerCase();
    rows.forEach(row => {
        const text = row.textContent.toLowerCase();
        const match = !q || text.includes(q);
        row.style.display = match ? '' : 'none';
        if (match) count++;
    });
    document.getElementById('csv-count').textContent = q ? count + ' / ' + rows.length + ' rows' : rows.length + ' rows';
}
document.addEventListener('DOMContentLoaded', function() { filterCSV(''); });
</script>`)
	return result.String()
}

func parseCSVLine(line string) []string {
	var result []string
	var current strings.Builder
	inQuotes := false

	for i := 0; i < len(line); i++ {
		c := line[i]
		if c == '"' {
			if inQuotes && i+1 < len(line) && line[i+1] == '"' {
				current.WriteByte('"')
				i++
			} else {
				inQuotes = !inQuotes
			}
		} else if c == ',' && !inQuotes {
			result = append(result, current.String())
			current.Reset()
		} else {
			current.WriteByte(c)
		}
	}
	result = append(result, current.String())
	return result
}

func buildHTML(title, filePath, content, contentClass string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <!-- Prism CSS -->
    <link rel="stylesheet" href="/cdn/cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/themes/prism-okaidia.min.css">
    <link rel="stylesheet" href="/cdn/cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/plugins/line-numbers/prism-line-numbers.min.css">
    <!-- KaTeX CSS -->
    <link rel="stylesheet" href="/cdn/cdn.jsdelivr.net/npm/katex@0.16.9/dist/katex.min.css">
    <style>
        :root {
            --bg-primary: #fafafa;
            --bg-secondary: white;
            --bg-code: #f1f5f9;
            --text-primary: #333;
            --text-secondary: #64748b;
            --border-color: #e2e8f0;
            --header-bg: #2d3748;
            --link-color: #3b82f6;
            --accent-color: #3b82f6;
        }
        .dark-mode, .theme-dark {
            --bg-primary: #1a1a2e;
            --bg-secondary: #16213e;
            --bg-code: #0f3460;
            --text-primary: #e4e4e7;
            --text-secondary: #a1a1aa;
            --border-color: #3f3f46;
            --header-bg: #0f3460;
            --link-color: #60a5fa;
            --accent-color: #60a5fa;
        }
        .theme-sepia {
            --bg-primary: #f4ecd8;
            --bg-secondary: #faf8f1;
            --bg-code: #ede4d0;
            --text-primary: #5c4b37;
            --text-secondary: #8b7355;
            --border-color: #d4c4a8;
            --header-bg: #6b5344;
            --link-color: #8b5a2b;
            --accent-color: #a0522d;
        }
        .theme-nord {
            --bg-primary: #2e3440;
            --bg-secondary: #3b4252;
            --bg-code: #434c5e;
            --text-primary: #eceff4;
            --text-secondary: #d8dee9;
            --border-color: #4c566a;
            --header-bg: #3b4252;
            --link-color: #88c0d0;
            --accent-color: #81a1c1;
        }
        .theme-solarized-light {
            --bg-primary: #fdf6e3;
            --bg-secondary: #eee8d5;
            --bg-code: #eee8d5;
            --text-primary: #657b83;
            --text-secondary: #93a1a1;
            --border-color: #93a1a1;
            --header-bg: #073642;
            --link-color: #268bd2;
            --accent-color: #2aa198;
        }
        .theme-solarized-dark {
            --bg-primary: #002b36;
            --bg-secondary: #073642;
            --bg-code: #073642;
            --text-primary: #839496;
            --text-secondary: #586e75;
            --border-color: #586e75;
            --header-bg: #073642;
            --link-color: #268bd2;
            --accent-color: #2aa198;
        }
        * { box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
            line-height: 1.6;
            margin: 0;
            padding: 0;
            background: var(--bg-primary);
            color: var(--text-primary);
            transition: background 0.3s, color 0.3s;
        }
        /* App container - flexbox layout */
        .app-container {
            display: flex;
            min-height: 100vh;
        }
        /* Sidebar */
        .sidebar {
            width: 280px;
            min-width: 280px;
            background: var(--bg-secondary);
            border-right: 1px solid var(--border-color);
            display: flex;
            flex-direction: column;
            transition: margin-left 0.3s ease;
            position: fixed;
            top: 0;
            left: 0;
            height: 100vh;
            overflow: hidden;
            z-index: 100;
        }
        .sidebar-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 12px 16px;
            border-bottom: 1px solid var(--border-color);
            background: var(--header-bg);
            color: white;
            font-size: 14px;
            font-weight: 600;
        }
        .sidebar-close {
            background: transparent;
            border: none;
            color: white;
            font-size: 20px;
            cursor: pointer;
            padding: 0 4px;
            line-height: 1;
        }
        .sidebar-close:hover { opacity: 0.7; }
        .sidebar-content {
            flex: 1;
            overflow-y: auto;
        }
        /* Breadcrumb in sidebar */
        .sidebar-breadcrumb {
            padding: 8px 12px;
            font-size: 12px;
            color: var(--text-secondary);
            border-bottom: 1px solid var(--border-color);
            background: var(--bg-code);
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
        }
        .sidebar-breadcrumb a {
            color: var(--link-color);
            text-decoration: none;
        }
        .sidebar-breadcrumb a:hover { text-decoration: underline; }
        /* File tree */
        .file-tree { font-size: 13px; }
        .file-tree ul {
            list-style: none;
            margin: 0;
            padding: 0;
        }
        .file-tree .tree-item {
            display: flex;
            align-items: center;
            padding: 6px 12px;
            cursor: pointer;
            color: var(--text-primary);
            text-decoration: none;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
        }
        .file-tree .tree-item:hover { background: var(--bg-code); }
        .file-tree .tree-item.active {
            background: var(--accent-color);
            color: white;
        }
        .file-tree .tree-item.disabled {
            color: var(--text-secondary);
            opacity: 0.5;
            cursor: not-allowed;
        }
        .file-tree .tree-item.disabled:hover {
            background: transparent;
        }
        .file-tree .tree-icon {
            margin-right: 8px;
            font-size: 14px;
            flex-shrink: 0;
        }
        .file-tree .tree-name {
            overflow: hidden;
            text-overflow: ellipsis;
        }
        /* Star button for favorites */
        .star-btn {
            opacity: 0;
            background: none;
            border: none;
            cursor: pointer;
            padding: 0 4px;
            font-size: 14px;
            color: var(--text-secondary);
            transition: opacity 0.15s, color 0.15s;
            flex-shrink: 0;
            margin-left: auto;
        }
        .tree-item:hover .star-btn { opacity: 1; }
        .star-btn:hover { color: #f59e0b; }
        .star-btn.favorited { opacity: 1; color: #f59e0b; }
        /* Favorites Section */
        .favorites-section {
            border-bottom: 1px solid var(--border-color);
        }
        .favorites-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 8px 12px;
            font-size: 12px;
            font-weight: 600;
            color: var(--text-secondary);
            cursor: pointer;
            background: var(--bg-code);
            user-select: none;
        }
        .favorites-header:hover { background: var(--bg-primary); }
        .favorites-header .toggle-icon {
            transition: transform 0.2s;
        }
        .favorites-section.collapsed .favorites-header .toggle-icon {
            transform: rotate(-90deg);
        }
        .favorites-list {
            max-height: 250px;
            overflow-y: auto;
            font-size: 13px;
        }
        .favorites-list ul { list-style: none; padding: 0; margin: 0; }
        .favorites-list li { margin: 0; }
        .favorites-section.collapsed .favorites-list { display: none; }
        .favorites-empty {
            padding: 12px;
            font-size: 12px;
            color: var(--text-secondary);
            text-align: center;
            font-style: italic;
        }
        .favorites-group {
            border-bottom: 1px solid var(--border-color);
        }
        .favorites-group:last-child { border-bottom: none; }
        .favorites-group-header {
            display: flex;
            align-items: center;
            padding: 6px 12px;
            font-size: 11px;
            color: var(--text-secondary);
            background: var(--bg-secondary);
            cursor: pointer;
            user-select: none;
        }
        .favorites-group-header:hover { background: var(--bg-code); }
        .favorites-group-path {
            flex: 1;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
            direction: rtl;
            text-align: left;
        }
        .favorites-group-path span {
            direction: ltr;
            unicode-bidi: bidi-override;
        }
        .favorites-group-count {
            margin-left: 8px;
            font-size: 10px;
            background: var(--border-color);
            padding: 1px 5px;
            border-radius: 8px;
        }
        .favorites-group.collapsed .favorites-group-items { display: none; }
        .favorites-group-items.file-tree .tree-item {
            padding-left: 24px;
        }
        .favorites-group-items.file-tree ul {
            list-style: none;
            padding: 0;
            margin: 0;
        }
        /* Recent Files Section */
        .recent-section {
            border-bottom: 1px solid var(--border-color);
        }
        .recent-header {
            display: flex;
            align-items: center;
            padding: 8px 12px;
            font-size: 12px;
            font-weight: 600;
            color: var(--text-secondary);
            cursor: pointer;
            background: var(--bg-code);
            user-select: none;
            gap: 8px;
        }
        .recent-header:hover { background: var(--bg-primary); }
        .recent-header span:first-child { flex: 1; }
        .recent-clear-btn {
            background: none;
            border: none;
            color: var(--text-secondary);
            cursor: pointer;
            font-size: 10px;
            padding: 2px 4px;
            opacity: 0.5;
        }
        .recent-clear-btn:hover { opacity: 1; color: #ef4444; }
        .recent-list {
            max-height: 200px;
            overflow-y: auto;
            font-size: 13px;
        }
        .recent-list ul { list-style: none; padding: 0; margin: 0; }
        .recent-list li { margin: 0; }
        .recent-section.collapsed .recent-list { display: none; }
        .recent-empty {
            padding: 12px;
            font-size: 12px;
            color: var(--text-secondary);
            text-align: center;
            font-style: italic;
        }
        .recent-remove-btn {
            opacity: 0;
            background: none;
            border: none;
            color: var(--text-secondary);
            cursor: pointer;
            font-size: 10px;
            padding: 2px 4px;
            margin-left: auto;
        }
        .tree-item:hover .recent-remove-btn { opacity: 0.7; }
        .recent-remove-btn:hover { color: #ef4444; opacity: 1; }
        /* Navigation Panels */
        .panels-container {
            display: flex;
            flex-direction: column;
            flex: 1;
            overflow: hidden;
        }
        .nav-panel {
            display: flex;
            flex-direction: column;
            flex: 1;
            min-height: 120px;
            border-bottom: 2px solid var(--border-color);
            overflow: hidden;
        }
        .nav-panel:last-child { border-bottom: none; }
        .panel-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 4px 12px;
            background: var(--bg-code);
            border-bottom: 1px solid var(--border-color);
            font-size: 11px;
            color: var(--text-secondary);
        }
        .panel-header-title { font-weight: 600; }
        .panel-controls { display: flex; gap: 4px; }
        .panel-btn {
            background: none;
            border: none;
            cursor: pointer;
            padding: 2px 6px;
            font-size: 14px;
            color: var(--text-secondary);
            border-radius: 3px;
        }
        .panel-btn:hover {
            background: var(--border-color);
            color: var(--text-primary);
        }
        .panel-btn.add-panel { color: #22c55e; }
        .panel-btn.close-panel { color: #ef4444; }
        .panel-btn:disabled {
            opacity: 0.3;
            cursor: not-allowed;
        }
        .panel-btn:disabled:hover { background: none; }
        .panel-content {
            flex: 1;
            overflow-y: auto;
        }
        /* Panel resize handles */
        .panel-resize-handle {
            height: 6px;
            background: var(--border-color);
            cursor: ns-resize;
            flex-shrink: 0;
            position: relative;
            transition: background 0.2s;
        }
        .panel-resize-handle:hover,
        .panel-resize-handle.resizing {
            background: var(--link-color);
        }
        .panel-resize-handle::after {
            content: '';
            position: absolute;
            left: 50%%;
            top: 50%%;
            transform: translate(-50%%, -50%%);
            width: 30px;
            height: 2px;
            background: var(--text-secondary);
            border-radius: 1px;
            opacity: 0;
            transition: opacity 0.2s;
        }
        .panel-resize-handle:hover::after,
        .panel-resize-handle.resizing::after {
            opacity: 1;
        }
        .nav-panel.resizing {
            user-select: none;
        }
        body.panel-resizing {
            cursor: ns-resize !important;
            user-select: none !important;
        }
        body.panel-resizing * {
            cursor: ns-resize !important;
        }
        /* Main content area */
        .main-content {
            flex: 1;
            min-width: 0;
            display: flex;
            flex-direction: column;
            max-width: 100%%;
            margin-left: 280px;
            transition: margin-left 0.3s ease;
        }
        .sidebar-hidden .main-content {
            margin-left: 0;
        }
        .sidebar-hidden .sidebar {
            margin-left: -280px;
        }
        .main-content .content {
            flex: 1;
            padding: 20px;
            overflow-y: auto;
            max-width: 900px;
        }
        .header {
            background: var(--header-bg);
            color: white;
            padding: 10px 20px;
            font-family: monospace;
            font-size: 14px;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .header-left {
            display: flex;
            align-items: center;
            gap: 12px;
            overflow: hidden;
        }
        .header-left span {
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
        }
        .sidebar-toggle {
            background: transparent;
            border: 1px solid rgba(255,255,255,0.3);
            color: white;
            padding: 4px 8px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 16px;
            flex-shrink: 0;
        }
        .sidebar-toggle:hover { background: rgba(255,255,255,0.1); }
        .header-controls {
            display: flex;
            gap: 10px;
        }
        .print-btn {
            background: transparent;
            border: 1px solid rgba(255,255,255,0.3);
            color: white;
            padding: 4px 8px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 16px;
        }
        .print-btn:hover { background: rgba(255,255,255,0.1); }
        .theme-selector {
            background: rgba(255,255,255,0.1);
            border: 1px solid rgba(255,255,255,0.3);
            color: white;
            padding: 4px 8px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 13px;
        }
        .theme-selector option {
            background: var(--bg-secondary);
            color: var(--text-primary);
        }
        .content {
            background: var(--bg-secondary);
            padding: 20px;
            border: 1px solid var(--border-color);
            border-top: none;
            border-radius: 0 0 8px 8px;
        }
        /* TOC */
        .toc {
            background: var(--bg-code);
            border: 1px solid var(--border-color);
            border-radius: 8px;
            padding: 12px 16px;
            margin-bottom: 20px;
        }
        .toc summary {
            font-weight: 600;
            cursor: pointer;
            padding: 4px 0;
        }
        .toc-nav {
            display: flex;
            flex-direction: column;
            gap: 4px;
            margin-top: 8px;
        }
        .toc-nav a {
            color: var(--link-color);
            text-decoration: none;
            font-size: 14px;
            display: block;
        }
        .toc-nav a:hover { text-decoration: underline; }
        /* Headers with anchors */
        .header-anchor { position: relative; }
        .anchor-link {
            color: var(--text-secondary);
            text-decoration: none;
            opacity: 0;
            margin-left: 8px;
            transition: opacity 0.2s;
        }
        .header-anchor:hover .anchor-link { opacity: 1; }
        .anchor-link:hover { color: var(--link-color); }
        /* Markdown styles */
        .markdown h1, .markdown h2, .markdown h3 { color: var(--text-primary); margin-top: 1.5em; }
        .markdown h1 { border-bottom: 2px solid var(--border-color); padding-bottom: 0.3em; }
        .markdown code {
            background: var(--bg-code);
            padding: 2px 6px;
            border-radius: 4px;
            font-size: 0.9em;
        }
        .markdown pre {
            border-radius: 8px;
            overflow-x: auto;
            margin: 0;
        }
        .markdown pre code {
            background: none;
            padding: 0;
            font-size: 14px;
        }
        /* Code block with copy button */
        .code-block {
            position: relative;
            margin: 1em 0;
        }
        .copy-btn {
            position: absolute;
            top: 8px;
            right: 8px;
            background: rgba(255,255,255,0.1);
            border: 1px solid rgba(255,255,255,0.2);
            color: #ccc;
            padding: 4px 8px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 12px;
            opacity: 0;
            transition: opacity 0.2s;
        }
        .code-block:hover .copy-btn { opacity: 1; }
        .copy-btn:hover { background: rgba(255,255,255,0.2); }
        .copy-btn.copied { background: #22c55e; color: white; }
        /* Line numbers */
        pre.line-numbers {
            padding-left: 3.8em;
            counter-reset: linenumber;
        }
        pre.line-numbers > code {
            position: relative;
            white-space: inherit;
        }
        .line-numbers .line-numbers-rows {
            position: absolute;
            pointer-events: none;
            top: 0;
            font-size: 100%%;
            left: -3.8em;
            width: 3em;
            letter-spacing: -1px;
            border-right: 1px solid #999;
            user-select: none;
        }
        .line-numbers-rows > span {
            display: block;
            counter-increment: linenumber;
        }
        .line-numbers-rows > span:before {
            content: counter(linenumber);
            color: #999;
            display: block;
            padding-right: 0.8em;
            text-align: right;
        }
        /* Highlight */
        mark {
            background: #fef08a;
            padding: 1px 4px;
            border-radius: 2px;
        }
        .dark-mode mark {
            background: #854d0e;
            color: #fef9c3;
        }
        .markdown blockquote {
            border-left: 4px solid var(--accent-color);
            margin: 0;
            padding-left: 16px;
            color: var(--text-secondary);
        }
        .markdown a { color: var(--link-color); }
        .markdown .task-list { list-style: none; padding-left: 0; }
        .markdown .task-item { display: flex; align-items: flex-start; gap: 8px; margin: 4px 0; }
        .markdown .checkbox { font-size: 1.2em; line-height: 1; }
        .markdown .checkbox.checked { color: #22c55e; }
        .markdown .checkbox.unchecked { color: #9ca3af; }
        .markdown table { border-collapse: collapse; width: 100%%; }
        .markdown th, .markdown td {
            border: 1px solid var(--border-color);
            padding: 8px 12px;
            text-align: left;
        }
        .markdown th { background: var(--bg-code); }
        /* Math */
        .math-block {
            overflow-x: auto;
            padding: 16px;
            text-align: center;
            margin: 1em 0;
        }
        .math-inline { display: inline; }
        /* Footnotes */
        .footnotes-sep {
            margin-top: 2em;
            border-color: var(--border-color);
        }
        .footnotes {
            font-size: 0.9em;
            color: var(--text-secondary);
        }
        .footnotes h4 {
            margin: 0.5em 0;
            font-size: 1em;
            color: var(--text-primary);
        }
        .footnotes-list {
            padding-left: 1.5em;
            margin: 0.5em 0;
        }
        .footnote-item {
            margin: 0.5em 0;
            line-height: 1.5;
        }
        .footnote-ref a {
            color: var(--link-color);
            text-decoration: none;
            font-weight: 500;
        }
        .footnote-ref a:hover {
            text-decoration: underline;
        }
        .footnote-backref {
            color: var(--link-color);
            text-decoration: none;
            margin-left: 4px;
        }
        .footnote-backref:hover {
            text-decoration: underline;
        }
        /* Mermaid */
        .mermaid {
            background: white;
            padding: 16px;
            border-radius: 8px;
            margin: 1em 0;
            text-align: center;
        }
        .dark-mode .mermaid { background: #f8fafc; }
        /* PlantUML */
        .plantuml {
            background: white;
            padding: 16px;
            border-radius: 8px;
            margin: 1em 0;
            text-align: center;
        }
        .plantuml img {
            max-width: 100%%;
            height: auto;
        }
        .dark-mode .plantuml { background: #f8fafc; }
        /* Lightbox */
        .lightbox {
            display: none;
            position: fixed;
            top: 0;
            left: 0;
            width: 100%%;
            height: 100%%;
            background: rgba(0,0,0,0.9);
            z-index: 1000;
            justify-content: center;
            align-items: center;
            flex-direction: column;
        }
        .lightbox.active { display: flex; }
        .lightbox img {
            max-width: 90%%;
            max-height: 80%%;
            object-fit: contain;
        }
        .lightbox-caption {
            color: white;
            margin-top: 16px;
            font-size: 14px;
        }
        .lightbox-close {
            position: absolute;
            top: 20px;
            right: 30px;
            color: white;
            font-size: 40px;
            cursor: pointer;
        }
        .json-toolbar, .search-toolbar {
            display: flex;
            gap: 10px;
            margin-bottom: 15px;
            align-items: center;
        }
        .json-toolbar input, .search-toolbar input {
            flex: 1;
            padding: 8px 12px;
            border: 1px solid var(--border-color);
            border-radius: 6px;
            font-size: 14px;
            background: var(--bg-secondary);
            color: var(--text-primary);
        }
        .json-toolbar button, .search-toolbar button {
            padding: 8px 16px;
            border: none;
            border-radius: 6px;
            background: var(--accent-color);
            color: white;
            cursor: pointer;
            font-size: 14px;
        }
        .json-toolbar button:hover, .search-toolbar button:hover { opacity: 0.9; }
        .search-toolbar .search-count { color: var(--text-secondary); font-size: 14px; }
        /* YAML/TOML toolbars */
        .yaml-toolbar, .toml-toolbar {
            display: flex;
            gap: 10px;
            margin-bottom: 15px;
            align-items: center;
        }
        .yaml-toolbar button, .toml-toolbar button {
            padding: 8px 16px;
            border: none;
            border-radius: 6px;
            background: var(--accent-color);
            color: white;
            cursor: pointer;
            font-size: 14px;
        }
        .yaml-toolbar button:hover, .toml-toolbar button:hover { opacity: 0.9; }
        /* CSV table styles */
        .csv-toolbar {
            display: flex;
            gap: 10px;
            margin-bottom: 15px;
            align-items: center;
        }
        .csv-toolbar input {
            flex: 1;
            max-width: 300px;
            padding: 8px 12px;
            border: 1px solid var(--border-color);
            border-radius: 6px;
            font-size: 14px;
            background: var(--bg-secondary);
            color: var(--text-primary);
        }
        .csv-toolbar span { color: var(--text-secondary); font-size: 14px; }
        .csv-container { overflow-x: auto; }
        .csv-table {
            width: 100%%;
            border-collapse: collapse;
            font-size: 14px;
            font-family: 'SF Mono', Monaco, 'Courier New', monospace;
        }
        .csv-table th, .csv-table td {
            padding: 10px 12px;
            border: 1px solid var(--border-color);
            text-align: left;
        }
        .csv-table th {
            background: var(--bg-code);
            font-weight: 600;
            position: sticky;
            top: 0;
        }
        .csv-table tbody tr:hover {
            background: var(--bg-code);
        }
        .csv-table tbody tr:nth-child(even) {
            background: var(--bg-secondary);
        }
        .json-tree {
            font-family: 'SF Mono', Monaco, 'Courier New', monospace;
            font-size: 14px;
            line-height: 1.5;
        }
        .json-tree ul { list-style: none; padding-left: 20px; margin: 0; }
        .json-tree > ul { padding-left: 0; }
        .json-tree li { position: relative; }
        .json-toggle {
            cursor: pointer;
            user-select: none;
            display: inline-block;
            width: 16px;
            color: var(--text-secondary);
        }
        .json-toggle:hover { color: var(--text-primary); }
        .json-key { color: #0550ae; font-weight: 500; }
        .dark-mode .json-key { color: #7dd3fc; }
        .json-string { color: #0a3069; }
        .dark-mode .json-string { color: #86efac; }
        .json-number { color: #0550ae; }
        .dark-mode .json-number { color: #fbbf24; }
        .json-boolean { color: #cf222e; }
        .dark-mode .json-boolean { color: #f87171; }
        .json-null { color: #6e7781; font-style: italic; }
        .json-bracket { color: var(--text-secondary); }
        .json-collapsed > ul { display: none; }
        .json-collapsed > .json-preview { display: inline; }
        .json-preview { display: none; color: var(--text-secondary); font-style: italic; }
        .json-highlight, .search-highlight { background: #fef08a; border-radius: 2px; }
        .dark-mode .json-highlight, .dark-mode .search-highlight { background: #854d0e; color: #fef9c3; }
        .search-current { background: #f97316; color: white; }
        .text {
            font-family: 'SF Mono', Monaco, 'Courier New', monospace;
            white-space: pre-wrap;
            font-size: 14px;
        }
        /* Link Preview */
        .link-preview {
            display: none;
            position: fixed;
            z-index: 1000;
            background: var(--bg-secondary);
            border: 1px solid var(--border-color);
            border-radius: 8px;
            box-shadow: 0 4px 20px rgba(0,0,0,0.2);
            max-width: 400px;
            max-height: 300px;
            overflow: hidden;
        }
        .link-preview.visible { display: block; }
        .link-preview-header {
            padding: 8px 12px;
            font-size: 12px;
            font-weight: 600;
            background: var(--bg-code);
            border-bottom: 1px solid var(--border-color);
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
        }
        .link-preview-content {
            padding: 12px;
            font-size: 13px;
            max-height: 240px;
            overflow-y: auto;
            line-height: 1.5;
        }
        .link-preview-content h1,
        .link-preview-content h2,
        .link-preview-content h3 {
            font-size: 14px;
            margin: 0 0 8px 0;
        }
        .link-preview-content p {
            margin: 0 0 8px 0;
        }
        .link-preview-content pre {
            font-size: 11px;
            padding: 6px;
            margin: 8px 0;
        }
        .link-preview-loading {
            color: var(--text-secondary);
            font-style: italic;
        }
        .link-preview-error {
            color: #ef4444;
        }
        /* Print styles */
        @media print {
            body { background: white; }
            .sidebar, .header, .lightbox, .toc, .copy-btn, .search-toolbar,
            .json-toolbar, .yaml-toolbar, .toml-toolbar, .csv-toolbar { display: none !important; }
            .app-container { display: block; }
            .main-content { margin: 0; padding: 0; max-width: none; }
            .content {
                border: none;
                padding: 0;
                background: white;
                color: black;
            }
            .markdown h1, .markdown h2, .markdown h3,
            .markdown h4, .markdown h5, .markdown h6 { color: black; }
            .markdown a { color: #2563eb; }
            .markdown pre, .markdown code {
                background: #f3f4f6;
                border: 1px solid #e5e7eb;
            }
            .markdown blockquote {
                border-left-color: #9ca3af;
                background: #f9fafb;
            }
            .markdown table { border-color: #d1d5db; }
            .markdown th { background: #f3f4f6; }
            .anchor-link { display: none; }
            .footnotes { border-top: 1px solid #d1d5db; }
            @page {
                margin: 2cm;
                size: A4;
            }
        }
    </style>
</head>
<body>
    <div class="app-container">
        <aside class="sidebar" id="sidebar">
            <div class="sidebar-header">
                <span>Files <small style="opacity: 0.6; font-size: 10px;">v1.9.0</small></span>
                <button class="sidebar-close" onclick="toggleSidebar()" title="Hide sidebar">&times;</button>
            </div>
            <div class="sidebar-content" id="sidebar-content">
                <!-- Populated by JavaScript -->
            </div>
        </aside>
        <main class="main-content">
            <div class="header">
                <div class="header-left">
                    <button class="sidebar-toggle" onclick="toggleSidebar()" title="Toggle sidebar">☰</button>
                    <span>%s</span>
                </div>
                <div class="header-controls">
                    <button class="print-btn" onclick="printDocument()" title="Print / Export PDF">🖨️</button>
                    <select class="theme-selector" id="theme-selector" onchange="setTheme(this.value)" title="Select theme">
                        <option value="light">☀️ Light</option>
                        <option value="dark">🌙 Dark</option>
                        <option value="sepia">📜 Sepia</option>
                        <option value="nord">❄️ Nord</option>
                        <option value="solarized-light">🌅 Solarized Light</option>
                        <option value="solarized-dark">🌃 Solarized Dark</option>
                    </select>
                </div>
            </div>
            <div class="content %s">%s</div>
        </main>
    </div>

    <!-- Lightbox -->
    <div class="lightbox" id="lightbox" onclick="closeLightbox()">
        <span class="lightbox-close">&times;</span>
        <img id="lightbox-img" src="" alt="">
        <div class="lightbox-caption" id="lightbox-caption"></div>
    </div>

    <!-- Link Preview Popup -->
    <div class="link-preview" id="link-preview">
        <div class="link-preview-header" id="link-preview-header"></div>
        <div class="link-preview-content" id="link-preview-content"></div>
    </div>

    <!-- Prism JS -->
    <script src="/cdn/cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/prism.min.js"></script>
    <script src="/cdn/cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/plugins/line-numbers/prism-line-numbers.min.js"></script>
    <script src="/cdn/cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/components/prism-python.min.js"></script>
    <script src="/cdn/cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/components/prism-bash.min.js"></script>
    <script src="/cdn/cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/components/prism-javascript.min.js"></script>
    <script src="/cdn/cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/components/prism-json.min.js"></script>
    <script src="/cdn/cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/components/prism-typescript.min.js"></script>
    <script src="/cdn/cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/components/prism-css.min.js"></script>
    <script src="/cdn/cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/components/prism-sql.min.js"></script>
    <script src="/cdn/cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/components/prism-yaml.min.js"></script>
    <script src="/cdn/cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/components/prism-toml.min.js"></script>
    <script src="/cdn/cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/components/prism-go.min.js"></script>
    <script src="/cdn/cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/components/prism-rust.min.js"></script>
    <script src="/cdn/cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/components/prism-java.min.js"></script>
    <script src="/cdn/cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/components/prism-c.min.js"></script>
    <script src="/cdn/cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/components/prism-cpp.min.js"></script>
    <!-- KaTeX JS -->
    <script src="/cdn/cdn.jsdelivr.net/npm/katex@0.16.9/dist/katex.min.js"></script>
    <script src="/cdn/cdn.jsdelivr.net/npm/katex@0.16.9/dist/contrib/auto-render.min.js"></script>
    <!-- Mermaid JS -->
    <script src="/cdn/cdn.jsdelivr.net/npm/mermaid@10/dist/mermaid.min.js"></script>
    <script>
        // Theme management
        const themes = ['light', 'dark', 'sepia', 'nord', 'solarized-light', 'solarized-dark'];
        const darkThemes = ['dark', 'nord', 'solarized-dark'];

        function setTheme(theme) {
            // Remove all theme classes
            themes.forEach(t => {
                document.body.classList.remove('theme-' + t);
                document.body.classList.remove('dark-mode');
            });
            // Add new theme class
            if (theme !== 'light') {
                document.body.classList.add('theme-' + theme);
                if (darkThemes.includes(theme)) {
                    document.body.classList.add('dark-mode');
                }
            }
            // Save preference
            localStorage.setItem('theme', theme);
            // Update selector
            document.getElementById('theme-selector').value = theme;
            // Re-render mermaid for theme
            if (typeof mermaid !== 'undefined') {
                mermaid.initialize({
                    startOnLoad: false,
                    theme: darkThemes.includes(theme) ? 'dark' : 'default'
                });
                document.querySelectorAll('.mermaid').forEach(el => {
                    el.removeAttribute('data-processed');
                });
                mermaid.init();
            }
        }
        // Load saved theme
        const savedTheme = localStorage.getItem('theme') || 'light';
        // Migration from old darkMode setting
        if (!localStorage.getItem('theme') && localStorage.getItem('darkMode') === 'true') {
            setTheme('dark');
        } else {
            setTheme(savedTheme);
        }

        // Print / Export PDF
        function printDocument() {
            window.print();
        }

        // Link Preview
        let linkPreviewTimeout = null;
        let linkPreviewCache = {};
        const linkPreview = document.getElementById('link-preview');
        const linkPreviewHeader = document.getElementById('link-preview-header');
        const linkPreviewContent = document.getElementById('link-preview-content');

        function isInternalLink(href) {
            if (!href) return false;
            // Internal links are absolute paths starting with /
            if (href.startsWith('/') && !href.startsWith('//')) return true;
            // Relative paths without protocol
            if (!href.includes('://') && !href.startsWith('#')) return true;
            return false;
        }

        function showLinkPreview(e, href) {
            const filename = href.split('/').pop();
            linkPreviewHeader.textContent = filename;
            linkPreviewContent.innerHTML = '<span class="link-preview-loading">Loading...</span>';

            // Position the popup
            const rect = e.target.getBoundingClientRect();
            let left = rect.left;
            let top = rect.bottom + 5;

            // Ensure popup stays within viewport
            if (left + 400 > window.innerWidth) {
                left = window.innerWidth - 410;
            }
            if (top + 300 > window.innerHeight) {
                top = rect.top - 305;
            }

            linkPreview.style.left = left + 'px';
            linkPreview.style.top = top + 'px';
            linkPreview.classList.add('visible');

            // Check cache first
            if (linkPreviewCache[href]) {
                linkPreviewContent.innerHTML = linkPreviewCache[href];
                return;
            }

            // Fetch preview
            fetch('/preview' + href)
                .then(res => res.ok ? res.text() : Promise.reject('Not found'))
                .then(html => {
                    // Truncate content for preview
                    const div = document.createElement('div');
                    div.innerHTML = html;
                    // Remove TOC and toolbars
                    div.querySelectorAll('.toc, .search-toolbar, .json-toolbar').forEach(el => el.remove());
                    // Get first ~500 chars of text content
                    let preview = div.innerHTML;
                    if (preview.length > 2000) {
                        preview = preview.substring(0, 2000) + '...';
                    }
                    linkPreviewCache[href] = preview;
                    linkPreviewContent.innerHTML = preview;
                })
                .catch(() => {
                    linkPreviewContent.innerHTML = '<span class="link-preview-error">Preview not available</span>';
                });
        }

        function hideLinkPreview() {
            linkPreview.classList.remove('visible');
        }

        // Attach hover listeners to internal links
        document.addEventListener('DOMContentLoaded', function() {
            document.querySelectorAll('.markdown a, .content a').forEach(link => {
                const href = link.getAttribute('href');
                if (!isInternalLink(href) || href.startsWith('#')) return;

                link.addEventListener('mouseenter', function(e) {
                    clearTimeout(linkPreviewTimeout);
                    linkPreviewTimeout = setTimeout(() => showLinkPreview(e, href), 300);
                });

                link.addEventListener('mouseleave', function() {
                    clearTimeout(linkPreviewTimeout);
                    linkPreviewTimeout = setTimeout(hideLinkPreview, 200);
                });
            });

            // Keep preview visible when hovering over it
            linkPreview.addEventListener('mouseenter', function() {
                clearTimeout(linkPreviewTimeout);
            });
            linkPreview.addEventListener('mouseleave', function() {
                linkPreviewTimeout = setTimeout(hideLinkPreview, 200);
            });
        });

        // Sidebar functionality
        let sidebarOpen = localStorage.getItem('sidebarOpen') !== 'false';

        // ===== Favorites Management =====
        function getFavorites() {
            try {
                return JSON.parse(localStorage.getItem('fileViewerFavorites') || '[]');
            } catch { return []; }
        }
        function saveFavorites(favorites) {
            localStorage.setItem('fileViewerFavorites', JSON.stringify(favorites));
        }
        function addFavorite(path, name, isDir) {
            const favorites = getFavorites();
            if (!favorites.find(f => f.path === path)) {
                favorites.push({ path, name, isDir });
                saveFavorites(favorites);
            }
        }
        function removeFavorite(path) {
            const favorites = getFavorites().filter(f => f.path !== path);
            saveFavorites(favorites);
        }
        function isFavorite(path) {
            return getFavorites().some(f => f.path === path);
        }
        function toggleFavorite(path, name, isDir) {
            if (isFavorite(path)) {
                removeFavorite(path);
            } else {
                addFavorite(path, name, isDir);
            }
            renderSidebar();
        }

        // ===== Recent Files Management =====
        const MAX_RECENT_FILES = 15;
        function getRecentFiles() {
            try {
                return JSON.parse(localStorage.getItem('fileViewerRecent') || '[]');
            } catch { return []; }
        }
        function saveRecentFiles(recent) {
            localStorage.setItem('fileViewerRecent', JSON.stringify(recent));
        }
        function addRecentFile(path, name) {
            if (!path || path === '/' || name === 'Claude Code') return;
            let recent = getRecentFiles();
            // Remove if already exists (will re-add at top)
            recent = recent.filter(r => r.path !== path);
            // Add at beginning
            recent.unshift({ path, name, timestamp: Date.now() });
            // Limit to max
            if (recent.length > MAX_RECENT_FILES) {
                recent = recent.slice(0, MAX_RECENT_FILES);
            }
            saveRecentFiles(recent);
        }
        function removeRecentFile(path) {
            const recent = getRecentFiles().filter(r => r.path !== path);
            saveRecentFiles(recent);
            renderSidebar();
        }
        function clearRecentFiles() {
            localStorage.removeItem('fileViewerRecent');
            renderSidebar();
        }

        // ===== Panel Management =====
        function getCurrentFileDir() {
            const headerSpan = document.querySelector('.header-left span');
            const filepath = headerSpan ? headerSpan.textContent : '';
            if (filepath && filepath !== 'Claude Code' && filepath.includes('/')) {
                return filepath.substring(0, filepath.lastIndexOf('/')) || '/';
            }
            return '/';
        }
        function getPanelState() {
            const currentDir = getCurrentFileDir();
            try {
                const state = JSON.parse(localStorage.getItem('fileViewerPanels'));
                if (state && state.panels && state.panels.length > 0) {
                    // Always update Panel 1 to show current file's directory
                    state.panels[0].dir = currentDir;
                    return state;
                }
            } catch {}
            // Default: single panel with current directory
            return { panels: [{ id: 1, dir: currentDir }], nextId: 2 };
        }
        function savePanelState(state) {
            localStorage.setItem('fileViewerPanels', JSON.stringify(state));
        }
        function addPanel() {
            const state = getPanelState();
            if (state.panels.length >= 4) return;
            const lastDir = state.panels[state.panels.length - 1].dir;
            state.panels.push({ id: state.nextId++, dir: lastDir });
            savePanelState(state);
            renderSidebar();
        }
        function removePanel(id) {
            const state = getPanelState();
            if (state.panels.length <= 1) return;
            state.panels = state.panels.filter(p => p.id !== id);
            savePanelState(state);
            renderSidebar();
        }
        function updatePanelDir(id, dir) {
            const state = getPanelState();
            const panel = state.panels.find(p => p.id === id);
            if (panel) {
                panel.dir = dir;
                savePanelState(state);
            }
        }
        function updatePanelHeight(id, height) {
            const state = getPanelState();
            const panel = state.panels.find(p => p.id === id);
            if (panel) {
                panel.height = height;
                savePanelState(state);
            }
        }

        // ===== Panel Resizing =====
        let resizeState = null;
        function startPanelResize(e, panelAboveId, panelBelowId) {
            e.preventDefault();
            const panelAbove = document.getElementById('panel-' + panelAboveId);
            const panelBelow = document.getElementById('panel-' + panelBelowId);
            if (!panelAbove || !panelBelow) return;

            resizeState = {
                panelAboveId,
                panelBelowId,
                startY: e.clientY,
                startHeightAbove: panelAbove.offsetHeight,
                startHeightBelow: panelBelow.offsetHeight,
                handle: e.target
            };

            e.target.classList.add('resizing');
            document.body.classList.add('panel-resizing');
            document.addEventListener('mousemove', doPanelResize);
            document.addEventListener('mouseup', stopPanelResize);
        }
        function doPanelResize(e) {
            if (!resizeState) return;
            const delta = e.clientY - resizeState.startY;
            const minHeight = 100;
            let newHeightAbove = resizeState.startHeightAbove + delta;
            let newHeightBelow = resizeState.startHeightBelow - delta;

            if (newHeightAbove < minHeight) {
                newHeightAbove = minHeight;
                newHeightBelow = resizeState.startHeightAbove + resizeState.startHeightBelow - minHeight;
            }
            if (newHeightBelow < minHeight) {
                newHeightBelow = minHeight;
                newHeightAbove = resizeState.startHeightAbove + resizeState.startHeightBelow - minHeight;
            }

            const panelAbove = document.getElementById('panel-' + resizeState.panelAboveId);
            const panelBelow = document.getElementById('panel-' + resizeState.panelBelowId);
            if (panelAbove) {
                panelAbove.style.flex = 'none';
                panelAbove.style.height = newHeightAbove + 'px';
            }
            if (panelBelow) {
                panelBelow.style.flex = 'none';
                panelBelow.style.height = newHeightBelow + 'px';
            }
        }
        function stopPanelResize() {
            if (!resizeState) return;
            const panelAbove = document.getElementById('panel-' + resizeState.panelAboveId);
            const panelBelow = document.getElementById('panel-' + resizeState.panelBelowId);
            if (panelAbove) updatePanelHeight(resizeState.panelAboveId, panelAbove.offsetHeight);
            if (panelBelow) updatePanelHeight(resizeState.panelBelowId, panelBelow.offsetHeight);

            resizeState.handle.classList.remove('resizing');
            document.body.classList.remove('panel-resizing');
            document.removeEventListener('mousemove', doPanelResize);
            document.removeEventListener('mouseup', stopPanelResize);
            resizeState = null;
        }

        // ===== Toggle Sidebar =====
        function toggleSidebar() {
            const container = document.querySelector('.app-container');
            sidebarOpen = !sidebarOpen;
            container.classList.toggle('sidebar-hidden', !sidebarOpen);
            localStorage.setItem('sidebarOpen', sidebarOpen);
        }

        // ===== Render Sidebar =====
        function renderSidebar() {
            const content = document.getElementById('sidebar-content');
            while (content.firstChild) content.removeChild(content.firstChild);

            // Favorites section
            const favorites = getFavorites();
            const favSection = document.createElement('div');
            favSection.className = 'favorites-section';
            if (localStorage.getItem('favoritesCollapsed') === 'true') {
                favSection.classList.add('collapsed');
            }

            const favHeader = document.createElement('div');
            favHeader.className = 'favorites-header';
            favHeader.onclick = () => {
                favSection.classList.toggle('collapsed');
                localStorage.setItem('favoritesCollapsed', favSection.classList.contains('collapsed'));
                chevron.textContent = favSection.classList.contains('collapsed') ? '▶' : '▼';
            };
            const favTitle = document.createElement('span');
            favTitle.textContent = '⭐ Favorites (' + favorites.length + ')';
            const chevron = document.createElement('span');
            chevron.className = 'favorites-chevron';
            chevron.textContent = favSection.classList.contains('collapsed') ? '▶' : '▼';
            favHeader.appendChild(favTitle);
            favHeader.appendChild(chevron);
            favSection.appendChild(favHeader);

            const favList = document.createElement('div');
            favList.className = 'favorites-list';
            if (favorites.length === 0) {
                const empty = document.createElement('div');
                empty.className = 'favorites-empty';
                empty.textContent = 'No favorites yet';
                favList.appendChild(empty);
            } else {
                // Group favorites by directory
                const groups = {};
                favorites.forEach(fav => {
                    const dir = fav.path.substring(0, fav.path.lastIndexOf('/')) || '/';
                    if (!groups[dir]) groups[dir] = [];
                    groups[dir].push(fav);
                });

                // Sort directories and render groups
                Object.keys(groups).sort().forEach(dir => {
                    const groupDiv = document.createElement('div');
                    groupDiv.className = 'favorites-group';
                    const groupKey = 'favGroup_' + dir.replace(/[^a-zA-Z0-9]/g, '_');
                    if (localStorage.getItem(groupKey) === 'collapsed') {
                        groupDiv.classList.add('collapsed');
                    }

                    // Group header with directory path
                    const header = document.createElement('div');
                    header.className = 'favorites-group-header';
                    header.onclick = () => {
                        groupDiv.classList.toggle('collapsed');
                        localStorage.setItem(groupKey, groupDiv.classList.contains('collapsed') ? 'collapsed' : '');
                    };
                    const pathSpan = document.createElement('div');
                    pathSpan.className = 'favorites-group-path';
                    pathSpan.title = dir;
                    const pathText = document.createElement('span');
                    pathText.textContent = dir;
                    pathSpan.appendChild(pathText);
                    const countSpan = document.createElement('span');
                    countSpan.className = 'favorites-group-count';
                    countSpan.textContent = groups[dir].length;
                    header.appendChild(pathSpan);
                    header.appendChild(countSpan);
                    groupDiv.appendChild(header);

                    // Group items
                    const itemsDiv = document.createElement('div');
                    itemsDiv.className = 'favorites-group-items file-tree';
                    const ul = document.createElement('ul');
                    groups[dir].forEach(fav => {
                        const li = document.createElement('li');
                        const a = document.createElement('a');
                        a.className = 'tree-item';
                        a.href = fav.isDir ? 'javascript:void(0)' : fav.path;
                        if (fav.isDir) {
                            a.onclick = () => {
                                const state = getPanelState();
                                state.panels[0].dir = fav.path;
                                savePanelState(state);
                                renderSidebar();
                            };
                        }
                        if (!fav.isDir && location.pathname === fav.path) {
                            a.classList.add('active');
                        }
                        const icon = document.createElement('span');
                        icon.className = 'tree-icon';
                        icon.textContent = fav.isDir ? '📁' : getFileIcon(fav.path.substring(fav.path.lastIndexOf('.')), false);
                        const name = document.createElement('span');
                        name.className = 'tree-name';
                        name.textContent = fav.name;
                        const star = document.createElement('button');
                        star.className = 'star-btn favorited';
                        star.textContent = '★';
                        star.title = 'Remove from favorites';
                        star.onclick = (e) => { e.preventDefault(); e.stopPropagation(); toggleFavorite(fav.path, fav.name, fav.isDir); };
                        a.appendChild(icon);
                        a.appendChild(name);
                        a.appendChild(star);
                        li.appendChild(a);
                        ul.appendChild(li);
                    });
                    itemsDiv.appendChild(ul);
                    groupDiv.appendChild(itemsDiv);
                    favList.appendChild(groupDiv);
                });
            }
            favSection.appendChild(favList);
            content.appendChild(favSection);

            // Recent files section
            const recentFiles = getRecentFiles();
            const recentSection = document.createElement('div');
            recentSection.className = 'recent-section';
            if (localStorage.getItem('recentCollapsed') === 'true') {
                recentSection.classList.add('collapsed');
            }

            const recentHeader = document.createElement('div');
            recentHeader.className = 'recent-header';
            recentHeader.onclick = (e) => {
                if (e.target.tagName === 'BUTTON') return;
                recentSection.classList.toggle('collapsed');
                localStorage.setItem('recentCollapsed', recentSection.classList.contains('collapsed'));
                recentChevron.textContent = recentSection.classList.contains('collapsed') ? '▶' : '▼';
            };
            const recentTitle = document.createElement('span');
            recentTitle.textContent = '🕐 Recent (' + recentFiles.length + ')';
            const recentChevron = document.createElement('span');
            recentChevron.className = 'recent-chevron';
            recentChevron.textContent = recentSection.classList.contains('collapsed') ? '▶' : '▼';
            const clearBtn = document.createElement('button');
            clearBtn.className = 'recent-clear-btn';
            clearBtn.textContent = '✕';
            clearBtn.title = 'Clear history';
            clearBtn.onclick = (e) => { e.stopPropagation(); clearRecentFiles(); };
            if (recentFiles.length === 0) clearBtn.style.display = 'none';
            recentHeader.appendChild(recentTitle);
            recentHeader.appendChild(clearBtn);
            recentHeader.appendChild(recentChevron);
            recentSection.appendChild(recentHeader);

            const recentList = document.createElement('div');
            recentList.className = 'recent-list file-tree';
            if (recentFiles.length === 0) {
                const empty = document.createElement('div');
                empty.className = 'recent-empty';
                empty.textContent = 'No recent files';
                recentList.appendChild(empty);
            } else {
                const ul = document.createElement('ul');
                recentFiles.forEach(file => {
                    const li = document.createElement('li');
                    const a = document.createElement('a');
                    a.className = 'tree-item';
                    a.href = file.path;
                    if (location.pathname === file.path) {
                        a.classList.add('active');
                    }
                    const icon = document.createElement('span');
                    icon.className = 'tree-icon';
                    icon.textContent = getFileIcon(file.path.substring(file.path.lastIndexOf('.')), false);
                    const name = document.createElement('span');
                    name.className = 'tree-name';
                    name.textContent = file.name;
                    name.title = file.path;
                    const removeBtn = document.createElement('button');
                    removeBtn.className = 'recent-remove-btn';
                    removeBtn.textContent = '✕';
                    removeBtn.title = 'Remove from history';
                    removeBtn.onclick = (e) => { e.preventDefault(); e.stopPropagation(); removeRecentFile(file.path); };
                    a.appendChild(icon);
                    a.appendChild(name);
                    a.appendChild(removeBtn);
                    li.appendChild(a);
                    ul.appendChild(li);
                });
                recentList.appendChild(ul);
            }
            recentSection.appendChild(recentList);
            content.appendChild(recentSection);

            // Panels container
            const panelsContainer = document.createElement('div');
            panelsContainer.className = 'panels-container';
            const state = getPanelState();

            state.panels.forEach((panel, index) => {
                // Add resize handle before panel (except first)
                if (index > 0) {
                    const handle = document.createElement('div');
                    handle.className = 'panel-resize-handle';
                    handle.title = 'Drag to resize panels';
                    const prevPanelId = state.panels[index - 1].id;
                    handle.onmousedown = (e) => startPanelResize(e, prevPanelId, panel.id);
                    panelsContainer.appendChild(handle);
                }
                const panelEl = createPanelElement(panel, index, state.panels.length);
                // Apply saved height if available
                if (panel.height) {
                    panelEl.style.flex = 'none';
                    panelEl.style.height = panel.height + 'px';
                }
                panelsContainer.appendChild(panelEl);
            });

            content.appendChild(panelsContainer);

            // Load directories after panels are in the DOM
            state.panels.forEach(panel => {
                loadDirectoryForPanel(panel.id, panel.dir);
            });
        }

        function createPanelElement(panel, index, totalPanels) {
            const panelEl = document.createElement('div');
            panelEl.className = 'nav-panel';
            panelEl.id = 'panel-' + panel.id;

            const header = document.createElement('div');
            header.className = 'panel-header';
            const title = document.createElement('span');
            title.className = 'panel-title';
            title.textContent = 'Panel ' + (index + 1);

            const buttons = document.createElement('div');
            buttons.className = 'panel-buttons';

            // Add panel button (only on last panel)
            if (index === totalPanels - 1 && totalPanels < 4) {
                const addBtn = document.createElement('button');
                addBtn.className = 'panel-btn add-panel';
                addBtn.textContent = '+';
                addBtn.title = 'Add panel';
                addBtn.onclick = addPanel;
                buttons.appendChild(addBtn);
            }

            // Close panel button
            const closeBtn = document.createElement('button');
            closeBtn.className = 'panel-btn close-panel';
            closeBtn.textContent = '×';
            closeBtn.title = totalPanels <= 1 ? 'Cannot remove last panel' : 'Remove panel';
            closeBtn.disabled = totalPanels <= 1;
            closeBtn.onclick = () => removePanel(panel.id);
            buttons.appendChild(closeBtn);

            header.appendChild(title);
            header.appendChild(buttons);
            panelEl.appendChild(header);

            const panelContent = document.createElement('div');
            panelContent.className = 'panel-content';
            panelContent.id = 'panel-content-' + panel.id;
            panelEl.appendChild(panelContent);

            return panelEl;
        }

        async function loadDirectoryForPanel(panelId, dir) {
            const container = document.getElementById('panel-content-' + panelId);
            if (!container) return;
            try {
                const res = await fetch('/files?dir=' + encodeURIComponent(dir));
                const data = await res.json();
                if (data.error) {
                    container.textContent = 'Error: ' + data.error;
                    container.style.padding = '12px';
                    return;
                }
                updatePanelDir(panelId, data.dir);
                renderFileTreeForPanel(container, data, panelId);
            } catch (e) {
                container.textContent = 'Failed to load directory';
                container.style.padding = '12px';
            }
        }

        function renderFileTreeForPanel(container, data, panelId) {
            while (container.firstChild) container.removeChild(container.firstChild);
            container.style.padding = '';

            // Breadcrumb
            const breadcrumb = document.createElement('div');
            breadcrumb.className = 'sidebar-breadcrumb';
            const parts = data.dir.split('/').filter(p => p);
            let path = '';

            const rootLink = document.createElement('a');
            rootLink.href = 'javascript:void(0)';
            rootLink.textContent = '/';
            rootLink.title = 'Root';
            rootLink.onclick = () => loadDirectoryForPanel(panelId, '/');
            breadcrumb.appendChild(rootLink);

            parts.forEach((part, i) => {
                path += '/' + part;
                const p = path;
                if (i === parts.length - 1) {
                    breadcrumb.appendChild(document.createTextNode(part));
                } else {
                    const link = document.createElement('a');
                    link.href = 'javascript:void(0)';
                    link.textContent = part;
                    link.onclick = () => loadDirectoryForPanel(panelId, p);
                    breadcrumb.appendChild(link);
                    breadcrumb.appendChild(document.createTextNode('/'));
                }
            });
            container.appendChild(breadcrumb);

            // File tree
            const tree = document.createElement('div');
            tree.className = 'file-tree';
            const ul = document.createElement('ul');

            // Parent directory
            if (data.dir !== '/') {
                const li = document.createElement('li');
                const a = document.createElement('a');
                a.className = 'tree-item';
                a.href = 'javascript:void(0)';
                a.onclick = () => loadDirectoryForPanel(panelId, data.parent);
                const iconSpan = document.createElement('span');
                iconSpan.className = 'tree-icon';
                iconSpan.textContent = '📁';
                const nameSpan = document.createElement('span');
                nameSpan.className = 'tree-name';
                nameSpan.textContent = '..';
                a.appendChild(iconSpan);
                a.appendChild(nameSpan);
                li.appendChild(a);
                ul.appendChild(li);
            }

            data.files.forEach(file => {
                const icon = getFileIcon(file.ext, file.isDir);
                const li = document.createElement('li');
                const favorited = isFavorite(file.path);

                if (file.isDir) {
                    const a = document.createElement('a');
                    a.className = 'tree-item';
                    a.href = 'javascript:void(0)';
                    a.onclick = () => loadDirectoryForPanel(panelId, file.path);
                    const iconSpan = document.createElement('span');
                    iconSpan.className = 'tree-icon';
                    iconSpan.textContent = icon;
                    const nameSpan = document.createElement('span');
                    nameSpan.className = 'tree-name';
                    nameSpan.textContent = file.name;
                    const star = document.createElement('button');
                    star.className = 'star-btn' + (favorited ? ' favorited' : '');
                    star.textContent = favorited ? '★' : '☆';
                    star.title = favorited ? 'Remove from favorites' : 'Add to favorites';
                    star.onclick = (e) => { e.preventDefault(); e.stopPropagation(); toggleFavorite(file.path, file.name, true); };
                    a.appendChild(iconSpan);
                    a.appendChild(nameSpan);
                    a.appendChild(star);
                    li.appendChild(a);
                } else if (file.viewable) {
                    const a = document.createElement('a');
                    a.className = 'tree-item';
                    a.href = file.path;
                    if (location.pathname === file.path) a.classList.add('active');
                    const iconSpan = document.createElement('span');
                    iconSpan.className = 'tree-icon';
                    iconSpan.textContent = icon;
                    const nameSpan = document.createElement('span');
                    nameSpan.className = 'tree-name';
                    nameSpan.textContent = file.name;
                    const star = document.createElement('button');
                    star.className = 'star-btn' + (favorited ? ' favorited' : '');
                    star.textContent = favorited ? '★' : '☆';
                    star.title = favorited ? 'Remove from favorites' : 'Add to favorites';
                    star.onclick = (e) => { e.preventDefault(); e.stopPropagation(); toggleFavorite(file.path, file.name, false); };
                    a.appendChild(iconSpan);
                    a.appendChild(nameSpan);
                    a.appendChild(star);
                    li.appendChild(a);
                } else {
                    const span = document.createElement('span');
                    span.className = 'tree-item disabled';
                    span.title = file.size > 5242880 ? 'File too large (>5MB)' : 'Binary file';
                    const iconSpan = document.createElement('span');
                    iconSpan.className = 'tree-icon';
                    iconSpan.textContent = icon;
                    const nameSpan = document.createElement('span');
                    nameSpan.className = 'tree-name';
                    nameSpan.textContent = file.name;
                    span.appendChild(iconSpan);
                    span.appendChild(nameSpan);
                    li.appendChild(span);
                }
                ul.appendChild(li);
            });

            tree.appendChild(ul);
            container.appendChild(tree);
        }

        function getFileIcon(ext, isDir) {
            if (isDir) return '📁';
            const icons = {
                '.md': '📝', '.markdown': '📝',
                '.json': '📋',
                '.txt': '📄', '.text': '📄',
                '.html': '🌐', '.htm': '🌐',
                '.css': '🎨',
                '.js': '⚡',
                '.ts': '💎',
                '.go': '🔵',
                '.py': '🐍',
                '.rs': '🦀',
                '.java': '☕',
                '.c': '⚙️', '.cpp': '⚙️', '.h': '📎',
                '.sh': '🖥️', '.bash': '🖥️',
                '.yml': '⚙️', '.yaml': '⚙️',
                '.xml': '📰',
                '.svg': '🖼️', '.png': '🖼️', '.jpg': '🖼️', '.jpeg': '🖼️', '.gif': '🖼️',
                '.pdf': '📕',
            };
            return icons[ext] || '📄';
        }

        function initSidebar() {
            const container = document.querySelector('.app-container');
            if (!sidebarOpen) {
                container.classList.add('sidebar-hidden');
            }
            // Track current file in recent history
            const headerSpan = document.querySelector('.header-left span');
            const filepath = headerSpan ? headerSpan.textContent : '';
            const filename = filepath.split('/').pop();
            if (filepath && filename && filename !== 'Claude Code') {
                addRecentFile(filepath, filename);
            }
            renderSidebar();
        }

        // Copy code functionality
        function copyCode(id) {
            const code = document.getElementById(id);
            if (!code) return;
            const text = code.textContent;
            navigator.clipboard.writeText(text).then(() => {
                const btn = code.closest('.code-block').querySelector('.copy-btn');
                btn.textContent = '✓ Copied!';
                btn.classList.add('copied');
                setTimeout(() => {
                    btn.textContent = '📋 Copy';
                    btn.classList.remove('copied');
                }, 2000);
            });
        }

        // Lightbox
        function openLightbox(src, alt) {
            document.getElementById('lightbox-img').src = src;
            document.getElementById('lightbox-caption').textContent = alt || '';
            document.getElementById('lightbox').classList.add('active');
        }
        function closeLightbox() {
            document.getElementById('lightbox').classList.remove('active');
        }
        document.addEventListener('keydown', e => {
            if (e.key === 'Escape') closeLightbox();
        });

        // Live reload
        let lastMtime = null;
        setInterval(async () => {
            try {
                const res = await fetch('/mtime' + location.pathname);
                const mtime = await res.text();
                if (lastMtime === null) lastMtime = mtime;
                else if (mtime !== lastMtime) location.reload();
            } catch (e) {}
        }, 1000);

        // Text search functionality
        let searchMatches = [];
        let currentMatch = -1;
        let originalContent = null;

        function initSearch(contentId) {
            const content = document.getElementById(contentId);
            if (content) originalContent = content.cloneNode(true);
        }

        function textSearch(query, contentId) {
            const content = document.getElementById(contentId);
            const countEl = document.getElementById('search-count');
            if (!content || !originalContent) return;

            content.replaceWith(originalContent.cloneNode(true));
            const newContent = document.getElementById(contentId);
            originalContent = newContent.cloneNode(true);

            if (!query) {
                if (countEl) countEl.textContent = '';
                searchMatches = [];
                currentMatch = -1;
                return;
            }

            searchMatches = [];
            currentMatch = -1;
            const walker = document.createTreeWalker(newContent, NodeFilter.SHOW_TEXT, null, false);
            const textNodes = [];
            while (walker.nextNode()) textNodes.push(walker.currentNode);

            const regex = new RegExp('(' + query.replace(/[.*+?^${}()|[\]\\]/g, '\\$&') + ')', 'gi');
            textNodes.forEach(function(node) {
                if (regex.test(node.textContent)) {
                    const span = document.createElement('span');
                    const parts = node.textContent.split(regex);
                    parts.forEach(function(part) {
                        if (part.toLowerCase() === query.toLowerCase()) {
                            const mark = document.createElement('span');
                            mark.className = 'search-highlight';
                            mark.textContent = part;
                            searchMatches.push(mark);
                            span.appendChild(mark);
                        } else {
                            span.appendChild(document.createTextNode(part));
                        }
                    });
                    node.parentNode.replaceChild(span, node);
                }
            });

            if (countEl) countEl.textContent = searchMatches.length + ' résultat(s)';
            if (searchMatches.length > 0) goToMatch(0);
        }

        function goToMatch(index) {
            if (searchMatches.length === 0) return;
            if (currentMatch >= 0 && searchMatches[currentMatch]) {
                searchMatches[currentMatch].classList.remove('search-current');
            }
            currentMatch = (index + searchMatches.length) %% searchMatches.length;
            searchMatches[currentMatch].classList.add('search-current');
            searchMatches[currentMatch].scrollIntoView({ behavior: 'smooth', block: 'center' });
        }

        function nextMatch() { goToMatch(currentMatch + 1); }
        function prevMatch() { goToMatch(currentMatch - 1); }

        // Initialize on load
        document.addEventListener('DOMContentLoaded', function() {
            // Initialize Sidebar
            initSidebar();
            // Initialize Mermaid
            if (typeof mermaid !== 'undefined') {
                mermaid.initialize({
                    startOnLoad: true,
                    theme: document.body.classList.contains('dark-mode') ? 'dark' : 'default'
                });
            }
            // Initialize KaTeX
            if (typeof renderMathInElement !== 'undefined') {
                renderMathInElement(document.body, {
                    delimiters: [
                        {left: '$$', right: '$$', display: true},
                        {left: '$', right: '$', display: false}
                    ],
                    throwOnError: false
                });
            }
        });
    </script>
</body>
</html>`, title, filePath, contentClass, content)
}
