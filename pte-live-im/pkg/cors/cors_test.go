package cors

import "testing"

func TestMatchOrigin(t *testing.T) {
	allowed := []string{
		"http://localhost:11521",
		"http://127.0.0.1:11521",
		"https://admin.ptelive.com",
		"*.ptelive.com",
	}

	cases := []struct {
		origin string
		want   bool
	}{
		{"http://localhost:11521", true},
		{"https://admin.ptelive.com", true},
		{"https://im.ptelive.com", true},
		{"http://evil.ptelive.com", false},
	}

	for _, tc := range cases {
		got := false
		for _, item := range allowed {
			if matchOrigin(tc.origin, item) {
				got = true
				break
			}
		}
		if got != tc.want {
			t.Fatalf("origin=%q got=%v want=%v", tc.origin, got, tc.want)
		}
	}
}
