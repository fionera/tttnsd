// +build linux

package client

import (
	"github.com/miekg/dns"
)

func newConfFromOS() (*dns.ClientConfig, error) {
	return dns.ClientConfigFromFile("/etc/resolv.conf")
}
