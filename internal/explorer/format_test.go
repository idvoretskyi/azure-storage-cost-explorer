package explorer

import "testing"

func TestFormatBytes(t *testing.T) {
	cases := []struct {
		in   float64
		want string
	}{
		{0, "0 B"},
		{1, "1.00 B"},
		{1024, "1.00 KB"},
		{1024 * 1024, "1.00 MB"},
		{1024 * 1024 * 1024, "1.00 GB"},
		{1.5 * 1024 * 1024, "1.50 MB"},
	}
	for _, c := range cases {
		got := FormatBytes(c.in)
		if got != c.want {
			t.Errorf("FormatBytes(%v) = %q; want %q", c.in, got, c.want)
		}
	}
}

func TestParseResourceGroup(t *testing.T) {
	id := "/subscriptions/abc/resourceGroups/myRG/providers/Microsoft.Storage/storageAccounts/foo"
	if got := ParseResourceGroup(id); got != "myRG" {
		t.Errorf("ParseResourceGroup = %q; want %q", got, "myRG")
	}
	if got := ParseResourceGroup("/no/match/here"); got != "" {
		t.Errorf("ParseResourceGroup empty case = %q", got)
	}
}
