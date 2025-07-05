package utils

import (
	"crypto/rand"
	"fmt"
	"strings"
)

func GenerateTrackingNumber() (string, error) {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	
	// GoExpress tracking number format: GEX + 8 characters
	return fmt.Sprintf("GEX%X", bytes), nil
}

func ValidateTrackingNumber(trackingNumber string) bool {
	return strings.HasPrefix(trackingNumber, "GEX") && len(trackingNumber) == 11
}