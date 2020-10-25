// +build windows

package client

func newConfFromOS() (*dns.ClientConfig, error) {
	return dns.ClientConfigFromFile("/msys64/etc/resolv.conf")
}
