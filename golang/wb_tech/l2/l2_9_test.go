package l2

import "testing"

func TestUnpack_OK(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"a4bc2d5e", "aaaabccddddde"},
		{"abcd", "abcd"},
		{"qwe\\4\\5", "qwe45"},
		{"qwe\\45", "qwe44444"},
	}
	for _, test := range tests {
		got, err := Unpack(test.in)
		if err != nil {
			t.Fatalf("Unpack(%q) unexpected error: %v", test.in, err)
		}
		if got != test.want {
			t.Fatalf("Unpack(%q) = %q, want %q", test.in, got, test.want)
		}
	}
}