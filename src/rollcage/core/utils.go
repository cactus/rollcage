package core

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/user"
	"path"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/cactus/gologit"
	"github.com/cheggaaa/pb"
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

// Fetch http file url to destination dest, with or without progress.
func FetchHTTPFile(url string, dest string, progress bool) (err error) {
	gologit.Debugf("Creating file: %s\n", dest)
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	var r io.Reader

	gologit.Debugf("Fetching url: %s\n", url)
	resp, err := http.Get(url)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Server return non-200 status: %v", resp.Status)
	}

	msgPrefix := fmt.Sprintf("%s: ", path.Base(dest))
	var bar *pb.ProgressBar
	i, _ := strconv.Atoi(resp.Header.Get("Content-Length"))
	if i > 0 && progress {
		bar = pb.New(i).Prefix(msgPrefix).SetUnits(pb.U_BYTES)
		bar.ShowSpeed = true
		bar.RefreshRate = time.Millisecond * 700
		bar.ShowFinalTime = false
		bar.ShowTimeLeft = false
		bar.Start()
		defer bar.Finish()
		r = bar.NewProxyReader(resp.Body)
	} else {
		r = resp.Body
	}
	_, err = io.Copy(out, r)
	return err
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
