package logmiddleware

import (
	"bytes"
	"fmt"
	"github.com/kataras/iris/v12/context"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

type requestLoggerMiddleware struct {
	config Config
}

func New(cfg ...Config) context.Handler {
	c := DefaultConfig()
	if len(cfg) > 0 {
		c = cfg[0]
	}
	c.BuildSkipper()
	l := &requestLoggerMiddleware{config: c}
	return l.ServeHTTP
}

func (l *requestLoggerMiddleware) ServeHTTP(ctx context.Context) {
	// skip logs and serve the main request immediately
	if l.config.Skip != nil {
		if l.config.Skip(ctx) {
			ctx.Next()
			return
		}
	}

	apiCall := ApiCall{}
	startTime := time.Now()

	if l.config.RequestBody {
		apiCall.RequestBody = *ReadBody(ctx.Request())
	}

	ctx.Record()
	ctx.Next()

	// no time.Since in order to format it well after
	endTime := time.Now()
	apiCall.Latency = endTime.Sub(startTime)

	if l.config.Status {
		apiCall.ResponseCode = ctx.GetStatusCode()
	}

	if l.config.IP {
		apiCall.IP = ctx.RemoteAddr()
	}

	if l.config.Method {
		apiCall.MethodType = ctx.Method()
	}

	if l.config.Path {
		if l.config.Query {
			apiCall.CurrentPath = ctx.Request().URL.RequestURI()
		} else {
			apiCall.CurrentPath = ctx.Path()
		}
	}

	if l.config.ResponseBody {
		apiCall.ResponseBody = string(ctx.Recorder().Body())
	}

	apiCall.ContextValues = make(map[string]string, 0)
	if ctxKeys := l.config.MessageContextKeys; len(ctxKeys) > 0 {
		for _, key := range ctxKeys {
			msg := ctx.Values().Get(key)
			apiCall.ContextValues[key] = fmt.Sprintf("%v", msg)
		}
	}

	apiCall.RequestHeader = make(map[string]string, 0)
	if headerKeys := l.config.MessageRequestHeaderKeys; len(headerKeys) > 0 {
		for _, key := range headerKeys {
			msg := ctx.GetHeader(key)
			apiCall.RequestHeader[key] = fmt.Sprintf("%v", msg)
		}
	}

	apiCall.ResponseHeader = make(map[string]string, 0)
	if headerKeys := l.config.MessageResponseHeaderKeys; len(headerKeys) > 0 {
		writer := ctx.Recorder().Naive()
		for _, key := range headerKeys {
			for headerKey, headerValues := range writer.Header() {
				if headerKey == key {
					apiCall.ResponseHeader[key] = strings.Join(headerValues, " ")
				}
			}
		}
	}

	// print the logs
	if logFunc := l.config.LogFunc; logFunc != nil {
		logFunc(&apiCall)
		return
	} else if logFuncCtx := l.config.LogFuncCtx; logFuncCtx != nil {
		logFuncCtx(ctx, apiCall.Latency)
		return
	}

	fmt.Println("=================")
	// no new line, the framework's logger is responsible how to render each log.
	line := fmt.Sprintf("%v %4v %s %s %s", apiCall.ResponseCode, apiCall.Latency, apiCall.IP, apiCall.MethodType, apiCall.CurrentPath)

	// if context.StatusCodeNotSuccessful(ctx.GetStatusCode()) {
	// 	ctx.Application().Logger().Warn(line)
	// } else {
	ctx.Application().Logger().Info(line)
	// }

}

func ReadBody(req *http.Request) *string {
	save := req.Body
	var err error
	if req.Body == nil {
		req.Body = nil
	} else {
		save, req.Body, err = drainBody(req.Body)
		if err != nil {
			return nil
		}
	}
	b := bytes.NewBuffer([]byte(""))
	chunked := len(req.TransferEncoding) > 0 && req.TransferEncoding[0] == "chunked"
	if req.Body == nil {
		return nil
	}
	var dest io.Writer = b
	if chunked {
		dest = httputil.NewChunkedWriter(dest)
	}
	_, err = io.Copy(dest, req.Body)
	if chunked {
		dest.(io.Closer).Close()
	}
	req.Body = save
	body := b.String()
	return &body
}

// One of the copies, say from b to r2, could be avoided by using a more
// elaborate trick where the other copy is made during Request/Response.Write.
// This would complicate things too much, given that these functions are for
// debugging only.
func drainBody(b io.ReadCloser) (r1, r2 io.ReadCloser, err error) {
	var buf bytes.Buffer
	if _, err = buf.ReadFrom(b); err != nil {
		return nil, nil, err
	}
	if err = b.Close(); err != nil {
		return nil, nil, err
	}
	return ioutil.NopCloser(&buf), ioutil.NopCloser(bytes.NewReader(buf.Bytes())), nil
}
