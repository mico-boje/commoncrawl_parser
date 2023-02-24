package utils

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

func ValidateURL(strURL string) error {
	u, err := url.Parse(strURL)
	if err != nil {
		return err
	}

	// check if URL is using HTTPS
	if u.Scheme != "https" {
		return fmt.Errorf("URL is not using HTTPS: %s", strURL)
	}

	// check if URL is an IP address
	ipRegex := `^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`
	ipMatch, err := regexp.MatchString(ipRegex, u.Hostname())
	if err != nil {
		return err
	}
	if ipMatch {
		return fmt.Errorf("URL is an IP address: %s", strURL)
	}

	allowedTLDs := []string{".edu", ".com", ".gov", ".gov.uk", ".mil", ".bank", ".airforce"}
	foundTLD := false
	for _, tld := range allowedTLDs {
		if strings.HasSuffix(u.Hostname(), tld) {
			foundTLD = true
			break
		}
	}
	if !foundTLD {
		return fmt.Errorf("unsupported top-level domain %s", u.Hostname())
	}

	return nil
}
