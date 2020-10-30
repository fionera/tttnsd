package client

import (
	"strings"
	"time"

	"github.com/miekg/dns"
)

type Client struct {
	addr   []string
	client *dns.Client
}

func newConfFromOS() (*dns.ClientConfig, error) {
	return dns.ClientConfigFromFile("/etc/resolv.conf")
}

func NewFromOS() (*Client, error) {
	os, err := newConfFromOS()
	if err != nil {
		return nil, err
	}

	var addresses []string
	for _, server := range os.Servers {
		if strings.Contains(server, ":") {
			server = "[" + server + "]"
		}
		addresses = append(addresses, server+":"+os.Port)
	}

	return NewClient(addresses...), nil
}

func NewClient(addresses ...string) *Client {
	return &Client{
		addr:   addresses,
		client: &dns.Client{},
	}
}

func (c *Client) Exchange(m *dns.Msg) (r *dns.Msg, rtt time.Duration, err error) {
	for _, s := range c.addr {
		r, rtt, err = c.client.Exchange(m, s)
		if err != nil {
			continue
		}

		return
	}

	return
}
