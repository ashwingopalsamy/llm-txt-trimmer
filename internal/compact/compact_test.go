package compact

import "testing"

func TestSafeCollapsesOnlyRedundantBlankLines(t *testing.T) {
	in := "\n\nhello  \n\n\n\nworld\n\n"
	got, _, err := Transform(in, Safe)
	if err != nil {
		t.Fatal(err)
	}
	want := "hello  \n\nworld"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestCompactJoinsWrappedProseButKeepsStructure(t *testing.T) {
	in := "# Title\n\nThis is a wrapped\nparagraph for an LLM.\n\n- one\n- two\n"
	got, _, err := Transform(in, Compact)
	if err != nil {
		t.Fatal(err)
	}
	want := "# Title\n\nThis is a wrapped paragraph for an LLM.\n\n- one\n- two"
	if got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestDensePreservesFenceContents(t *testing.T) {
	in := "Before   text\ncontinued.\n\n```go\nfunc main() {\n    println(\"x\")\n}\n```\n\nAfter   text.\n"
	got, _, err := Transform(in, Dense)
	if err != nil {
		t.Fatal(err)
	}
	want := "Before text continued.\n```go\nfunc main() {\n    println(\"x\")\n}\n```\nAfter text."
	if got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestFrontMatterPreserved(t *testing.T) {
	in := "---\ntitle:  Keep  This\n---\n\nBody   here\n"
	got, _, err := Transform(in, Dense)
	if err != nil {
		t.Fatal(err)
	}
	want := "---\ntitle:  Keep  This\n---\nBody here"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestTamilUnicode(t *testing.T) {
	in := "தமிழ்    மொழி\nமிகவும் அழகு\n"
	got, _, err := Transform(in, Dense)
	if err != nil {
		t.Fatal(err)
	}
	want := "தமிழ் மொழி மிகவும் அழகு"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestInvalidMode(t *testing.T) {
	if _, _, err := Transform("x", Mode("wat")); err == nil {
		t.Fatal("expected error")
	}
}
