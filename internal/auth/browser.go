// Package auth provides browser-based authentication for the StackEye CLI.
package auth

import (
	"fmt"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
)

// OpenBrowser opens the specified URL in the default browser.
// Returns an error if the browser cannot be opened.
func OpenBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}

// BuildWebUIURL constructs the web UI authentication URL.
//
// Parameters:
//   - apiURL: The API URL (e.g., "https://api.stackeye.io")
//   - callbackURL: The local callback URL (e.g., "http://127.0.0.1:12345/callback")
//
// Returns the full web UI URL with the callback parameter.
func BuildWebUIURL(apiURL, callbackURL string) (string, error) {
	debugf("BuildWebUIURL: apiURL=%s, callbackURL=%s", apiURL, callbackURL)

	webURL, err := APIURLToWebURL(apiURL)
	if err != nil {
		debugf("BuildWebUIURL: APIURLToWebURL failed: %v", err)
		return "", err
	}
	debugf("BuildWebUIURL: webURL after transformation=%s", webURL)

	// Build full URL with callback parameter
	u, err := url.Parse(webURL)
	if err != nil {
		debugf("BuildWebUIURL: url.Parse failed: %v", err)
		return "", fmt.Errorf("failed to parse web URL: %w", err)
	}

	u.Path = "/cli-auth"
	q := u.Query()
	q.Set("callback", callbackURL)
	u.RawQuery = q.Encode()

	finalURL := u.String()
	debugf("BuildWebUIURL: final URL=%s", finalURL)

	return finalURL, nil
}

// APIURLToWebURL converts an API URL to the corresponding web UI URL.
//
// Transformations based on actual infrastructure:
//   - api.stackeye.io -> app.stackeye.io (production)
//   - api-dev.stackeye.io -> app-dev.stackeye.io
//   - api-staging.stackeye.io -> app-staging.stackeye.io
//
// For non-standard URLs (e.g., localhost, custom domains), returns unchanged.
func APIURLToWebURL(apiURL string) (string, error) {
	debugf("APIURLToWebURL: input=%s", apiURL)

	u, err := url.Parse(apiURL)
	if err != nil {
		debugf("APIURLToWebURL: url.Parse failed: %v", err)
		return "", fmt.Errorf("failed to parse API URL: %w", err)
	}
	debugf("APIURLToWebURL: parsed scheme=%s, host=%s", u.Scheme, u.Host)

	host := u.Host
	originalHost := host

	// Production: api.stackeye.io -> app.stackeye.io
	if host == "api.stackeye.io" {
		host = "app.stackeye.io"
		debugf("APIURLToWebURL: production transformation: %s -> %s", originalHost, host)
	} else if strings.HasPrefix(host, "api-") && strings.HasSuffix(host, ".stackeye.io") {
		// Environment-specific: api-dev.stackeye.io -> app-dev.stackeye.io
		//                       api-staging.stackeye.io -> app-staging.stackeye.io
		env := strings.TrimPrefix(host, "api-")
		env = strings.TrimSuffix(env, ".stackeye.io")
		debugf("APIURLToWebURL: extracted environment=%s", env)

		host = "app-" + env + ".stackeye.io"
		debugf("APIURLToWebURL: env transformation: %s -> %s", originalHost, host)
	} else {
		// For non-standard URLs (localhost, custom domains), return unchanged
		debugf("APIURLToWebURL: non-standard URL, returning unchanged: %s", apiURL)
		return apiURL, nil
	}

	webURL := &url.URL{
		Scheme: u.Scheme,
		Host:   host,
	}

	result := webURL.String()
	debugf("APIURLToWebURL: final result=%s", result)

	return result, nil
}

// IsLocalhost checks if an IP address is a localhost address.
// Accepts both IPv4 (127.x.x.x) and IPv6 (::1) localhost addresses.
func IsLocalhost(ip string) bool {
	// Check for IPv4 localhost
	if ip == "127.0.0.1" || strings.HasPrefix(ip, "127.") {
		return true
	}
	// Check for IPv6 localhost
	if ip == "::1" || ip == "[::1]" {
		return true
	}
	return false
}
