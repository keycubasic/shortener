package app

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"strings"
)

func GetKeyFromPath(path string) (string, error) {
	pathParams := strings.Split(path, "/")
	if len(pathParams) == 1 {
		return "", fmt.Errorf("Paramerter not found in source path %s", path)
	} else if len(pathParams) > 2 {
		return "", fmt.Errorf("Too many paramerters in source path %s", path)
	} else {
		return pathParams[1], nil
	}
}

func CreateShortLink(body []byte) (string, error) {
	// https://byby.dev/url-shortener
	md5checksum := md5.Sum(body)
	base64encoded := base64.URLEncoding.EncodeToString([]byte(md5checksum[:]))
	return base64encoded[:8], nil
}
