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

type DubboLogger struct {
	Logger
	dynamicLevel zap.AtomicLevel
}

// NewDubboLogger new a DubboLogger
func NewDubboLogger(lg Logger, lv zap.AtomicLevel) *DubboLogger {
	return &DubboLogger{
		Logger:       lg,
		dynamicLevel: lv,
	}
}

// SetLoggerLevel use for set logger level
func (dl *DubboLogger) SetLoggerLevel(level string) bool {
	if _, ok := dl.Logger.(*zap.SugaredLogger); ok {
		if lv, err := zapcore.ParseLevel(level); err == nil {
			dl.dynamicLevel.SetLevel(lv)
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

// initZapLoggerWithSyncer init zap Logger with syncer
func initZapLoggerWithSyncer(conf *Config) *zap.Logger {
	core := zapcore.NewCore(
		conf.getEncoder(),
		conf.getLogWriter(),
		zap.NewAtomicLevelAt(conf.ZapConfig.Level.Level()),
	)

	return zap.New(core, zap.AddCaller(), zap.AddCallerSkip(conf.CallerSkip))
}

// getEncoder get encoder by config, zapcore support json and console encoder
func (c *Config) getEncoder() zapcore.Encoder {
	if c.ZapConfig.Encoding == "json" {
		return zapcore.NewJSONEncoder(c.ZapConfig.EncoderConfig)
	} else if c.ZapConfig.Encoding == "console" {
		return zapcore.NewConsoleEncoder(c.ZapConfig.EncoderConfig)
	}
	return nil
}
