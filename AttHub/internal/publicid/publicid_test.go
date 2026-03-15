package publicid

import "testing"

func TestFromAttachment(t *testing.T) {
	got, err := FromAttachment(42, "abc123", "demo.pdf", 0)
	if err != nil {
		t.Fatalf("FromAttachment returned error: %v", err)
	}
	if len(got) != Length {
		t.Fatalf("expected length=%d, got %q", Length, got)
	}

	if _, err := Normalize(got); err != nil {
		t.Fatalf("expected generated id to be valid, got err=%v", err)
	}

	got2, err := FromAttachment(42, "abc123", "demo.pdf", 1)
	if err != nil {
		t.Fatalf("FromAttachment with different attempt returned error: %v", err)
	}
	if got2 == got {
		t.Fatalf("expected different hash ids for different attempts, got same id=%q", got)
	}
}

func TestNormalize(t *testing.T) {
	cases := []struct {
		Input   string
		WantErr bool
	}{
		{Input: "a1b2c3d4e5f6", WantErr: false},
		{Input: "A1B2C3D4E5F6", WantErr: false},
		{Input: "A1B2C3", WantErr: true},
		{Input: "G1B2C3D4E5F6", WantErr: true},
	}

	for _, c := range cases {
		_, err := Normalize(c.Input)
		if c.WantErr && err == nil {
			t.Fatalf("Normalize(%q) expected error, got nil", c.Input)
		}
		if !c.WantErr && err != nil {
			t.Fatalf("Normalize(%q) expected nil, got err=%v", c.Input, err)
		}
	}
}
