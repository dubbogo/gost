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
	"context"
	"path/filepath"
	"testing"
)

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

const testFile = "ctx_logging_test.go"

type traceIDKey struct{}

// traceLogger adds the trace ID in the context to each entry, the way a
// tracing-aware logger would.
type traceLogger struct {
	*DubboLogger
}

func (l *traceLogger) with(ctx context.Context) *zap.SugaredLogger {
	sugar := l.Logger.(*zap.SugaredLogger)
	if id, ok := ctx.Value(traceIDKey{}).(string); ok {
		return sugar.With("trace_id", id)
	}
	return sugar
}

func (l *traceLogger) CtxDebug(ctx context.Context, args ...interface{}) { l.with(ctx).Debug(args...) }
func (l *traceLogger) CtxInfo(ctx context.Context, args ...interface{})  { l.with(ctx).Info(args...) }
func (l *traceLogger) CtxWarn(ctx context.Context, args ...interface{})  { l.with(ctx).Warn(args...) }
func (l *traceLogger) CtxError(ctx context.Context, args ...interface{}) { l.with(ctx).Error(args...) }

func (l *traceLogger) CtxDebugf(ctx context.Context, fmt string, args ...interface{}) {
	l.with(ctx).Debugf(fmt, args...)
}

func (l *traceLogger) CtxInfof(ctx context.Context, fmt string, args ...interface{}) {
	l.with(ctx).Infof(fmt, args...)
}

func (l *traceLogger) CtxWarnf(ctx context.Context, fmt string, args ...interface{}) {
	l.with(ctx).Warnf(fmt, args...)
}

func (l *traceLogger) CtxErrorf(ctx context.Context, fmt string, args ...interface{}) {
	l.with(ctx).Errorf(fmt, args...)
}

// skippingTraceLogger is a traceLogger that also implements CallerSkipper.
type skippingTraceLogger struct {
	traceLogger
}

func (l *skippingTraceLogger) CallerSkipped() CtxLogger {
	sugar := l.Logger.(*zap.SugaredLogger)
	return &skippingTraceLogger{traceLogger{&DubboLogger{
		Logger:       sugar.Desugar().WithOptions(zap.AddCallerSkip(1)).Sugar(),
		DynamicLevel: l.DynamicLevel,
	}}}
}

var (
	_ CtxLogger     = (*traceLogger)(nil)
	_ CtxLogger     = (*skippingTraceLogger)(nil)
	_ CallerSkipper = (*skippingTraceLogger)(nil)
)

// observedBase builds a DubboLogger configured like InitLogger: it reports
// the caller and skips the frame of the package-level functions.
func observedBase(t *testing.T) (*DubboLogger, *observer.ObservedLogs) {
	core, logs := observer.New(zapcore.DebugLevel)
	sugar := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)).Sugar()

	old := GetLogger()
	t.Cleanup(func() { SetLogger(old) })

	return &DubboLogger{Logger: sugar, DynamicLevel: zap.NewAtomicLevelAt(zapcore.DebugLevel)}, logs
}

func logAll(ctx context.Context) {
	CtxDebug(ctx, "debug")
	CtxInfo(ctx, "info")
	CtxWarn(ctx, "warn")
	CtxError(ctx, "error")
	CtxDebugf(ctx, "%s", "debugf")
	CtxInfof(ctx, "%s", "infof")
	CtxWarnf(ctx, "%s", "warnf")
	CtxErrorf(ctx, "%s", "errorf")
}

var wantEntries = []struct {
	level   zapcore.Level
	message string
}{
	{zapcore.DebugLevel, "debug"},
	{zapcore.InfoLevel, "info"},
	{zapcore.WarnLevel, "warn"},
	{zapcore.ErrorLevel, "error"},
	{zapcore.DebugLevel, "debugf"},
	{zapcore.InfoLevel, "infof"},
	{zapcore.WarnLevel, "warnf"},
	{zapcore.ErrorLevel, "errorf"},
}

func assertEntries(t *testing.T, logs *observer.ObservedLogs, traceID string, callerFile string) {
	t.Helper()
	entries := logs.AllUntimed()
	require.Len(t, entries, len(wantEntries))
	for i, want := range wantEntries {
		got := entries[i]
		assert.Equal(t, want.level, got.Level, want.message)
		assert.Equal(t, want.message, got.Message)
		assert.Equal(t, callerFile, filepath.Base(got.Caller.File), want.message)
		if traceID == "" {
			assert.Empty(t, got.Context, want.message)
		} else {
			assert.Equal(t, traceID, got.ContextMap()["trace_id"], want.message)
		}
	}
}

func TestCtxFuncsFallBackToPlainMethods(t *testing.T) {
	base, logs := observedBase(t)
	SetLogger(base)

	logAll(context.WithValue(context.Background(), traceIDKey{}, "abc"))

	assertEntries(t, logs, "", testFile)
}

func TestCtxFuncsUseCtxLogger(t *testing.T) {
	base, logs := observedBase(t)
	SetLogger(&skippingTraceLogger{traceLogger{base}})

	logAll(context.WithValue(context.Background(), traceIDKey{}, "abc"))

	assertEntries(t, logs, "abc", testFile)
}

func TestCtxFuncsWithoutCallerSkipper(t *testing.T) {
	base, logs := observedBase(t)
	SetLogger(&traceLogger{base})

	logAll(context.WithValue(context.Background(), traceIDKey{}, "abc"))

	// The context still reaches the logger. Without CallerSkipper the
	// reported caller is the package-level function.
	assertEntries(t, logs, "abc", "ctx_logging.go")
}

func TestCtxMethodsCalledDirectly(t *testing.T) {
	base, logs := observedBase(t)
	l := &skippingTraceLogger{traceLogger{base}}

	l.CtxInfof(context.WithValue(context.Background(), traceIDKey{}, "abc"), "%s", "direct")

	entries := logs.AllUntimed()
	require.Len(t, entries, 1)
	assert.Equal(t, testFile, filepath.Base(entries[0].Caller.File))
	assert.Equal(t, "abc", entries[0].ContextMap()["trace_id"])
}

func TestPlainFuncsUnchangedWithCtxLogger(t *testing.T) {
	base, logs := observedBase(t)
	SetLogger(&skippingTraceLogger{traceLogger{base}})

	Infof("%s", "plain")

	entries := logs.AllUntimed()
	require.Len(t, entries, 1)
	assert.Equal(t, "plain", entries[0].Message)
	assert.Equal(t, testFile, filepath.Base(entries[0].Caller.File))
	assert.Empty(t, entries[0].Context)
}
