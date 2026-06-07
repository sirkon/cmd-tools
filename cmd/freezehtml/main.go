package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/atotto/clipboard"
)

type Config struct {
	Embed    bool
	BGColor  string
	FontSize string
	Title    string
}

func main() {
	config := Config{}
	flag.BoolVar(&config.Embed, "embed", false, "Embed fonts as base64 in HTML")
	flag.StringVar(&config.BGColor, "bg", "", "Background color (auto-detect if empty)")
	flag.StringVar(&config.FontSize, "font-size", "14px", "Font size for terminal text")
	flag.StringVar(&config.Title, "title", "Terminal Output", "HTML page title")
	flag.Parse()

	var result string
	var err error

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		result, err = processStdin(config)
	} else {
		result, err = processFile(config)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(result)
}

func processStdin(config Config) (string, error) {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("reading stdin: %w", err)
	}

	cmd := exec.Command("aha", "--no-header")
	cmd.Stdin = bytes.NewReader(input)

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("running aha: %w", err)
	}

	bodyContent := out.String()

	if config.BGColor == "" {
		config.BGColor = detectBG(bodyContent)
	}

	return buildCompleteHTML(bodyContent, config), nil
}

func processFile(config Config) (string, error) {
	filePath, err := findHTMLFile()
	if err != nil {
		return "", fmt.Errorf("finding HTML file: %w", err)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("reading file %s: %w", filePath, err)
	}

	html := string(content)

	if config.BGColor == "" {
		config.BGColor = detectBG(html)
	}

	if isCompleteHTML(html) {
		html = enhanceExistingHTML(html, config)
	} else {
		html = buildCompleteHTML(html, config)
	}

	err = os.WriteFile(filePath, []byte(html), 0644)
	if err != nil {
		return "", fmt.Errorf("writing file %s: %w", filePath, err)
	}

	fmt.Fprintf(os.Stderr, "✅ Updated %s\n", filePath)

	return html, nil
}

func findHTMLFile() (string, error) {
	path, err := clipboard.ReadAll()
	if err == nil {
		path = strings.TrimSpace(path)
		if isValidHTMLFile(path) {
			return path, nil
		}
	}

	ghosttyDir := filepath.Join(os.TempDir(), "ghostty")
	if latest := findLatestHTML(ghosttyDir, true); latest != "" {
		fmt.Fprintf(os.Stderr, "📋 Found Ghostty screenshot: %s\n", latest)
		return latest, nil
	}

	if latest := findLatestHTML(os.TempDir(), false); latest != "" {
		fmt.Fprintf(os.Stderr, "📋 Found HTML in temp: %s\n", latest)
		return latest, nil
	}

	return "", fmt.Errorf("no HTML file found in clipboard or temp directories")
}

func findLatestHTML(dir string, requireGhostty bool) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}

	var latest string
	var latestTime time.Time

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(strings.ToLower(entry.Name()), ".html") {
			continue
		}

		if requireGhostty && !strings.Contains(strings.ToLower(entry.Name()), "ghostty") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().After(latestTime) {
			latest = filepath.Join(dir, entry.Name())
			latestTime = info.ModTime()
		}
	}

	return latest
}

