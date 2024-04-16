package gen

import (
	"crypto/md5"
	"encoding/base64"
	"regexp"
)

var re *regexp.Regexp

func init() {
	re = regexp.MustCompile("[+/=_]")
}

func GenKey(link string) string {
	md5checksum := md5.Sum([]byte(link))
	base64encoded := base64.URLEncoding.EncodeToString([]byte(md5checksum[:]))
	base64encoded = re.ReplaceAllString(base64encoded, "")
	return base64encoded[:8]
}
