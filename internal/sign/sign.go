package sign

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
)

// MD5Lower returns md5 hex digest in lowercase.
func MD5Lower(value string) string {
	sum := md5.Sum([]byte(value))
	return hex.EncodeToString(sum[:])
}

// Concat joins parts without separators.
func Concat(parts ...string) string {
	var b strings.Builder
	for _, p := range parts {
		b.WriteString(p)
	}
	return b.String()
}

// Sign joins parts as-is and calculates lowercase MD5.
func Sign(parts ...string) string {
	return MD5Lower(Concat(parts...))
}
