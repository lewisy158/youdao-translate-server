package logging

import (
	"fmt"
	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path"
	"strings"
	"time"
)

var logger *zap.SugaredLogger

// Init 默认初始化
func Init(logDir, logName string) {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "name",
		CallerKey:      "line",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   encodeCaller,
		EncodeName:     zapcore.FullNameEncoder,
	}
	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	err := os.MkdirAll(logDir, os.ModePerm)
	if err != nil {
		panic(fmt.Sprintf("日志模块初始化失败, 无法创建文件夹, 原因: %v", err))
	}

	logPath := path.Join(logDir, logName)
	logNameList := strings.Split(logName, ".")
	logBackupPath := path.Join(logDir, strings.Join(logNameList[:len(logNameList)-1], ".")) + ".%Y%m%d.log"
	ioWriter, err := rotatelogs.New(
		logBackupPath,                             // 日志切分文件路径
		rotatelogs.WithLinkName(logPath),          // 日志文件link路径
		rotatelogs.WithMaxAge(time.Hour*24*30),    // 日志最大保存30天
		rotatelogs.WithRotationTime(time.Hour*24), // 日志切分时间
	)
	if err != nil {
		panic(fmt.Sprintf("日志模块初始化失败, 无法创建日志ioWriter, 原因: %v", err))
	}
	allWriter := zapcore.AddSync(ioWriter)

	core := zapcore.NewTee(
		zapcore.NewCore(encoder, zapcore.Lock(os.Stdout), zapcore.DebugLevel),
		zapcore.NewCore(encoder, allWriter, zapcore.InfoLevel),
	)
	skip := zap.AddCallerSkip(1)
	logger = zap.New(core, zap.AddCaller(), skip).Sugar()
	Info("日志模块初始化完成")
}

// encodeCaller 自定义行号显示
func encodeCaller(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString("[" + caller.TrimmedPath() + "]")
}

func Debug(msg string) {
	if logger == nil {
		fmt.Println(msg)
	}
	logger.Debug(msg)
}

func Debugf(format string, v ...any) {
	if logger == nil {
		fmt.Println(fmt.Sprintf(format, v...))
	} else {
		logger.Debugf(fmt.Sprintf(format, v...))
	}
}

func Info(msg string) {
	if logger == nil {
		fmt.Println(msg)
	} else {
		logger.Info(msg)
	}
}

func Infof(format string, v ...any) {
	if logger == nil {
		fmt.Println(fmt.Sprintf(format, v...))
	} else {
		logger.Infof(fmt.Sprintf(format, v...))
	}
}

func Warn(msg string) {
	if logger == nil {
		fmt.Println(msg)
	} else {
		logger.Warn(msg)
	}
}

func Warnf(format string, v ...any) {
	if logger == nil {
		fmt.Println(fmt.Sprintf(format, v...))
	} else {
		logger.Warn(fmt.Sprintf(format, v...))
	}
}

func Error(msg string) {
	if logger == nil {
		fmt.Println(msg)
	} else {
		logger.Error(msg)
	}
}

func Errorf(format string, v ...any) {
	if logger == nil {
		fmt.Println(fmt.Sprintf(format, v...))
	} else {
		logger.Error(fmt.Sprintf(format, v...))
	}
}

func Panic(msg string) {
	if logger == nil {
		panic(msg)
	} else {
		logger.Panic(msg)
	}
}

func Panicf(format string, v ...any) {
	if logger == nil {
		panic(fmt.Sprintf(format, v...))
	} else {
		logger.Panic(fmt.Sprintf(format, v...))
	}
}

func Fatal(msg string) {
	if logger == nil {
		fmt.Println(msg)
		os.Exit(1)
	} else {
		logger.Fatal(msg)
	}
}

func Fatalf(format string, v ...any) {
	if logger == nil {
		fmt.Println(fmt.Sprintf(format, v...))
		os.Exit(1)
	} else {
		logger.Fatal(fmt.Sprintf(format, v...))
	}
}
