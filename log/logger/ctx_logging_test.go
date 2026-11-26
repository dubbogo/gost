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
	"runtime"
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

// nextLine returns the line number of the statement after the call.
func nextLine() int {
	_, _, line, _ := runtime.Caller(1)
	return line + 1
}

// logAll calls each package-level Ctx function and returns their lines.
func logAll(ctx context.Context) []int {
	var lines []int
	lines = append(lines, nextLine())
	CtxDebug(ctx, "debug")
	lines = append(lines, nextLine())
	CtxInfo(ctx, "info")
	lines = append(lines, nextLine())
	CtxWarn(ctx, "warn")
	lines = append(lines, nextLine())
	CtxError(ctx, "error")
	lines = append(lines, nextLine())
	CtxDebugf(ctx, "%s", "debugf")
	lines = append(lines, nextLine())
	CtxInfof(ctx, "%s", "infof")
	lines = append(lines, nextLine())
	CtxWarnf(ctx, "%s", "warnf")
	lines = append(lines, nextLine())
	CtxErrorf(ctx, "%s", "errorf")
	return lines
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

// assertEntries checks the entries logged by logAll. A nil lines means the
// caller is expected in callerFile at a line this test does not know.
func assertEntries(t *testing.T, logs *observer.ObservedLogs, traceID string, callerFile string, lines []int) {
	t.Helper()
	entries := logs.AllUntimed()
	require.Len(t, entries, len(wantEntries))
	for i, want := range wantEntries {
		got := entries[i]
		assert.Equal(t, want.level, got.Level, want.message)
		assert.Equal(t, want.message, got.Message)
		assert.Equal(t, callerFile, filepath.Base(got.Caller.File), want.message)
		if lines != nil {
			assert.Equal(t, lines[i], got.Caller.Line, want.message)
		}
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

	lines := logAll(context.WithValue(context.Background(), traceIDKey{}, "abc"))

	assertEntries(t, logs, "", testFile, lines)
}

func TestCtxFuncsUseCtxLogger(t *testing.T) {
	base, logs := observedBase(t)
	SetLogger(&skippingTraceLogger{traceLogger{base}})

	lines := logAll(context.WithValue(context.Background(), traceIDKey{}, "abc"))

	assertEntries(t, logs, "abc", testFile, lines)
}

func TestCtxFuncsWithoutCallerSkipper(t *testing.T) {
	base, logs := observedBase(t)
	SetLogger(&traceLogger{base})

	logAll(context.WithValue(context.Background(), traceIDKey{}, "abc"))

	// The context still reaches the logger. Without CallerSkipper the
	// reported caller is the package-level function.
	assertEntries(t, logs, "abc", "ctx_logging.go", nil)
}

func TestCtxMethodsCalledDirectly(t *testing.T) {
	base, logs := observedBase(t)
	l := &skippingTraceLogger{traceLogger{base}}

	line := nextLine()
	l.CtxInfof(context.WithValue(context.Background(), traceIDKey{}, "abc"), "%s", "direct")

	entries := logs.AllUntimed()
	require.Len(t, entries, 1)
	assert.Equal(t, testFile, filepath.Base(entries[0].Caller.File))
	assert.Equal(t, line, entries[0].Caller.Line)
	assert.Equal(t, "abc", entries[0].ContextMap()["trace_id"])
}

func TestPlainFuncsUnchangedWithCtxLogger(t *testing.T) {
	base, logs := observedBase(t)
	SetLogger(&skippingTraceLogger{traceLogger{base}})

	line := nextLine()
	Infof("%s", "plain")

	entries := logs.AllUntimed()
	require.Len(t, entries, 1)
	assert.Equal(t, "plain", entries[0].Message)
	assert.Equal(t, testFile, filepath.Base(entries[0].Caller.File))
	assert.Equal(t, line, entries[0].Caller.Line)
	assert.Empty(t, entries[0].Context)
}
