package pkg

import (
	"errors"
	"regexp"
)

func ValidateArgSyntax(param string) error {
	if len(param) > 8 {
		return errors.New("invalid length format \nExample: 20.5.1")
	}

	regex := regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)

	if !regex.MatchString(param) {
		return errors.New("invalid version format \nExample: 20.5.1")
	}

	return nil
}
