package password

import (
	"strings"
	"testing"
)

func TestHash_ValidFormat(t *testing.T) {
	encoded, err := Hash("correct-horse-battery-staple")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Must begin with the Argon2id prefix.
	if !strings.HasPrefix(encoded, "$argon2id$v=19$") {
		t.Errorf("unexpected prefix in encoded hash: %q", encoded)
	}

	// Must contain exactly 5 dollar-sign separators producing 6 parts.
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 {
		t.Errorf("expected 6 parts after splitting on '$', got %d: %q", len(parts), encoded)
	}

	// Parameter segment must encode our secure defaults.
	if parts[3] != "m=65536,t=3,p=4" {
		t.Errorf("unexpected parameter segment %q", parts[3])
	}

	// Salt and hash segments must be non-empty.
	if parts[4] == "" {
		t.Error("encoded salt is empty")
	}
	if parts[5] == "" {
		t.Error("encoded hash is empty")
	}
}

func TestVerify_CorrectPassword(t *testing.T) {
	const pw = "hunter2"

	encoded, err := Hash(pw)
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}

	ok, err := Verify(pw, encoded)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if !ok {
		t.Error("expected Verify to return true for correct password")
	}
}

func TestVerify_WrongPassword(t *testing.T) {
	encoded, err := Hash("correct-password")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}

	ok, err := Verify("wrong-password", encoded)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if ok {
		t.Error("expected Verify to return false for wrong password")
	}
}

func TestVerify_MalformedHash(t *testing.T) {
	cases := []string{
		"",
		"notahash",
		"$argon2id$v=19$m=65536,t=3,p=4$invalidsalt",          // too few parts
		"$argon2id$v=19$m=65536,t=3,p=4$invalidsalt$hash$extra", // too many parts
		"$bcrypt$v=19$m=65536,t=3,p=4$c2FsdA$aGFzaA",           // wrong algorithm
		"$argon2id$v=0$m=65536,t=3,p=4$c2FsdA$aGFzaA",          // wrong version
		"$argon2id$v=19$badparams$c2FsdA$aGFzaA",                // bad parameter segment
		"$argon2id$v=19$m=65536,t=3,p=4$!!!$aGFzaA",            // bad base64 salt
		"$argon2id$v=19$m=65536,t=3,p=4$c2FsdA$!!!",            // bad base64 hash
	}

	for _, tc := range cases {
		ok, err := Verify("password", tc)
		if err == nil {
			t.Errorf("expected error for malformed hash %q, got ok=%v", tc, ok)
		}
		if ok {
			t.Errorf("expected false for malformed hash %q", tc)
		}
	}
}

func TestHash_DifferentPasswordsDifferentHashes(t *testing.T) {
	h1, err := Hash("password-one")
	if err != nil {
		t.Fatalf("Hash 1: %v", err)
	}

	h2, err := Hash("password-two")
	if err != nil {
		t.Fatalf("Hash 2: %v", err)
	}

	if h1 == h2 {
		t.Error("different passwords produced identical hashes")
	}
}

func TestHash_SamePasswordDifferentHashes(t *testing.T) {
	const pw = "same-password"

	h1, err := Hash(pw)
	if err != nil {
		t.Fatalf("Hash 1: %v", err)
	}

	h2, err := Hash(pw)
	if err != nil {
		t.Fatalf("Hash 2: %v", err)
	}

	if h1 == h2 {
		t.Error("same password produced identical hashes (salt not varying)")
	}
}
