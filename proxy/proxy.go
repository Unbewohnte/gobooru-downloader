package proxy

import (
	"errors"
	"net/http"
	"net/url"

	"golang.org/x/net/proxy"
)

func NewProxyClient(proxyURL string) (*http.Client, error) {
	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		return nil, err
	}

	switch parsedURL.Scheme {
	case "http", "https":
		transport := &http.Transport{
			Proxy: http.ProxyURL(parsedURL),
		}
		return &http.Client{Transport: transport}, nil

	case "socks5":
		auth := &proxy.Auth{}
		if parsedURL.User != nil {
			auth.User = parsedURL.User.Username()
			auth.Password, _ = parsedURL.User.Password()
		}

		dialer, err := proxy.SOCKS5("tcp", parsedURL.Host, auth, proxy.Direct)
		if err != nil {
			return nil, err
		}

		transport := &http.Transport{
			Dial: dialer.Dial,
		}
		return &http.Client{Transport: transport}, nil

	default:
		return nil, errors.New("Proxy type not supported " + parsedURL.Scheme)
	}
}

func DoRequest(client *http.Client, method string, url string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return client.Do(req)
}
