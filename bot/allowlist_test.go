package bot

import (
	"os"
	"testing"
)

func TestParseAllowedChatIDs_Empty(t *testing.T) {
	os.Unsetenv("ALLOWED_CHAT_IDS")
	result := parseAllowedChatIDs()
	if result != nil {
		t.Errorf("Expected nil for empty env var, got %v", result)
	}
}

func TestParseAllowedChatIDs_Single(t *testing.T) {
	os.Setenv("ALLOWED_CHAT_IDS", "12345")
	defer os.Unsetenv("ALLOWED_CHAT_IDS")
	result := parseAllowedChatIDs()
	if !result[12345] {
		t.Errorf("Expected 12345 to be allowed")
	}
	if result[99999] {
		t.Errorf("Expected 99999 to not be allowed")
	}
}

func TestParseAllowedChatIDs_Multiple(t *testing.T) {
	os.Setenv("ALLOWED_CHAT_IDS", "111, 222, 333")
	defer os.Unsetenv("ALLOWED_CHAT_IDS")
	result := parseAllowedChatIDs()
	for _, id := range []int64{111, 222, 333} {
		if !result[id] {
			t.Errorf("Expected %d to be allowed", id)
		}
	}
	if len(result) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(result))
	}
}

func TestParseAllowedChatIDs_InvalidIgnored(t *testing.T) {
	os.Setenv("ALLOWED_CHAT_IDS", "111, abc, 333")
	defer os.Unsetenv("ALLOWED_CHAT_IDS")
	result := parseAllowedChatIDs()
	if len(result) != 2 {
		t.Errorf("Expected 2 valid entries, got %d", len(result))
	}
}
