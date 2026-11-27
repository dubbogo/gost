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
)

// CtxLogger is a Logger that also takes the context of the call, for example
// to add trace fields to each entry. Set it with SetLogger like any Logger.
// When the global logger is a CtxLogger, the package-level Ctx functions call
// its Ctx methods. Otherwise they drop the context and call the plain method.
type CtxLogger interface {
	Logger

	CtxDebug(ctx context.Context, args ...interface{})
	CtxInfo(ctx context.Context, args ...interface{})
	CtxWarn(ctx context.Context, args ...interface{})
	CtxError(ctx context.Context, args ...interface{})

	CtxDebugf(ctx context.Context, fmt string, args ...interface{})
	CtxInfof(ctx context.Context, fmt string, args ...interface{})
	CtxWarnf(ctx context.Context, fmt string, args ...interface{})
	CtxErrorf(ctx context.Context, fmt string, args ...interface{})
}

// CallerSkipper is implemented by a CtxLogger that reports the caller of each
// entry. A package-level Ctx function adds one stack frame between its call
// site and the CtxLogger method, so it logs through CallerSkipped, which
// returns the logger with one more frame skipped. The reported caller is then
// the call site of the Ctx function. Calling the Ctx methods directly on the
// CtxLogger keeps working as before.
type CallerSkipper interface {
	CallerSkipped() CtxLogger
}

// ctxLogger returns the global logger as a CtxLogger, adjusted for the extra
// stack frame of the package-level Ctx functions.
func ctxLogger() (CtxLogger, bool) {
	l, ok := GetLogger().(CtxLogger)
	if !ok {
		return nil, false
	}
	if s, ok := l.(CallerSkipper); ok {
		return s.CallerSkipped(), true
	}
	return l, true
}

// CtxDebug logs at debug level with the context of the call.
func CtxDebug(ctx context.Context, args ...interface{}) {
	if l, ok := ctxLogger(); ok {
		l.CtxDebug(ctx, args...)
		return
	}
	GetLogger().Debug(args...)
}

// CtxInfo logs at info level with the context of the call.
func CtxInfo(ctx context.Context, args ...interface{}) {
	if l, ok := ctxLogger(); ok {
		l.CtxInfo(ctx, args...)
		return
	}
	GetLogger().Info(args...)
}

// CtxWarn logs at warning level with the context of the call.
func CtxWarn(ctx context.Context, args ...interface{}) {
	if l, ok := ctxLogger(); ok {
		l.CtxWarn(ctx, args...)
		return
	}
	GetLogger().Warn(args...)
}

// CtxError logs at error level with the context of the call.
func CtxError(ctx context.Context, args ...interface{}) {
	if l, ok := ctxLogger(); ok {
		l.CtxError(ctx, args...)
		return
	}
	GetLogger().Error(args...)
}

// CtxDebugf logs a templated message at debug level with the context of the call.
func CtxDebugf(ctx context.Context, fmt string, args ...interface{}) {
	if l, ok := ctxLogger(); ok {
		l.CtxDebugf(ctx, fmt, args...)
		return
	}
	GetLogger().Debugf(fmt, args...)
}

// CtxInfof logs a templated message at info level with the context of the call.
func CtxInfof(ctx context.Context, fmt string, args ...interface{}) {
	if l, ok := ctxLogger(); ok {
		l.CtxInfof(ctx, fmt, args...)
		return
	}
	GetLogger().Infof(fmt, args...)
}

// CtxWarnf logs a templated message at warning level with the context of the call.
func CtxWarnf(ctx context.Context, fmt string, args ...interface{}) {
	if l, ok := ctxLogger(); ok {
		l.CtxWarnf(ctx, fmt, args...)
		return
	}
	GetLogger().Warnf(fmt, args...)
}

// CtxErrorf logs a templated message at error level with the context of the call.
func CtxErrorf(ctx context.Context, fmt string, args ...interface{}) {
	if l, ok := ctxLogger(); ok {
		l.CtxErrorf(ctx, fmt, args...)
		return
	}
	GetLogger().Errorf(fmt, args...)
}
