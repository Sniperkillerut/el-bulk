package authutil

import (
	"net/url"
	"os"
	"strings"
)

// GetCookieDomain returns the root domain for cookies based on the FRONTEND_ORIGIN.
// For example, if origin is "https://elbulk.com", it returns "elbulk.com".
// If origin is "https://demo.elbulk.co.uk", it handles common TLDs but 
// for elbulk.com specific production, it simplifies to the registrable domain.
func GetCookieDomain() string {
	origin := os.Getenv("FRONTEND_ORIGIN")
	if origin == "" {
		return ""
	}

	u, err := url.Parse(origin)
	if err != nil {
		return ""
	}

	host := u.Hostname()
	
	// If it's localhost or an IP, don't set a domain attribute (browser defaults to host-only)
	if host == "localhost" || host == "127.0.0.1" || strings.Contains(host, ":") {
		return ""
	}

	// For production elbulk.com / api.elbulk.com, we want ".elbulk.com"
	parts := strings.Split(host, ".")
	if len(parts) >= 2 {
		// Basic registrable domain extraction (e.g., elbulk.com)
		// Note: This is an simplification. For complex TLDs (com.au), more logic is needed.
		// But for elbulk.com it works perfectly.
		return "." + strings.Join(parts[len(parts)-2:], ".")
	}

	return ""
}
