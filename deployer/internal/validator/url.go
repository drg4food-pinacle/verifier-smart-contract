package validator

import (
	"net"
	"net/url"
	"regexp"
	"strconv"

	"github.com/go-playground/validator/v10"
)

const (
	minPort = 0
	maxPort = 65535
)

func gethUrl(fl validator.FieldLevel) bool {
	u, err := url.Parse(fl.Field().String())
	if err != nil {
		return false
	}

	// Acceptable schemes
	switch u.Scheme {
	case "http", "https", "ws", "wss":
		// OK
	default:
		return false
	}

	host := u.Hostname()
	if host != "localhost" && net.ParseIP(host) == nil {
		// Check if it's a valid hostname (DNS)
		if !isValidHostname(host) {
			return false
		}
	}

	// Validate port if present
	if u.Port() != "" {
		port, err := strconv.Atoi(u.Port())
		if err != nil {
			return false
		}
		if port < minPort || port > maxPort {
			return false
		}
	}

	return true
}

func isValidHostname(host string) bool {
	re := regexp.MustCompile(`^([a-zA-Z0-9-]+\.)*[a-zA-Z0-9-]+$`)
	return re.MatchString(host)
}
