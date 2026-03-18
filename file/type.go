package file

import (
	"path"
	"strings"
)

type Type int

const (
	TypeUnknown = Type(iota)
	TypePHP
	TypeBlade
	TypeEnv
	TypeTinker
)

// TypeByFilename finds the filetype based on filename
func TypeByFilename(filename string) Type {
	if strings.HasSuffix(filename, ".blade.php") {
		return TypeBlade
	}

	if strings.HasSuffix(filename, ".php") {
		return TypePHP
	}

	if strings.HasSuffix(filename, ".tinker") {
		return TypeTinker
	}

	filename = path.Base(filename)
	if filename == ".env" || strings.HasPrefix(filename, ".env.") {
		return TypeEnv
	}

	return TypeUnknown
}
