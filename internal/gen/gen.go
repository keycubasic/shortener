package gen

import (
	"crypto/md5"
	"encoding/base64"
	"regexp"
)

func GenKey(link string) string {
	md5checksum := md5.Sum([]byte(link))
	base64encoded := base64.URLEncoding.EncodeToString([]byte(md5checksum[:]))
	re := regexp.MustCompile("[+/=_]")
	base64encoded = re.ReplaceAllString(base64encoded, "")
	key := base64encoded[:8]
	return key
}
