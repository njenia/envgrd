package scanner

import (
	"os"
	"path/filepath"
	"strings"
)

// Language represents a programming language
type Language string

const (
	LanguageJavaScript Language = "javascript"
	LanguageTypeScript Language = "typescript"
	LanguageGo         Language = "go"
	LanguagePython     Language = "python"
	LanguageRust       Language = "rust"
	LanguageJava       Language = "java"
	LanguageUnknown    Language = "unknown"
)

// FileInfo contains information about a file to be scanned
type FileInfo struct {
	Path          string
	Language      Language
	InIgnoredPath bool // True if this file is in a folder that should be ignored
}

// Scanner handles file discovery and filtering
type Scanner struct {
	excludeDirs  map[string]bool // Directory names to exclude (e.g., "node_modules")
	excludePaths []string        // Path patterns to exclude (e.g., "src/config", "k8s/*")
	excludeGlobs []string
	includeGlobs []string
	scanRoot     string // Root path being scanned (for relative path matching)
}

// NewScanner creates a new scanner with default exclusions
func NewScanner() *Scanner {
	return &Scanner{
		excludeDirs: map[string]bool{
			"node_modules": true,
			"vendor":       true,
			".git":         true,
			"build":        true,
			"dist":         true,
			"bin":          true,
			"out":          true,
			".next":        true,
			".cache":       true,
		},
	}
}

// SetExcludeGlobs sets glob patterns to exclude
func (s *Scanner) SetExcludeGlobs(globs []string) {
	s.excludeGlobs = globs
}

// SetIncludeGlobs sets glob patterns to include (overrides excludes)
func (s *Scanner) SetIncludeGlobs(globs []string) {
	s.includeGlobs = globs
}

// AddExcludeDirs adds additional directories to exclude from scanning
// Can be directory names (e.g., "config") or paths (e.g., "src/config")
func (s *Scanner) AddExcludeDirs(dirs []string) {
	for _, dir := range dirs {
		// If it contains a path separator, treat it as a path pattern
		if strings.Contains(dir, "/") || strings.Contains(dir, "\\") {
			s.excludePaths = append(s.excludePaths, dir)
		} else {
			// Otherwise treat it as a directory name
			s.excludeDirs[dir] = true
		}
	}
}

// SetScanRoot sets the root path being scanned (for relative path matching)
func (s *Scanner) SetScanRoot(root string) {
	s.scanRoot = root
}

// detectLanguage determines the language from file extension
func detectLanguage(path string) Language {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".js", ".jsx", ".mjs":
		return LanguageJavaScript
	case ".ts", ".tsx":
		return LanguageTypeScript
	case ".go":
		return LanguageGo
	case ".py":
		return LanguagePython
	case ".rs":
		return LanguageRust
	case ".java":
		return LanguageJava
	default:
		return LanguageUnknown
	}
}

// isExcludedDir checks if a directory should be excluded by name only
// Path-based exclusions are handled separately for files (we want to scan files in ignored paths)
func (s *Scanner) isExcludedDir(name string, _ string) bool {
	// Only check directory name exclusions (like node_modules, vendor, etc.)
	// Don't check path-based exclusions here - we want to scan files in ignored paths
	return s.excludeDirs[name]
}

// isBinaryFile checks if a file is likely binary
func isBinaryFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	binaryExts := map[string]bool{
		".png": true, ".jpg": true, ".jpeg": true, ".gif": true,
		".pdf": true, ".zip": true, ".tar": true, ".gz": true,
		".exe": true, ".dll": true, ".so": true, ".dylib": true,
		".woff": true, ".woff2": true, ".ttf": true, ".eot": true,
		".ico": true, ".svg": true, ".mp4": true, ".mp3": true,
	}
	return binaryExts[ext]
}

// matchesGlob checks if a path matches any of the glob patterns
func matchesGlob(path string, globs []string) bool {
	for _, glob := range globs {
		matched, _ := filepath.Match(glob, filepath.Base(path))
		if matched {
			return true
		}
		// Also try matching against full path
		matched, _ = filepath.Match(glob, path)
		if matched {
			return true
		}
	}
	return false
}

// shouldInclude checks if a file should be included based on include/exclude globs
func (s *Scanner) shouldInclude(path string) bool {
	// If include globs are specified, file must match at least one
	if len(s.includeGlobs) > 0 {
		return matchesGlob(path, s.includeGlobs)
	}
	// If exclude globs are specified, file must not match any
	if len(s.excludeGlobs) > 0 {
		return !matchesGlob(path, s.excludeGlobs)
	}
	return true
}

// isInIgnoredPath checks if a file path is within an ignored folder
func (s *Scanner) isInIgnoredPath(filePath string) bool {
	if s.scanRoot == "" || len(s.excludePaths) == 0 {
		return false
	}

	// Get relative path from scan root
	relPath, err := filepath.Rel(s.scanRoot, filePath)
	if err != nil {
		return false
	}

	// Normalize path separators to forward slashes for comparison
	relPathNormalized := filepath.ToSlash(relPath)

	// Check if any exclude path matches
	for _, excludePath := range s.excludePaths {
		// Normalize exclude path to forward slashes
		excludePathNormalized := filepath.ToSlash(excludePath)

		// Check if the file path starts with the exclude path
		if relPathNormalized == excludePathNormalized {
			return true
		}
		if strings.HasPrefix(relPathNormalized, excludePathNormalized+"/") {
			return true
		}
		// Support patterns like "src/config/*"
		if strings.HasSuffix(excludePathNormalized, "/*") {
			prefix := strings.TrimSuffix(excludePathNormalized, "/*")
			if strings.HasPrefix(relPathNormalized, prefix+"/") || relPathNormalized == prefix {
				return true
			}
		}
	}

	return false
}

// Scan recursively walks a directory and returns files to parse
func (s *Scanner) Scan(rootPath string) ([]FileInfo, error) {
	var files []FileInfo

	// Set scan root for relative path matching
	s.scanRoot = rootPath

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories that should be excluded (by name, not by path)
		// We want to scan files in ignored paths to track variables
		if info.IsDir() {
			// Only skip if it's excluded by name (like node_modules, vendor, etc.)
			// Don't skip if it's only in an ignored path - we want to scan those files
			if s.excludeDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if this file is in an ignored path
		inIgnoredPath := s.isInIgnoredPath(path)

		// If in ignored path, we still want to parse it to track variables,
		// but we'll exclude them from the missing report

		// Skip binary files
		if isBinaryFile(path) {
			return nil
		}

		// Check include/exclude globs
		if !s.shouldInclude(path) {
			return nil
		}

		// Detect language
		lang := detectLanguage(path)
		if lang == LanguageUnknown {
			return nil
		}

		files = append(files, FileInfo{
			Path:          path,
			Language:      lang,
			InIgnoredPath: inIgnoredPath,
		})

		return nil
	})

	return files, err
}
