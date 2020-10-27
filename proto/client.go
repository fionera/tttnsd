package proto

import (
	"fmt"

	"github.com/miekg/dns"

	"github.com/fionera/tttnsd/client"
)

type Client struct {
	initURL string
	c       *client.Client
	baseURL string
}

func NewClient(initURL string) (*Client, error) {
	c, err := client.NewFromOS()
	if err != nil {
		return nil, err
	}

	cl := &Client{
		c:       c,
		initURL: initURL,
	}

	answer, _, err := cl.doQuery(initURL)
	if err != nil {
		return nil, fmt.Errorf("invalid initial address: %v", err)
	}

	si := &ServerInfo{}
	si.Decode(answer)

	cl.baseURL = si.BaseURL

	return cl, nil
}

func (c *Client) GetDir(path ...string) ([]Item, error) {
	address := EncodeFolderInfoAddress(c.baseURL, path...)
	answer, extra, err := c.doQuery(address)
	if err != nil {
		return nil, err
	}

	fi := &FolderInfo{}
	fi.Decode(answer)

	var items []Item
	if len(extra) != 0 {
		for _, s := range extra {
			fp := &FolderPage{}
			fp.Decode(s)

			items = append(items, fp.Items...)
		}
	} else {
		for i := 0; i < fi.Pages; i++ {
			listAddress := EncodeListAddress(c.baseURL, i, path...)
			answer, _, err := c.doQuery(listAddress)
			if err != nil {
				return nil, err
			}

			fp := &FolderPage{}
			fp.Decode(answer)

			items = append(items, fp.Items...)
		}
	}

	return items, nil
}

func (c *Client) GetFile(itemID string, path ...string) (string, error) {
	address := EncodeItemAddress(c.baseURL, itemID, path...)
	answer, _, err := c.doQuery(address)
	if err != nil {
		return "", err
	}

	return answer, nil
}

func (c *Client) doQuery(address string) (answer string, extra []string, err error) {
	m := new(dns.Msg)
	m.SetQuestion(address+".", dns.TypeTXT)
	m.RecursionDesired = true

	r, _, err := c.c.Exchange(m)
	if err != nil {
		return
	}

	for _, a := range r.Answer {
		if txt, ok := a.(*dns.TXT); ok {
			answer = txt.Txt[0]
		}
	}

	for _, e := range r.Extra {
		if txt, ok := e.(*dns.TXT); ok {
			extra = append(extra, txt.Txt[0])
		}
	}

	if r.Rcode != dns.RcodeSuccess {
		err = fmt.Errorf("no data found")
	}

	return
}
