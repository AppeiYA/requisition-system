package generator

import "testing"

func TestTokenGenerator_Generate(t *testing.T) {
	t.Parallel()

	gen := &TokenGenerator{}

	first, err := gen.Generate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if first == "" {
		t.Fatal("expected non-empty token")
	}

	second, err := gen.Generate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if first == second {
		t.Fatal("expected consecutive tokens to differ")
	}
}

func TestTokenGenerator_Hash(t *testing.T) {
	t.Parallel()

	gen := &TokenGenerator{}

	if gen.Hash("token") != gen.Hash("token") {
		t.Fatal("expected hash to be deterministic")
	}
	if gen.Hash("token") == gen.Hash("other") {
		t.Fatal("expected different tokens to hash differently")
	}
	if gen.Hash("token") == "token" {
		t.Fatal("expected hash to differ from the raw token")
	}
}
