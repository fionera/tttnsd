package vfs

import (
	"bytes"
	"fmt"

	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
)

func newTorrent(torrentContent []byte) (*metainfo.MetaInfo, error) {
	var mi metainfo.MetaInfo
	d := bencode.NewDecoder(bytes.NewReader(torrentContent))
	err := d.Decode(&mi)
	if err != nil {
		return nil, fmt.Errorf("failed to decode torrent file: %v", err)
	}
	return &mi, nil
}
