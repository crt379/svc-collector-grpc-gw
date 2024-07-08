package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/crt379/svc-collector-grpc-gw/internal/ctxvalue"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// var json = jsoniter.ConfigCompatibleWithStandardLibrary

type rspBodyLogWriter struct {
	writer http.ResponseWriter
	body   *bytes.Buffer
	status int
}

func (w *rspBodyLogWriter) Header() http.Header {
	return w.writer.Header()
}

func (w *rspBodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.writer.Write(b)
}

func (w *rspBodyLogWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.writer.WriteHeader(statusCode)
}

func WithRecover(handler http.Handler) http.Handler {
	return http.HandlerFunc(
		func(writer http.ResponseWriter, request *http.Request) {
			defer func() {
				if r := recover(); r != nil {
					ctx := request.Context()
					logger, _ := ctxvalue.LoggerContext{}.GetValue(ctx)
					logger.Error("panic", zap.Any("error", r))

					header := writer.Header()
					header.Add("Content-Type", "application/json")
					writer.WriteHeader(http.StatusInternalServerError)
					writer.Write([]byte(`{"msg": "服务内部错误"}`))
				}
			}()

			handler.ServeHTTP(writer, request)
		},
	)
}

func WithLogger(handler http.Handler) http.Handler {
	return http.HandlerFunc(
		func(writer http.ResponseWriter, request *http.Request) {
			ctx := request.Context()
			logger, _ := ctxvalue.LoggerContext{}.GetValue(ctx)

			start := time.Now()
			clientip := request.RemoteAddr
			path := request.URL.Path
			raw := request.URL.RawQuery
			if raw != "" {
				path = path + "?" + raw
			}

			traceid := request.Header.Get("x-access-trace-id")
			if traceid == "" {
				traceid = uuid.New().String()
			}
			spanid := uuid.New().String()
			logger = logger.With(zap.String("trace_id", traceid), zap.String("span_id", spanid))

			ctx = ctxvalue.LoggerContext{}.NewContext(ctx, logger)
			request = request.WithContext(ctx)

			logger.Info(
				"请求开始",
				zap.Time("start_time", start),
				zap.String("client", clientip),
				zap.String("method", request.Method),
				zap.String("url", path),
			)

			var buf []byte
			buf, _ = json.Marshal(&request.Header)
			logger.Info("req headers", zap.ByteString("headers", buf))

			request.Header.Set("x-access-trace-id", traceid)

			reqbody, err := io.ReadAll(request.Body)
			if err != nil {
				logger.Info("read req body error", zap.String("error", err.Error()))
			} else {
				request.Body = io.NopCloser(bytes.NewBuffer(reqbody))
				logger.Info("req body", zap.ByteString("body", reqbody))
			}

			var bodybyte []byte
			blw := &rspBodyLogWriter{body: bytes.NewBuffer(bodybyte), writer: writer}
			handler.ServeHTTP(blw, request)

			end := time.Now()
			latency := end.Sub(start)

			body := ""
			if blw.status >= 400 || blw.body.Len() < 1024 {
				body = blw.body.String()
			}

			logger, _ = ctxvalue.LoggerContext{}.GetValue(ctx)
			logger.Info(
				"请求结束",
				zap.Time("end_time", end),
				zap.String("latency", latency.String()),
				zap.String("body", body),
			)
		},
	)
}
