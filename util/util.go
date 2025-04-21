package util

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"time"
)

func SHA1(in string) string {
	h := sha1.New()
	io.WriteString(h, in)
	return hex.EncodeToString(h.Sum(nil))
}

// based on suggestions from,
// https://groups.google.com/d/msg/golang-nuts/d0nF_k4dSx4/rPGgfXv6QCoJ
func UUID() (uuid string, err error) {
	b := make([]byte, 16)

	_, err = io.ReadFull(rand.Reader, b)
	if err != nil {
		return
	}

	uuid = fmt.Sprintf(
		"%x-%x-%x-%x-%x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:],
	)
	return
}

func Bold(str string) string {
	return fmt.Sprintf("\033[1m%s\033[0m", str)
}

func Color(color, str string) string {
	return fmt.Sprintf("\033[%sm%s\033[0m", color, str)
}

func TimestampToTime(timestamp *int64) (*time.Time, error) {
	if timestamp != nil {
		t := time.Unix(*timestamp, 0)
		return &t, nil
	}
	return &time.Time{}, errors.New("Not valid time")
}

type byModTime []os.FileInfo

func (f byModTime) Len() int      { return len(f) }
func (f byModTime) Swap(i, j int) { f[i], f[j] = f[j], f[i] }

func (f byModTime) Less(i, j int) bool {
	return f[i].ModTime().UnixNano() > f[j].ModTime().UnixNano()
}

// ReadDir reads the directory named by dirname and returns
// a list of sorted(by ModTime) directory entries.
func ReadDir(dirname string) ([]os.FileInfo, error) {
	f, err := os.Open(dirname)
	if err != nil {
		return nil, err
	}
	list, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return nil, err
	}
	sort.Sort(byModTime(list))
	return list, nil
}
