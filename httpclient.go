package httpclient

import (
	"compress/gzip"
	"crypto/tls"
	fhttp "github.com/useflyent/fhttp"
	"io"
	"net/url"
	"strings"
)

func DoRequest(request Request) (resp *fhttp.Response, err error) {
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

	tr := &fhttp.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	if request.Proxy != "" {
		if proxyURL, err := url.Parse(request.Proxy); err == nil {
			tr.Proxy = fhttp.ProxyURL(proxyURL)
		} else {
			return
		}
	}

	if err != nil {
		return
	}

	client := &fhttp.Client{
		Transport: tr,
	}

	if request.Jar != nil {
		client.Jar = request.Jar
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

	return client.Do(req)
}

func Do(request Request) (response string, err error) {
	resp, err := DoRequest(request)
	defer resp.Body.Close()
	sb := strings.Builder{}

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
