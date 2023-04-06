package httpclient

import (
	fhttpjar "github.com/useflyent/fhttp/cookiejar"
	"time"
)

type Request struct {
	Timeout time.Duration
	Jar     *fhttpjar.Jar
	Headers [][]string
	URL     string
	Proxy   string
	Body    string
	Host    string
	Verb    string
}
