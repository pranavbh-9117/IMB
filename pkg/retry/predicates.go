package retry

import (
	"errors"
	"net"
	"net/url"
	"strings"
)

func IsTransientError(err error) bool {
	if err == nil {
		return false
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout()
	}

	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return IsTransientError(urlErr.Err)
	}

	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "temporary") {
		return true
	}

	return false
}


func IsOAuthTransientError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	permanentKeywords := []string{
		"expired", "invalid audience", "invalid issuer",
		"email_verified", "unauthorized", "forbidden",
	}
	for _, kw := range permanentKeywords {
		if strings.Contains(msg, kw) {
			return false
		}
	}
	return IsTransientError(err)
}

func IsSMTPTransientError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	permanentKeywords := []string{
		"550", "551", "552", "553", "554", "535", "530",
		"user unknown", "relay access denied", "authentication failed",
		"invalid syntax", "bad recipient",
	}
	for _, kw := range permanentKeywords {
		if strings.Contains(msg, kw) {
			return false
		}
	}
	transientKeywords := []string{
		"421", "450", "451", "452",
		"greylisted", "try again later", "rate limit", "too many connections",
	}
	for _, kw := range transientKeywords {
		if strings.Contains(msg, kw) {
			return true
		}
	}
	return IsTransientError(err)
}
