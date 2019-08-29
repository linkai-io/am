package webhooks

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// FormatUnixTimeRemaining takes in a time in unix and returns a string of hours/days remaining
func FormatUnixTimeRemaining(remaining int64) string {
	validTime := time.Unix(remaining, 0)
	ts := validTime.Sub(time.Now()).Hours()
	expires := ""
	if ts <= float64(24) {
		expires = fmt.Sprintf("%.01f hours", ts)
	} else {
		expires = fmt.Sprintf("%.0f days", ts/float64(24))
	}
	return expires
}

func IntToString(in []int32) string {
	b := make([]string, len(in))
	for i, v := range in {
		b[i] = strconv.FormatInt(int64(v), 10)
	}

	return strings.Join(b, ",")
}

// wasRedirected adds a '/' to the end of the load url and returns true if the load url got redirected to a different url, false otherwise
func wasRedirected(loadURL, url string) bool {
	load := loadURL

	if !strings.HasSuffix(loadURL, "/") {
		load = loadURL + "/"
	}

	return strings.Compare(load, url) != 0
}
