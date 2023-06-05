// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slog_test

import (
	"context"
	"errors"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/moontrade/kirana/logger/slog"
	"github.com/moontrade/kirana/logger/slog/internal/testutil"
)

func ExampleGroup() {
	r, _ := http.NewRequest("GET", "localhost", nil)
	// ...

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{ReplaceAttr: testutil.RemoveTime}))
	slog.SetDefault(logger)

	slog.Info("finished",
		slog.Group("req",
			slog.String("method", r.Method),
			slog.String("url", r.URL.String())),
		slog.Int("status", http.StatusOK),
		slog.Duration("duration", time.Second))

	// Output:
	// level=INFO msg=finished req.method=GET req.url=localhost status=200 duration=1s
}

func TestGroup(t *testing.T) {
	r, _ := http.NewRequest("GET", "localhost", nil)
	// ...

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{ReplaceAttr: testutil.RemoveTime}))
	slog.SetDefault(logger)

	slog.Info("finished",
		slog.Group("req",
			slog.String("method", r.Method),
			slog.String("url", r.URL.String())),
		slog.Int("status", http.StatusOK),
		slog.Duration("duration", time.Second))
}

const TestMessage = "Test logging, but use a somewhat realistic message length."

var (
	TestTime     = time.Date(2022, time.May, 1, 0, 0, 0, 0, time.UTC)
	TestString   = "7e3b3b2aaeff56a7108fe11e154200dd/7819479873059528190"
	TestInt      = 32768
	TestDuration = 23 * time.Second
	TestError    = errors.New("fail")
)

var TestAttrs = []slog.Attr{
	slog.String("string", TestString),
	slog.Int("status", TestInt),
	slog.Duration("duration", TestDuration),
	slog.Time("time", TestTime),
	slog.Any("error", TestError),
}

func Benchmark(b *testing.B) {
	logger := slog.New(disabledHandler{})
	logger.HandlerRaw = disabledHandler{}
	slog.SetDefault(logger)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		//logger.LogAttrsRaw(slog.LevelInfo, TestMessage,
		//	slog.String("string", TestString),
		//	slog.Int("status", TestInt),
		//	//slog.Duration("duration", TestDuration),
		//	//slog.Time("time", TestTime),
		//	//slog.Any("error", TestError),
		//)

		logger.LogAttrsRaw(context.Background(), slog.LevelInfo, TestMessage,
			slog.String("string", TestString),
			slog.Int("status", TestInt),
			slog.Duration("duration", TestDuration),
			slog.Time("time", TestTime),
			slog.Any("error", TestError),
		)

		//slog.BenchmarkPC()
		//logger.Info("finished",
		//	//slog.Group("req",
		//	//	slog.String("method", r.Method),
		//	//	slog.String("url", r.URL.String())),
		//	slog.Int("status", http.StatusOK),
		//	slog.Duration("duration", time.Second))
	}
}

type disabledHandler struct{}

func (disabledHandler) Enabled(context.Context, slog.Level) bool { return true }
func (disabledHandler) Handle(ctx context.Context, r slog.Record) error {
	//panic("should not be called")
	return nil
}

func (disabledHandler) HandleRaw(r slog.Record, attr ...slog.Attr) error {
	return nil
}

func (disabledHandler) WithAttrs([]slog.Attr) slog.Handler {
	panic("disabledHandler: With unimplemented")
}

func (disabledHandler) WithGroup(string) slog.Handler {
	panic("disabledHandler: WithGroup unimplemented")
}
