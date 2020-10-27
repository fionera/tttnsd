package proto

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type AddressType int

const (
	Unknown AddressType = iota
	ServerInfoAddress
	ListAddress
	FolderInfoAddress
	ItemAddress
)

var (
	listRegex = regexp.MustCompile("^(?P<page_number>\\d+\\.|)(?P<folder_id>|[\\.\\w+]+|)list\\.$")
	getRegex  = regexp.MustCompile("^(?P<item_id>\\w+)((?P<folder_id>|[\\.\\w+]+)|).$")
)

func GetAddressType(baseAddress string, req string) AddressType {
	if req == baseAddress {
		return ServerInfoAddress
	}

	ts := strings.TrimSuffix(req, baseAddress)
	if listRegex.MatchString(ts) {
		matches := listRegex.FindStringSubmatch(ts)

		pageNumber := matches[1]
		if pageNumber == "" {
			return FolderInfoAddress
		}

		return ListAddress
	}

	if getRegex.MatchString(ts) {
		return ItemAddress
	}

	return Unknown
}

func EncodeListAddress(baseAddress string, page int, folderPath ...string) string {
	if len(folderPath) == 0 {
		return fmt.Sprintf("%d.list.%s", page, baseAddress)
	}

	return fmt.Sprintf("%d.%s.list.%s", page, strings.Join(folderPath, "."), baseAddress)
}

func DecodeListAddress(baseAddress string, req string) (int, []string) {
	// remove BaseDomain to make matching easier
	ts := strings.TrimSuffix(req, baseAddress)
	if !listRegex.MatchString(ts) {
		return 0, nil
	}

	matches := listRegex.FindStringSubmatch(ts)

	pageNumber := matches[1]
	folderIDs := matches[2]

	if folderIDs != "" {
		folderIDs = folderIDs[:len(folderIDs)-1]
	}

	if pageNumber != "" {
		pageNumber = pageNumber[:len(pageNumber)-1]
	} else {
		return 0, nil
	}

	p, err := strconv.Atoi(pageNumber)
	if err != nil {
		return 0, nil
	}

	if folderIDs == "" {
		return p, []string{}
	}

	return p, strings.Split(folderIDs, ".")
}

func EncodeFolderInfoAddress(baseAddress string, folderPath ...string) string {
	if len(folderPath) == 0 {
		return fmt.Sprintf("list.%s", baseAddress)
	}
	return fmt.Sprintf("%s.list.%s", strings.Join(folderPath, "."), baseAddress)
}

func DecodeFolderInfoAddress(baseAddress string, req string) []string {
	// remove BaseDomain to make matching easier
	ts := strings.TrimSuffix(req, baseAddress)
	if !listRegex.MatchString(ts) {
		return nil
	}

	matches := listRegex.FindStringSubmatch(ts)
	folderIDs := matches[2]

	if folderIDs != "" {
		folderIDs = folderIDs[:len(folderIDs)-1]
	} else {
		return []string{}
	}

	if matches[1] != "" {
		return nil
	}

	return strings.Split(folderIDs, ".")
}

func EncodeItemAddress(baseAddress string, itemID string, folderPath ...string) string {
	if len(folderPath) == 0 {
		return fmt.Sprintf("%s.%s", itemID, baseAddress)
	}

	return fmt.Sprintf("%s.%s.%s", itemID, strings.Join(folderPath, "."), baseAddress)
}

func DecodeItemAddress(baseAddress string, req string) (string, []string) {
	// remove BaseDomain to make matching easier
	ts := strings.TrimSuffix(req, baseAddress)
	if !getRegex.MatchString(ts) {
		return "", nil
	}

	matches := getRegex.FindStringSubmatch(ts)

	if matches[2] == "" {
		return matches[1], []string{}
	}

	return matches[1], strings.Split(matches[2][1:], ".")
}
