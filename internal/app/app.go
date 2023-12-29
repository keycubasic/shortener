package app

import (
	"crypto/md5"
	"encoding/base64"
	"regexp"

	"github.com/keycubasic/shortener/internal/storage"
)

func CreateShortLink(db *storage.Links, body []byte) (string, error) {
	// https://byby.dev/url-shortener
	link := string(body)
	// MD5 will generate a 32-character string from the set [a-fA-F0-9] (2^128)
	md5checksum := md5.Sum(body)
	// Base64 encode it to 32-character string from the set [a-zA-z0-9+/=]
	base64encoded := base64.URLEncoding.EncodeToString([]byte(md5checksum[:]))
	// Remove extra characters
	re := regexp.MustCompile("[+/=]")
	base64encoded = re.ReplaceAllString(base64encoded, "")
	// Trim up to eight characters from a Base64 set
	key := base64encoded[:8]
	/* Calculation probability of recurrence (P):
	Pr = 1/An^k = (n−k)!/n!
	n = 26 + 26 + 10 = 62 ([a-zA-z0-9])
	k = 8 ([:8])
	Pr = 7,335363631522695e-15
	The probability of recurrence is very very low.
	This approach provides a good balance between uniqueness and simplicity.
	*/
	db.SetLink(key, link)
	return key, nil
}
