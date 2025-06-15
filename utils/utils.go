package utils

import (
	"os"
	"path/filepath"
	"strings"
)

func IconFor(file os.FileInfo) string {
	name := file.Name()
	if file.IsDir() {
		if strings.HasPrefix(name, ".") {
			return "🫥" // Hidden folder
		}
		return "📁"
	}

	// Hidden file
	if strings.HasPrefix(name, ".") {
		return "🫥"
	}

	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".go":
		return "🐹"
	case ".py":
		return "🐍"
	case ".js", ".ts":
		return "📜"
	case ".sh":
		return "🧮"
	case ".html", ".htm", ".css":
		return "🌐"
	case ".c", ".cpp", ".h", ".hpp":
		return "🧱"
	case ".rs":
		return "🦀"
	case ".java":
		return "☕"
	case ".json", ".yaml", ".yml", ".toml", ".ini":
		return "🛠️"
	case ".md", ".txt":
		return "📝"
	case ".log":
		return "📜"
	case ".pdf":
		return "📕"
	case ".zip", ".tar", ".gz", ".rar", ".7z":
		return "📦"
	case ".jpg", ".jpeg", ".png", ".gif", ".svg", ".webp":
		return "🖼️"
	case ".mp4", ".mov", ".avi", ".mkv":
		return "🎞️"
	case ".mp3", ".wav", ".flac":
		return "🎵"
	case ".exe", ".bin", ".out", ".app":
		return "⚙️"
	}

	// Check if executable (Unix-like)
	if file.Mode()&0111 != 0 {
		return "⚙️"
	}

	return "📄" // Generic file
}
