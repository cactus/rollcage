package core

import (
	"io"
	"os"
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

func SplitOutput(output string) [][]string {
	splitlines := make([][]string, 0)
	lines := strings.Split(strings.TrimSpace(output), "\n")
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

// Copies file source to destination dest.
func CopyFile(source string, dest string) (err error) {
	sf, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sf.Close()
	df, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer df.Close()
	_, err = io.Copy(df, sf)
	if err == nil {
		si, err := os.Stat(source)
		if err != nil {
			err = os.Chmod(dest, si.Mode())
		}

	}
	return
}
