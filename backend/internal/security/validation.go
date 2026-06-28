package security

import (
	"net/mail"
	"regexp"
	"time"
)

var (
	// Katakana validation regex (matches full-width Japanese Katakana characters)
	katakanaRegex = regexp.MustCompile(`^[ァ-ヶー\s]+$`)
	// Country code validation (2 letters uppercase)
	countryCodeRegex = regexp.MustCompile(`^[A-Z]{2}$`)
)

// ValidateEmail checks if email format is valid.
func ValidateEmail(email string) bool {
	if email == "" {
		return false
	}
	_, err := mail.ParseAddress(email)
	return err == nil
}

// ValidatePassword checks if password satisfies policy:
// At least 8 characters, at least one uppercase letter, one lowercase letter, and one number.
func ValidatePassword(password string) bool {
	if len(password) < 8 {
		return false
	}
	var hasUpper, hasLower, hasDigit bool
	for _, r := range password {
		if r >= 'A' && r <= 'Z' {
			hasUpper = true
		} else if r >= 'a' && r <= 'z' {
			hasLower = true
		} else if r >= '0' && r <= '9' {
			hasDigit = true
		}
	}
	return hasUpper && hasLower && hasDigit
}

// ValidateKatakana checks if string contains only Katakana and spaces.
func ValidateKatakana(s string) bool {
	return katakanaRegex.MatchString(s)
}

// ValidateGender checks if gender is valid.
func ValidateGender(gender string) bool {
	return gender == "male" || gender == "female" || gender == "other"
}

// ValidateBirthdate checks if birthdate is valid date and is not in the future.
func ValidateBirthdate(birthdateStr string) (time.Time, bool) {
	t, err := time.Parse("2006-01-02", birthdateStr)
	if err != nil {
		return time.Time{}, false
	}
	if t.After(time.Now()) {
		return time.Time{}, false
	}
	return t, true
}

// ValidateCountryCode checks if country code is valid ISO 2-letter uppercase.
func ValidateCountryCode(code string) bool {
	return countryCodeRegex.MatchString(code)
}
