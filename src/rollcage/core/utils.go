package core

import (
	"bytes"
	"os/user"
	"strings"
	"unicode"

	"github.com/cactus/gologit"
)

func IsRoot() bool {
	u, err := user.Current()
	if err != nil {
		gologit.Fatal(err)
	}
	if u.Uid == "0" {
		return true
	}
	return false
}

func StringInSlice(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

func SplitOutput(output []byte) [][]string {
	splitlines := make([][]string, 0)
	lines := strings.Split(string(bytes.TrimSpace(output)), "\n")
	for _, line := range lines {
		splitlines = append(splitlines, strings.Split(line, "\t"))
	}
	return splitlines
}

func SplitFieldsQuoteSafe(s string) []string {
	haveQ := false
	return strings.FieldsFunc(s,
		func(c rune) bool {
			switch {
			case unicode.In(c, unicode.Quotation_Mark):
				if haveQ == true {
					haveQ = false
					return true
				}
				haveQ = true
				return false
			case unicode.IsSpace(c):
				if haveQ {
					return false
				}
				return true
			default:
				return false
			}
		})

}
