package session

import (
	"encoding/base64"
	"testing"
)

func TestGenerateToken_ValidBase64URL(t *testing.T) {
	tok, err := GenerateToken()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Must decode cleanly as raw URL-safe base64 (no padding).
	decoded, err := base64.RawURLEncoding.DecodeString(tok)
	if err != nil {
		t.Fatalf("token is not valid base64url: %v", err)
	}

	// 32 random bytes encoded without padding → 43 base64url chars.
	if len(decoded) != 32 {
		t.Errorf("expected 32 decoded bytes, got %d", len(decoded))
	}
}

func TestGenerateToken_CorrectLength(t *testing.T) {
	tok, err := GenerateToken()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// base64.RawURLEncoding of 32 bytes produces ceil(32*4/3) = 43 characters.
	const wantLen = 43
	if len(tok) != wantLen {
		t.Errorf("expected token length %d, got %d", wantLen, len(tok))
	}
}

func TestGenerateToken_NoPaddingCharacters(t *testing.T) {
	tok, err := GenerateToken()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	for _, ch := range tok {
		if ch == '=' {
			t.Errorf("token contains padding character '=': %q", tok)
		}
	}
}

func TestGenerateToken_URLSafeCharacters(t *testing.T) {
	// Run several iterations to exercise the character set check.
	for i := 0; i < 50; i++ {
		tok, err := GenerateToken()
		if err != nil {
			t.Fatalf("iteration %d: expected no error, got: %v", i, err)
		}
		for _, ch := range tok {
			if !isBase64URLChar(ch) {
				t.Errorf("iteration %d: token contains non-URL-safe char %q in %q", i, ch, tok)
			}
		}
	}
}

func TestGenerateToken_Unique(t *testing.T) {
	const n = 100
	seen := make(map[string]struct{}, n)
	for i := 0; i < n; i++ {
		tok, err := GenerateToken()
		if err != nil {
			t.Fatalf("iteration %d: expected no error, got: %v", i, err)
		}
		if _, dup := seen[tok]; dup {
			t.Fatalf("duplicate token generated at iteration %d: %q", i, tok)
		}
		seen[tok] = struct{}{}
	}
}

// isBase64URLChar reports whether r is a valid base64url character (RFC 4648 §5).
func isBase64URLChar(r rune) bool {
	return (r >= 'A' && r <= 'Z') ||
		(r >= 'a' && r <= 'z') ||
		(r >= '0' && r <= '9') ||
		r == '-' || r == '_'
}
