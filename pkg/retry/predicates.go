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
