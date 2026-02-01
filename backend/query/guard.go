package query

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strconv"
	"strings"
)

var (
	whereRegex = regexp.MustCompile(`(?i)\\bwhere\\b`)
)

func StatementHash(stmt string) string {
	sum := sha256.Sum256([]byte(stmt))
	return hex.EncodeToString(sum[:])
}

func IsSQLWrite(stmt string) bool {
	s := strings.TrimSpace(strings.ToLower(stmt))
	return strings.HasPrefix(s, "insert") ||
		strings.HasPrefix(s, "update") ||
		strings.HasPrefix(s, "delete") ||
		strings.HasPrefix(s, "create") ||
		strings.HasPrefix(s, "alter") ||
		strings.HasPrefix(s, "drop") ||
		strings.HasPrefix(s, "truncate")
}

func HasWhere(stmt string) bool {
	return whereRegex.MatchString(stmt)
}

func IsDangerous(stmt string) bool {
	s := strings.TrimSpace(strings.ToLower(stmt))
	return strings.HasPrefix(s, "drop") || strings.HasPrefix(s, "alter") || strings.HasPrefix(s, "truncate")
}

func EnforceLimit(stmt string, maxRows int) string {
	if maxRows <= 0 {
		return stmt
	}
	s := strings.ToLower(stmt)
	if strings.Contains(s, "limit") {
		return stmt
	}
	return strings.TrimSpace(stmt) + " LIMIT " + strconv.Itoa(maxRows)
}
