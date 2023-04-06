package httpclient

import (
	"compress/gzip"
	"crypto/tls"
	fhttp "github.com/useflyent/fhttp"
	"io"
	"net/url"
	"strings"
)

func Do(request Request) (response string, err error) {
	if request.Host == "" {
		if u, err := url.Parse(request.URL); err == nil {
			request.Host = u.Host
		}
	}

	sortedHeaders := fhttp.Header{}
	var orderKeys []string
	for _, v := range request.Headers {
		if len(v) != 2 {
			continue
		}
		key := v[0]
		value := v[1]
		if strings.ToLower(key) == "host" {
			value = request.Host
		}
		sortedHeaders[key] = []string{value}
		orderKeys = append(orderKeys, strings.ToLower(key))
	}
	sortedHeaders[fhttp.HeaderOrderKey] = orderKeys

	proxyURL, err := url.Parse(request.Proxy)
	tr := &fhttp.Transport{
		Proxy: fhttp.ProxyURL(proxyURL),
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	if err != nil {
		return
	}

	client := &fhttp.Client{
		Transport: tr,
		Jar:       request.Jar,
	}

	if request.Timeout > 0 {
		client.Timeout = request.Timeout
	}

	var body io.Reader
	if request.Body != "" {
		body = strings.NewReader(request.Body)
	}

	req, err := fhttp.NewRequest(request.Verb, request.URL, body)
	if err != nil {
		return
	}
	req.Host = request.Host
	req.Header = sortedHeaders

	resp, err := client.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()
	sb := strings.Builder{}

	// ToDo: use readResponse
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return response, err
		}
		defer gzReader.Close()
		if _, err = io.Copy(&sb, gzReader); err != nil {
			return response, err
		}
	default:
		_, err = io.Copy(&sb, resp.Body)
	}
	response = sb.String()
	return
}
