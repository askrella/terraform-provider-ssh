package ssh

import (
	. "github.com/onsi/gomega"
	"net"
	"net/url"
	"testing"
)

func TestHostParsing(t *testing.T) {
	RegisterTestingT(t)
	addresses := []string{
		"127.0.0.1", "localhost", "2a02:4f8:d014:b2f2::1",
	}

	for _, addr := range addresses {
		t.Run(addr, func(t *testing.T) {
			RegisterTestingT(t)

			ip := net.ParseIP(addr)
			if ip != nil {
				return
			}

			_, err := url.Parse(addr)
			Expect(err).ToNot(HaveOccurred())
		})
	}
}
