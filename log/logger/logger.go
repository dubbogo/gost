/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package logger

import (
	"github.com/sirupsen/logrus"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger Logger

func init() {
	InitLogger(nil)
}

type DubboLogger struct {
	Logger
	DynamicLevel zap.AtomicLevel
}

// SetLoggerLevel use for set logger level
func (dl *DubboLogger) SetLoggerLevel(level string) bool {
	if _, ok := dl.Logger.(*zap.SugaredLogger); ok {
		if lv, err := zapcore.ParseLevel(level); err == nil {
			dl.DynamicLevel.SetLevel(lv)
			return true
		}
	} else if l, ok := dl.Logger.(*logrus.Logger); ok {
		if lv, err := logrus.ParseLevel(level); err == nil {
			l.SetLevel(lv)
			return true
		}
	}
	return false
}

// Logger is the interface for Logger types
type Logger interface {
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Debug(args ...interface{})
	Fatal(args ...interface{})

	Infof(fmt string, args ...interface{})
	Warnf(fmt string, args ...interface{})
	Errorf(fmt string, args ...interface{})
	Debugf(fmt string, args ...interface{})
	Fatalf(fmt string, args ...interface{})
}

// InitLogger use for init logger by @conf
func InitLogger(conf *Config) {
	var (
		zapLogger *zap.Logger
		config    = &Config{}
	)
	if conf == nil || conf.ZapConfig == nil {
		zapLoggerEncoderConfig := zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "message",
			StacktraceKey:  "stacktrace",
			EncodeLevel:    zapcore.CapitalColorLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}
		config.ZapConfig = &zap.Config{
			Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
			Development:      false,
			Encoding:         "console",
			EncoderConfig:    zapLoggerEncoderConfig,
			OutputPaths:      []string{"stdout"},
			ErrorOutputPaths: []string{"stderr"},
		}
	} else {
		config.ZapConfig = conf.ZapConfig
	}

	if conf != nil {
		config.CallerSkip = conf.CallerSkip
	}

	if config.CallerSkip == 0 {
		config.CallerSkip = 1
	}

	if conf == nil || conf.LumberjackConfig == nil {
		zapLogger, _ = config.ZapConfig.Build(zap.AddCaller(), zap.AddCallerSkip(config.CallerSkip))
	} else {
		config.LumberjackConfig = conf.LumberjackConfig
		zapLogger = initZapLoggerWithSyncer(config)
	}
	logger = &DubboLogger{Logger: zapLogger.Sugar(), DynamicLevel: config.ZapConfig.Level}
}

// SetLogger sets logger for dubbo and getty
func SetLogger(log Logger) {
	logger = log
}

// GetLogger gets the logger
func GetLogger() Logger {
	return logger
}

// OpsLogger use for the SetLoggerLevel
type OpsLogger interface {
	Logger
	SetLoggerLevel(level string) bool
}

// SetLoggerLevel use for set logger level
func SetLoggerLevel(level string) bool {
	if l, ok := logger.(OpsLogger); ok {
		return l.SetLoggerLevel(level)
	}
	return false
}
