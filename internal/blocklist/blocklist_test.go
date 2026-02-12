package blocklist

import (
	"testing"
)

func TestBlocklistAdd(t *testing.T) {
	bl := NewBlocklist()

	bl.Add("ads.google.com")
	bl.Add("tracker.com")

	if bl.Count() != 2 {
		t.Errorf("Expected 2 domains, got %d", bl.Count())
	}

	if !bl.IsBlocked("ads.google.com") {
		t.Error("ads.google.com should be blocked")
	}

	if bl.IsBlocked("google.com") {
		t.Error("google.com should NOT be blocked")
	}

	if bl.IsBlocked("sifytulkarim.xyz") {
		t.Error("sifytulkarim.xyz should NOT be blocked")
	}
}

func TestNormalization(t *testing.T) {
	bl := NewBlocklist()

	bl.Add("Ads.Google.Com")

	// Should match regardless of case
    if !bl.IsBlocked("ads.google.com") {
        t.Error("Case-insensitive match failed")
    }
    
    if !bl.IsBlocked("ADS.GOOGLE.COM") {
        t.Error("Uppercase match failed")
    }
    
    // Should match with trailing dot
    if !bl.IsBlocked("ads.google.com.") {
        t.Error("Trailing dot match failed")
    }
}


func TestParseHostsLine(t *testing.T) {
    tests := []struct {
        input    string
        expected string
    }{
        {"0.0.0.0 ads.com", "ads.com"},
        {"127.0.0.1 localhost", "localhost"},
        {"# comment", ""},
        {"", ""},
        {"0.0.0.0 tracker.com # comment", "tracker.com"},
    }
    
    for _, test := range tests {
        result := parseHostsLine(test.input)
        if result != test.expected {
            t.Errorf("parseHostsLine(%q) = %q, want %q", 
                    test.input, result, test.expected)
        }
    }
}