package flog

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

func Log() *logrus.Entry {
	formatter := &logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timeLocal",
			logrus.FieldKeyLevel: "logLevel",
			logrus.FieldKeyMsg:   "short_message",
		},
	}
	formatter.TimestampFormat = "2006-01-02 15:04:05.000 +0800"
	logger.SetFormatter(formatter)
	group := os.Getenv("productName")
	env := os.Getenv("env")
	serviceName := os.Getenv("serviceName")
	podName := os.Getenv("MY_POD_NAME")
	return logger.WithFields(logrus.Fields{"group": group, "env": env, "serviceName": serviceName, "podName": podName})
}

func SetLevel(level logrus.Level) {
	logger.Level = level
}

func SetOutput(out io.Writer) {
	logger.SetOutput(out)
}

/*
func Debug(args ...interface{}) {
	logger.Debug(args...)
}

func Print(args ...interface{}) {
	logger.Info(args...)
}

func Info(args ...interface{}) {
	logger.Info(args...)
}

func Warn(args ...interface{}) {
	logger.Warn(args...)
}

func Error(args ...interface{}) {
	logger.Error(args...)
}

func Fatal(args ...interface{}) {
	logger.Fatal(args...)
	logrus.Exit(1)
}

func Panic(args ...interface{}) {
	logrus.Panic(args...)
}

// Entry Printf family functions

func Debugf(format string, args ...interface{}) {
	logrus.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	logrus.Infof(format, args...)
}

func Printf(format string, args ...interface{}) {
	logrus.Printf(format, args...)
}

func Warnf(format string, args ...interface{}) {
	logrus.Warnf(format, args...)
}

func Warningf(format string, args ...interface{}) {
	logger.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	logrus.Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	logrus.Fatalf(format, args...)
}

func Panicf(format string, args ...interface{}) {
	logrus.Panicf(format, args...)
}
*/
