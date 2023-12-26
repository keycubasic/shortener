package app

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"strings"
)

var shortLinks = make(map[string]string)

func CreateShortLink(body []byte) (string, error) {
	// https://byby.dev/url-shortener
	link := string(body)
	md5checksum := md5.Sum(body)
	base64encoded := base64.URLEncoding.EncodeToString([]byte(md5checksum[:]))
	key := base64encoded[:8]
	if val, ok := shortLinks[key]; ok && val != link {
		return "", fmt.Errorf("shortlink %s already used for URL %s", key, val)
	}
	shortLinks[key] = link
	return key, nil
}

func GetLinkByKey(key string) (string, bool) {
	val, ok := shortLinks[key]
	return val, ok
}

func SetLinkByKey(key string, link string) {
	shortLinks[key] = link
}

func GetKeyFromPath(path string) (string, error) {
	pathParams := strings.Split(path, "/")
	if len(pathParams) == 1 {
		return "", fmt.Errorf("paramerter not found in source path %s", path)
	} else if len(pathParams) > 2 {
		return "", fmt.Errorf("too many parameters in source path %s", path)
	} else {
		return pathParams[1], nil
	}
}
