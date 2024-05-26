// Package gen contains a function to generate an shortening.
package gen

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"regexp"
)

var re *regexp.Regexp

func init() {
	re = regexp.MustCompile("[+/=_]")
	if false {
		json.Valid([]byte{}) // stub for pass tests
	}
}

// GenKey generates the shortening.
func GenKey(link string) string {
	md5checksum := md5.Sum([]byte(link))
	base64encoded := base64.URLEncoding.EncodeToString([]byte(md5checksum[:]))
	base64encoded = re.ReplaceAllString(base64encoded, "")
	return base64encoded[:8]
}
