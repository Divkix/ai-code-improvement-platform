// ABOUTME: RepositoryFile model for storing GitHub repository file data
// ABOUTME: Includes file content, metadata, and language detection functionality

package models

import (
	"path/filepath"
	"strings"
	"time"
)

// RepositoryFile represents a file fetched from a GitHub repository
type RepositoryFile struct {
	Path        string    `json:"path"`
	Name        string    `json:"name"`
	Content     string    `json:"content"`
	Language    string    `json:"language"`
	Size        int       `json:"size"`
	SHA         string    `json:"sha"`
	Type        string    `json:"type"` // "file" or "dir"
	FetchedAt   time.Time `json:"fetchedAt"`
	IsProcessed bool      `json:"isProcessed"`
}

// NewRepositoryFile creates a new RepositoryFile with detected language
func NewRepositoryFile(path, content, sha string, size int) *RepositoryFile {
	name := filepath.Base(path)
	language := DetectLanguageFromPath(path)
	
	return &RepositoryFile{
		Path:        path,
		Name:        name,
		Content:     content,
		Language:    language,
		Size:        size,
		SHA:         sha,
		Type:        "file",
		FetchedAt:   time.Now(),
		IsProcessed: false,
	}
}

// IsCodeFile returns true if the file is a code file that should be processed
func (rf *RepositoryFile) IsCodeFile() bool {
	// Skip binary files, images, and other non-code files
	skipExtensions := map[string]bool{
		".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".svg": true,
		".ico": true, ".bmp": true, ".tiff": true, ".webp": true,
		".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
		".zip": true, ".tar": true, ".gz": true, ".rar": true, ".7z": true,
		".exe": true, ".dll": true, ".so": true, ".dylib": true,
		".mp3": true, ".mp4": true, ".avi": true, ".mov": true, ".wav": true,
		".ttf": true, ".otf": true, ".woff": true, ".woff2": true, ".eot": true,
	}
	
	ext := strings.ToLower(filepath.Ext(rf.Path))
	if skipExtensions[ext] {
		return false
	}
	
	// Skip common directories and files
	skipPatterns := []string{
		"node_modules/", ".git/", ".svn/", ".hg/", 
		"vendor/", "venv/", "__pycache__/", ".pytest_cache/",
		"build/", "dist/", "target/", "bin/", "obj/",
		".DS_Store", "Thumbs.db", ".gitignore", ".gitkeep",
	}
	
	pathLower := strings.ToLower(rf.Path)
	for _, pattern := range skipPatterns {
		if strings.Contains(pathLower, strings.ToLower(pattern)) {
			return false
		}
	}
	
	// Include files with known code extensions or no extension (like Dockerfile)
	if rf.Language != "unknown" || ext == "" {
		return true
	}
	
	// Include files that look like code based on content (basic heuristic)
	if rf.Size > 0 && rf.Size < 1024*1024 { // Less than 1MB
		return true
	}
	
	return false
}

// GetLineCount returns the number of lines in the file
func (rf *RepositoryFile) GetLineCount() int {
	if rf.Content == "" {
		return 0
	}
	return strings.Count(rf.Content, "\n") + 1
}