func isValidHTMLFile(path string) bool {
	if !strings.HasSuffix(strings.ToLower(path), ".html") {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func buildCompleteHTML(bodyContent string, config Config) string {
	fontCSS := getFontCSS(config)
	bodyContent = cleanAHAOutput(bodyContent)

	fontStack := "'JetBrains Mono', 'Nerd Font Symbols', 'Cascadia Code', 'Fira Code', 'Courier New', monospace"

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>%s</title>
<style>
%s
*{margin:0;padding:0;box-sizing:border-box}
html,body{height:100%%}
body{
    background-color:%s;
    color:#d4d4d4;
    font-family:%s;
    font-size:%s;
    line-height:1.5;
    padding:16px
}
pre,.terminal-content{
    font-family:%s !important;
    white-space:pre !important;
    word-wrap:normal !important;
    overflow-x:auto;
    margin:0;
    padding:0;
    line-height:1.5;
    tab-size:4;
    -moz-tab-size:4
}
pre div,pre span,pre code,
.terminal-content div,.terminal-content span,.terminal-content code,
div[style*="font-family: monospace"],
div[style*="font-family:monospace"]{
    font-family:%s !important;
    white-space:pre !important
}
.aha{font-family:inherit;white-space:pre}
.aha .bold,b,strong{font-weight:700}
.aha .italic,i,em{font-style:italic}
.aha .underline,u{text-decoration:underline}
.aha .blink{animation:blink 1s step-end infinite}
@keyframes blink{50%%{opacity:0}}
.aha .strike,s,del{text-decoration:line-through}
::-webkit-scrollbar{width:8px;height:8px}
::-webkit-scrollbar-track{background:%s}
::-webkit-scrollbar-thumb{background:#555;border-radius:4px}
::-webkit-scrollbar-thumb:hover{background:#777}
::selection{background-color:#264f78;color:#fff}
</style>
</head>
<body>
<pre class="terminal-content">%s</pre>
</body>
</html>`,
		config.Title,
		fontCSS,
		config.BGColor,
		fontStack,
		config.FontSize,
		fontStack,
		fontStack,
		config.BGColor,
		bodyContent)
}

func cleanAHAOutput(html string) string {
	html = strings.TrimSpace(html)
	html = strings.ReplaceAll(html, "<pre>", "")
	html = strings.ReplaceAll(html, "</pre>", "")
	html = fixSpanNesting(html)

	// Replace Ghostty's monospace font-family with our font stack
	fontStack := "'JetBrains Mono', 'Nerd Font Symbols', 'Cascadia Code', 'Fira Code', 'Courier New', monospace"
	html = strings.ReplaceAll(html, "font-family: monospace;", "font-family: "+fontStack+";")
	html = strings.ReplaceAll(html, "font-family:monospace;", "font-family:"+fontStack+";")

	return html
}

func fixSpanNesting(html string) string {
	html = strings.ReplaceAll(html, "</span></span>", "</span>")
	html = strings.ReplaceAll(html, "<span></span>", "")
	return html
}

func enhanceExistingHTML(html string, config Config) string {
	fontCSS := getFontCSS(config)
	fontStack := "'JetBrains Mono', 'Nerd Font Symbols', 'Cascadia Code', 'Fira Code', monospace"

	// Replace Ghostty's monospace in inline styles before wrapping
	html = strings.ReplaceAll(html, "font-family: monospace;", "font-family: "+fontStack+";")
	html = strings.ReplaceAll(html, "font-family:monospace;", "font-family:"+fontStack+";")

	styleBlock := fmt.Sprintf(`
<style>
%s
body{background-color:%s !important;font-family:%s !important;font-size:%s !important;padding:16px !important}
pre,code,.terminal,.aha,.terminal-content{font-family:%s !important;white-space:pre !important;word-wrap:normal !important}
pre span,pre code,.terminal span,.terminal code,
div[style*="font-family"]{font-family:%s !important;white-space:pre !important}
</style>`, fontCSS, config.BGColor, fontStack, config.FontSize, fontStack, fontStack)

	if strings.Contains(html, "</head>") {
		html = strings.Replace(html, "</head>", styleBlock+"\n</head>", 1)
	} else if strings.Contains(html, "<head>") {
		html = strings.Replace(html, "<head>", "<head>\n"+styleBlock, 1)
	} else {
		html = fmt.Sprintf(`<!DOCTYPE html>
<html><head><meta charset="UTF-8">%s</head><body>%s</body></html>`, styleBlock, html)
	}

	if !strings.Contains(html, "<pre") {
		bodyRe := regexp.MustCompile(`(?i)<body[^>]*>(.*?)</body>`)
		html = bodyRe.ReplaceAllString(html, `<body><pre class="terminal-content">$1</pre></body>`)
	}

	html = addFontToExistingStyles(html)
	return html
}

func addFontToExistingStyles(html string) string {
	re := regexp.MustCompile(`style="([^"]*)"`)
	return re.ReplaceAllStringFunc(html, func(match string) string {
		if !strings.Contains(match, "font-family") {
			styleContent := re.FindStringSubmatch(match)[1]
			if !strings.HasSuffix(styleContent, ";") && styleContent != "" {
				styleContent += ";"
			}
			return fmt.Sprintf(`style="%s font-family:'JetBrains Mono','Nerd Font Symbols',monospace"`, styleContent)
		}
		return match
	})
}

func getFontCSS(config Config) string {
	if config.Embed {
		return getEmbeddedFontCSS()
	}
	return getLinkedFontCSS()
}

func getLinkedFontCSS() string {
	return `@font-face{font-family:'JetBrains Mono';src:url('https://cdn.jsdelivr.net/npm/@fontsource/jetbrains-mono@5.0.18/files/jetbrains-mono-latin-400-normal.woff2') format('woff2');font-weight:400;font-style:normal;font-display:swap}
@font-face{font-family:'JetBrains Mono';src:url('https://cdn.jsdelivr.net/npm/@fontsource/jetbrains-mono@5.0.18/files/jetbrains-mono-latin-700-normal.woff2') format('woff2');font-weight:700;font-style:normal;font-display:swap}
@font-face{font-family:'Nerd Font Symbols';src:url('https://raw.githubusercontent.com/ryanoasis/nerd-fonts/master/patched-fonts/NerdFontsSymbolsOnly/SymbolsNerdFont-Regular.ttf') format('truetype');font-weight:400;font-style:normal;font-display:swap;unicode-range:U+E000-F8FF,U+2500-259F,U+2600-26FF,U+2700-27BF,U+E0A0-E0FF,U+F000-F8FF,U+FB00-FB4F}`
}

func getEmbeddedFontCSS() string {
	fmt.Fprintf(os.Stderr, "📦 Downloading fonts for embedding...\n")

	regularFont := downloadFont("https://cdn.jsdelivr.net/npm/@fontsource/jetbrains-mono@5.0.18/files/jetbrains-mono-latin-400-normal.woff2")
	boldFont := downloadFont("https://cdn.jsdelivr.net/npm/@fontsource/jetbrains-mono@5.0.18/files/jetbrains-mono-latin-700-normal.woff2")
	nerdFont := downloadFont("https://raw.githubusercontent.com/ryanoasis/nerd-fonts/master/patched-fonts/NerdFontsSymbolsOnly/SymbolsNerdFont-Regular.ttf")

	if len(regularFont) == 0 || len(boldFont) == 0 {
		fmt.Fprintf(os.Stderr, "⚠️  Failed to download JetBrains Mono, falling back to CDN links\n")
		return getLinkedFontCSS()
	}

	regularBase64 := base64.StdEncoding.EncodeToString(regularFont)
	boldBase64 := base64.StdEncoding.EncodeToString(boldFont)

	css := fmt.Sprintf(`@font-face{font-family:'JetBrains Mono';src:url(data:font/woff2;base64,%s) format('woff2');font-weight:400;font-style:normal;font-display:swap}
@font-face{font-family:'JetBrains Mono';src:url(data:font/woff2;base64,%s) format('woff2');font-weight:700;font-style:normal;font-display:swap}`, regularBase64, boldBase64)

	if len(nerdFont) > 0 {
		nerdBase64 := base64.StdEncoding.EncodeToString(nerdFont)
		css += fmt.Sprintf(`
@font-face{font-family:'Nerd Font Symbols';src:url(data:font/truetype;base64,%s) format('truetype');font-weight:400;font-style:normal;font-display:swap;unicode-range:U+E000-F8FF,U+2500-259F,U+2600-26FF,U+2700-27BF,U+E0A0-E0FF,U+F000-F8FF,U+FB00-FB4F}`, nerdBase64)
		fmt.Fprintf(os.Stderr, "✅ Fonts embedded — JetBrains Mono: %d KB + %d KB, Nerd Font: %d KB\n",
			len(regularFont)/1024, len(boldFont)/1024, len(nerdFont)/1024)
	} else {
		fmt.Fprintf(os.Stderr, "⚠️  Failed to download Nerd Font, icons will not render\n")
		fmt.Fprintf(os.Stderr, "✅ JetBrains Mono embedded: %d KB + %d KB\n",
			len(regularFont)/1024, len(boldFont)/1024)
	}

	return css
}
func downloadFont(url string) []byte {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Warning: Failed to download font from %s: %v\n", url, err)
		return []byte{}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "⚠️  Warning: Bad status %d from %s\n", resp.StatusCode, url)
		return []byte{}
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Warning: Failed to read font data: %v\n", err)
		return []byte{}
	}

	return data
}

func isCompleteHTML(html string) bool {
	html = strings.ToLower(strings.TrimSpace(html))
	return strings.HasPrefix(html, "<!doctype") ||
		strings.HasPrefix(html, "<html") ||
		strings.Contains(html, "<head")
}

func detectBG(html string) string {
	if bg := os.Getenv("TERM_BG"); bg != "" {
		return bg
	}
	if bg := os.Getenv("GHOSTTY_BACKGROUND"); bg != "" {
		return bg
	}

	patterns := []string{
		`background-color:\s*([#\w]+)`,
		`background:\s*([#\w]+)`,
	}
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(html)
		if len(matches) > 1 {
			color := matches[1]
			if strings.HasPrefix(color, "#") || strings.HasPrefix(color, "rgb") {
				return color
			}
		}
	}

	term := os.Getenv("TERM_PROGRAM")
	switch term {
	case "ghostty":
		return "#1a1b26"
	case "iTerm.app", "Apple_Terminal":
		return "#000000"
	default:
		return "#1e1e1e"
	}
}
