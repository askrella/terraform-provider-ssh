package ssh

import (
	. "github.com/onsi/gomega"
	"testing"
)

func TestParsePermissions(t *testing.T) {
	RegisterTestingT(t)

	tests := []struct {
		str      string
		expected uint32
	}{
		{"755", 0755},
		{"0755", 0755},
		{"777", 0777},
		{"0777", 0777},
		{"0600", 0600},
		{"600", 0600},
	}
	for _, test := range tests {
		t.Run(test.str, func(t *testing.T) {
			Expect(ParsePermissions(test.str)).To(Equal(test.expected))
		})
	}
}
