package logging

import (
	"io"
	"log"
	"os"
	"time"

	"github.com/crt379/svc-collector-grpc-gw/internal"
	"github.com/crt379/svc-collector-grpc-gw/internal/config"
	"github.com/crt379/svc-collector-grpc-gw/internal/storage"

	"github.com/go-redis/redis"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger *zap.Logger
	json   = jsoniter.ConfigCompatibleWithStandardLibrary
)

var loglistKey = "service-collector_log"

const (
	logTmFmtWithMS = "2006-01-02 15:04:05.000"
	zapLoggerKey   = "zap-logger"
)

func init() {
	var err error
	var file *os.File
	var fileWS zapcore.WriteSyncer
	var eonsoleE zapcore.Encoder
	var fileCore zapcore.Core
	var cores []zapcore.Core
	var core zapcore.Core

	filename := "service-collector.log"
	if config.AppConfig.Log.File != "" {
		filename = config.AppConfig.Log.File
	}

	file, err = os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModeAppend|os.ModePerm)
	if err != nil {
		panic(err)
	}

	level := zapcore.Level(config.AppConfig.Log.Level)
	if level < zapcore.DebugLevel || level >= zapcore.InvalidLevel {
		level = zapcore.DebugLevel
	}

	fileWS = zapcore.AddSync(file)
	eonsoleE = consoleEncoder()
	fileCore = zapcore.NewCore(eonsoleE, fileWS, level)

	cores = []zapcore.Core{fileCore}

	if config.AppConfig.Log.Redis.Enabled {
		logrdb, err := storage.NewRedisClient(
			config.AppConfig.Log.Redis.Host,
			config.AppConfig.Log.Redis.Port,
			config.AppConfig.Log.Redis.Password,
			config.AppConfig.Log.Redis.DB,
		)
		if err != nil {
			log.Panicf(err.Error())
		}
		if config.AppConfig.Log.Redis.Key != "" {
			loglistKey = config.AppConfig.Log.Redis.Key
		}
		rw := &redisWriter{
			R: logrdb,
			K: loglistKey,
		}
		rwWS := zapcore.AddSync(rw)
		encoderJ := jsonEncoder()
		redisCore := zapcore.NewCore(encoderJ, rwWS, level)
		cores = append(cores, redisCore)
	}

	core = zapcore.NewTee(cores...)
	logger = zap.New(
		core,
		zap.AddCaller(),
		zap.AddStacktrace(zap.ErrorLevel),
	)

	loggerG := logger.With(zap.String("host", config.AppConfig.Host), zap.String("source", internal.Name))
	zap.ReplaceGlobals(loggerG)
}

func consoleEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format(logTmFmtWithMS))
	}
	encoderConfig.NewReflectedEncoder = newReflectedEncoder

	return zapcore.NewConsoleEncoder(encoderConfig)
}

func jsonEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format(logTmFmtWithMS))
	}
	encoderConfig.NewReflectedEncoder = newReflectedEncoder

	return zapcore.NewJSONEncoder(encoderConfig)
}

func LoggerSync() {
	logger.Sync()
}

type redisWriter struct {
	K string
	R *redis.Client
}

func (w *redisWriter) Write(p []byte) (int, error) {
	if w.R == nil || w.K == "" {
		return 0, nil
	}

	n, err := w.R.RPush(w.K, p).Result()

	return int(n), err
}

func newReflectedEncoder(w io.Writer) zapcore.ReflectedEncoder {
	enc := &reflectedEncoder{w: w}
	return enc
}

type reflectedEncoder struct {
	w   io.Writer
	err error
}

func (enc *reflectedEncoder) Encode(obj any) error {
	if enc.err != nil {
		return enc.err
	}

	b, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	if _, err = enc.w.Write(b); err != nil {
		enc.err = err
	}
	return err
}

type ZapLogger struct {
	logger zap.Logger
}

// NewZapLogger 创建封装了zap的对象，该对象是对LoggerV2接口的实现
func NewZapLogger(logger *zap.Logger) *ZapLogger {
	return &ZapLogger{
		logger: *logger,
	}
}

func (zl *ZapLogger) Info(args ...interface{}) {
	zl.logger.Sugar().Info(args...)
}

func (zl *ZapLogger) Infoln(args ...interface{}) {
	zl.logger.Sugar().Info(args...)
}
func (zl *ZapLogger) Infof(format string, args ...interface{}) {
	zl.logger.Sugar().Infof(format, args...)
}

func (zl *ZapLogger) Warning(args ...interface{}) {
	zl.logger.Sugar().Warn(args...)
}

func (zl *ZapLogger) Warningln(args ...interface{}) {
	zl.logger.Sugar().Warn(args...)
}

func (zl *ZapLogger) Warningf(format string, args ...interface{}) {
	zl.logger.Sugar().Warnf(format, args...)
}

func (zl *ZapLogger) Error(args ...interface{}) {
	zl.logger.Sugar().Error(args...)
}

func (zl *ZapLogger) Errorln(args ...interface{}) {
	zl.logger.Sugar().Error(args...)
}

func (zl *ZapLogger) Errorf(format string, args ...interface{}) {
	zl.logger.Sugar().Errorf(format, args...)
}

func (zl *ZapLogger) Fatal(args ...interface{}) {
	zl.logger.Sugar().Fatal(args...)
}

func (zl *ZapLogger) Fatalln(args ...interface{}) {
	zl.logger.Sugar().Fatal(args...)
}

func (zl *ZapLogger) Fatalf(format string, args ...interface{}) {
	zl.logger.Sugar().Fatalf(format, args...)
}

func (zl *ZapLogger) V(v int) bool {
	return false
}
