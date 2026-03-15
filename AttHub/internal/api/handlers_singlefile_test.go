package api

import "testing"

func TestParseSingleFileURLSnippet(t *testing.T) {
	t.Run("extract url from first four lines", func(t *testing.T) {
		got := parseSingleFileURLSnippet("line1\nline2\nurl: https://x.com/geekbb/status/1\nline4\nline5\n", 4)
		if got == nil {
			t.Fatalf("expected extracted url, got nil")
		}
		want := "https://x.com/geekbb/status/1"
		if *got != want {
			t.Fatalf("expected %q, got %q", want, *got)
		}
	})

	t.Run("ignore url after first four lines", func(t *testing.T) {
		got := parseSingleFileURLSnippet("1\n2\n3\n4\nurl: https://example.com\n", 4)
		if got != nil {
			t.Fatalf("expected nil, got %q", *got)
		}
	})

	t.Run("support uppercase key and fullwidth colon", func(t *testing.T) {
		got := parseSingleFileURLSnippet("URL： https://example.com/page\n", 4)
		if got == nil {
			t.Fatalf("expected extracted url, got nil")
		}
		want := "https://example.com/page"
		if *got != want {
			t.Fatalf("expected %q, got %q", want, *got)
		}
	})
}
