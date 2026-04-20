package service

import (
	"testing"
	"time"
	"strings"
	"regexp"

	"github.com/stretchr/testify/assert"
)

func TestOrderService_GenerateOrderNumber(t *testing.T) {
	s := &OrderService{}

	orderNumber := s.GenerateOrderNumber()

	// Format should be EB-YYYYMMDD-HEX(16 chars)
	// Example: EB-20260325-ABCDEF1234567890

	assert.True(t, strings.HasPrefix(orderNumber, "EB-"), "Should start with EB-")

	today := time.Now().Format("20060102")
	assert.Contains(t, orderNumber, today, "Should contain today's date")

	// Regexp to match the full format
	// EB- followed by 8 digits, followed by -, followed by 16 hex chars
	re := regexp.MustCompile(`^EB-\d{8}-[0-9A-F]{16}$`)
	assert.True(t, re.MatchString(orderNumber), "Should match the expected format ^EB-\\d{8}-[0-9A-F]{16}$, got: %s", orderNumber)

	// Check for randomness (unlikely to have collision in 2 tries with 64 bits)
	anotherOrderNumber := s.GenerateOrderNumber()
	assert.NotEqual(t, orderNumber, anotherOrderNumber, "Sequential order numbers should be different")
}