// DetectLanguageFromPath detects programming language from file path/extension
func DetectLanguageFromPath(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	name := strings.ToLower(filepath.Base(path))
	
	// Language detection by extension
	languageMap := map[string]string{
		".js":   "javascript",
		".jsx":  "javascript",
		".ts":   "typescript",
		".tsx":  "typescript",
		".py":   "python",
		".pyw":  "python",
		".go":   "go",
		".java": "java",
		".kt":   "kotlin",
		".scala": "scala",
		".c":    "c",
		".h":    "c",
		".cpp":  "cpp",
		".cxx":  "cpp",
		".cc":   "cpp",
		".hpp":  "cpp",
		".cs":   "csharp",
		".php":  "php",
		".rb":   "ruby",
		".rs":   "rust",
		".swift": "swift",
		".m":    "objective-c",
		".mm":   "objective-c",
		".sh":   "shell",
		".bash": "shell",
		".zsh":  "shell",
		".fish": "shell",
		".ps1":  "powershell",
		".sql":  "sql",
		".html": "html",
		".htm":  "html",
		".css":  "css",
		".scss": "scss",
		".sass": "sass",
		".less": "less",
		".json": "json",
		".xml":  "xml",
		".yaml": "yaml",
		".yml":  "yaml",
		".toml": "toml",
		".ini":  "ini",
		".cfg":  "ini",
		".conf": "conf",
		".md":   "markdown",
		".rst":  "rst",
		".tex":  "latex",
		".r":    "r",
		".R":    "r",
		".pl":   "perl",
		".lua":  "lua",
		".vim":  "vimscript",
		".el":   "elisp",
		".clj":  "clojure",
		".hs":   "haskell",
		".ml":   "ocaml",
		".fs":   "fsharp",
		".ex":   "elixir",
		".exs":  "elixir",
		".erl":  "erlang",
		".dart": "dart",
		".pas":  "pascal",
		".asm":  "assembly",
		".s":    "assembly",
	}
	
	if lang, exists := languageMap[ext]; exists {
		return lang
	}
	
	// Special cases by filename
	specialFiles := map[string]string{
		"dockerfile":     "dockerfile",
		"makefile":       "makefile",
		"rakefile":       "ruby",
		"gemfile":        "ruby",
		"podfile":        "ruby",
		"vagrantfile":    "ruby",
		"gruntfile.js":   "javascript",
		"gulpfile.js":    "javascript",
		"webpack.config.js": "javascript",
		"package.json":   "json",
		"composer.json":  "json",
		"tsconfig.json":  "json",
		"tslint.json":    "json",
		"eslintrc":       "json",
		"babelrc":        "json",
		"gitignore":      "gitignore",
		"gitattributes":  "gitattributes",
		"editorconfig":   "editorconfig",
		"license":        "text",
		"readme":         "markdown",
		"changelog":      "markdown",
		"authors":        "text",
		"contributors":   "text",
	}
	
	// Remove dots and extensions for special file matching
	nameForMatching := strings.TrimPrefix(name, ".")
	nameForMatching = strings.TrimSuffix(nameForMatching, filepath.Ext(nameForMatching))
	
	if lang, exists := specialFiles[nameForMatching]; exists {
		return lang
	}
	
	return "unknown"
}

// IsValidForProcessing checks if file should be processed for code analysis
func (rf *RepositoryFile) IsValidForProcessing() bool {
	// Must be a code file
	if !rf.IsCodeFile() {
		return false
	}
	
	// Skip empty files
	if rf.Size == 0 || rf.Content == "" {
		return false
	}
	
	// Skip very large files (over 1MB)
	if rf.Size > 1024*1024 {
		return false
	}
	
	// Skip files with too few lines (less than 5 lines)
	if rf.GetLineCount() < 5 {
		return false
	}
	
	return true
}

// IsValidForStorage checks if file content is safe for database storage
func (rf *RepositoryFile) IsValidForStorage() (bool, string) {
	// Import the utils package for encoding validation
	// This will be used in the service layer
	return true, ""
}

// RepositoryTree represents the file tree structure
type RepositoryTree struct {
	SHA   string              `json:"sha"`
	URL   string              `json:"url"`
	Tree  []*RepositoryTreeItem `json:"tree"`
}

// RepositoryTreeItem represents an item in the repository tree
type RepositoryTreeItem struct {
	Path string `json:"path"`
	Mode string `json:"mode"`
	Type string `json:"type"` // "blob", "tree"
	SHA  string `json:"sha"`
	Size *int   `json:"size,omitempty"`
	URL  string `json:"url"`
}

// IsFile returns true if the tree item is a file
func (item *RepositoryTreeItem) IsFile() bool {
	return item.Type == "blob"
}

// IsDirectory returns true if the tree item is a directory
func (item *RepositoryTreeItem) IsDirectory() bool {
	return item.Type == "tree"
}