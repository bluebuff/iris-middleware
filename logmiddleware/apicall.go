package logmiddleware

import "time"

type ApiCall struct {
	IP string

	CurrentPath string
	MethodType  string

	RequestHeader  map[string]string
	ResponseHeader map[string]string
	ContextValues  map[string]string

	RequestBody  string
	ResponseBody string
	ResponseCode int
	Latency      time.Duration
}
