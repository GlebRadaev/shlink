package utils

import (
	"errors"
	"net/url"
)

func ValidateURL(urlStr string) (*url.URL, error) {
	uri, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return nil, errors.New("invalid URL format")
	}
	if !(uri.Scheme == "http" || uri.Scheme == "https") {
		return nil, errors.New("invalid URL scheme")
	}
	// off in development mode
	// _, err = net.LookupHost(uri.Host)
	// if err != nil {
	// 	return nil, errors.New("invalid domain name")
	// }

	return uri, nil
}
