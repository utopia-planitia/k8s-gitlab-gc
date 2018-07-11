package gc

import "testing"

var isHashbasedTests = []struct {
	in  string
	out bool
}{
	{"project-shop-ci-54823-3a5db1781ab7cde0c53a3b53d995b75ee5873243", true},
	{"project-shop-ci-feature-cloud-upload-service", false},
	{"ci-user-user-feature-upgrade-framework-5-9-21", false},
	{"ci-user-user-feature-tes-562-api-stock-import", false},
	{"ci-username-playground-project-selenium-86421-efaf6b99b3ae85f7a405", true},
	{"ci-username-playground-project-selenium-86421-efaf6b99b3ae85f", true},
	{"ci-username-playground-project-selenium-86421-efaf6b99b3ae85", false},
	{"ci-username-playground-project-selenium-86421-f", false},
}

func TestIsHashbased(t *testing.T) {
	for _, tt := range isHashbasedTests {
		t.Run(tt.in, func(t *testing.T) {
			s, err := isHashbased(tt.in)
			if err != nil {
				t.Errorf("isHashbased (%s) returned an error", tt.in)
			}
			if s != tt.out {
				t.Errorf("isHashbased (%s) => %t want %t", tt.in, s, tt.out)
			}
		})
	}
}
