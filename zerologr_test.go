package zerologr

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/rs/zerolog"
	"regexp"
	"strings"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogging(t *testing.T) {
	t.Parallel()

	VerbosityFieldName = ""

	tests := []struct {
		description  string
		zerologLevel zerolog.Level
		logFunc      func(log logr.Logger)
		formatter    func(interface{}) interface{}
		reportCaller bool
		defaultName  []string
		assertions   map[string]string
	}{
		{
			description: "basic logging",
			logFunc: func(log logr.Logger) {
				log.Info("hello, world")
			},
			assertions: map[string]string{
				"level":   "info",
				"message": "hello, world",
			},
		},
		{
			description: "set name once",
			logFunc: func(log logr.Logger) {
				log.WithName("main").Info("hello, world")
			},
			assertions: map[string]string{
				"level":   "info",
				"message": "hello, world",
				"logger":  "main",
			},
		},
		{
			description: "set name twice",
			logFunc: func(log logr.Logger) {
				log.WithName("main").WithName("subpackage").Info("hello, world")
			},
			assertions: map[string]string{
				"level":   "info",
				"message": "hello, world",
				"logger":  "main/subpackage",
			},
		},
		{
			description: "set name and values and name again",
			logFunc: func(log logr.Logger) {
				log.
					WithName("main").
					WithValues("k1", "v1", "k2", "v2").
					WithName("subpackage").
					Info("hello, world", "k3", "v3")
			},
			assertions: map[string]string{
				"level":   "info",
				"message": "hello, world",
				"logger":  "main/subpackage",
				"k1":      "v1",
				"k2":      "v2",
				"k3":      "v3",
			},
		},
		{
			description: "V(0) logging with info level set is shown",
			logFunc: func(log logr.Logger) {
				log.V(0).Info("hello, world")
			},
			assertions: map[string]string{
				"level":   "info",
				"message": "hello, world",
			},
		},
		{
			description: "V(2) logging with info level set is not shown",
			logFunc: func(log logr.Logger) {
				log.V(1).Info("hello, world")
				log.V(2).Info("hello, world")
			},
			assertions: nil,
		},
		{
			description:  "V(1) logging with debug level set is shown",
			zerologLevel: zerolog.DebugLevel,
			logFunc: func(log logr.Logger) {
				log.V(1).Info("hello, world")
			},
			assertions: map[string]string{
				"level":   "debug",
				"message": "hello, world",
			},
		},
		{
			description:  "V(2) logging with trace level set is shown",
			zerologLevel: zerolog.TraceLevel,
			logFunc: func(log logr.Logger) {
				log.V(2).Info("hello, world")
			},
			assertions: map[string]string{
				"level":   "trace",
				"message": "hello, world",
			},
		},
		{
			description:  "negative V-logging truncates to info",
			zerologLevel: zerolog.TraceLevel,
			logFunc: func(log logr.Logger) {
				log.V(-10).Info("hello, world")
			},
			assertions: map[string]string{
				"level":   "info",
				"message": "hello, world",
			},
		},
		{
			description:  "additive V-logging, negatives ignored",
			zerologLevel: zerolog.TraceLevel,
			logFunc: func(log logr.Logger) {
				log.V(0).V(1).V(-20).V(1).Info("hello, world")
			},
			assertions: map[string]string{
				"level":   "trace",
				"message": "hello, world",
			},
		},
		{
			description: "arguments are added while calling Info()",
			logFunc: func(log logr.Logger) {
				log.Info("hello, world", "animal", "walrus")
			},
			assertions: map[string]string{
				"level":   "info",
				"message": "hello, world",
				"animal":  "walrus",
			},
		},
		{
			description: "arguments are added after WithValues()",
			logFunc: func(log logr.Logger) {
				log.WithValues("color", "green").Info("hello, world", "animal", "walrus")
			},
			assertions: map[string]string{
				"level":   "info",
				"message": "hello, world",
				"animal":  "walrus",
				"color":   "green",
			},
		},
		{
			description: "error logs have the appropriate information",
			logFunc: func(log logr.Logger) {
				log.Error(errors.New("this is error"), "error occurred")
			},
			assertions: map[string]string{
				"level":   "error",
				"message": "error occurred",
				"error":   "this is error",
			},
		},
		{
			description: "error shown with lov severity logger",
			logFunc: func(log logr.Logger) {
				log.Error(errors.New("this is error"), "error occurred")
			},
			assertions: map[string]string{
				"level":   "error",
				"message": "error occurred",
				"error":   "this is error",
			},
		},
		{
			description: "bad number of arguments discards all",
			logFunc: func(log logr.Logger) {
				log.Info("hello, world", "animal", "walrus", "foo")
			},
			assertions: map[string]string{
				"level":   "info",
				"message": "hello, world",
				"-animal": "walrus",
			},
		},
		{
			description: "with default name",
			logFunc: func(log logr.Logger) {
				log.Info("hello, world")
			},
			defaultName: []string{"some", "name"},
			assertions: map[string]string{
				"level":   "info",
				"message": "hello, world",
				"logger":  "some/name",
			},
		},
		{
			description: "without report caller",
			logFunc: func(log logr.Logger) {
				log.Info("hello, world")
			},
			assertions: map[string]string{
				"level":   "info",
				"message": "hello, world",
				"-caller": "no-caller",
			},
		},
		{
			description: "with report caller",
			logFunc: func(log logr.Logger) {
				log.Info("hello, world")
			},
			reportCaller: true,
			assertions: map[string]string{
				"level":   "info",
				"message": "hello, world",
				"caller":  `~zerologr_test.go:\d+`,
			},
		},
		{
			description: "with report caller and depth",
			logFunc: func(log logr.Logger) {
				log.WithCallDepth(2).Info("hello, world")
			},
			reportCaller: true,
			assertions: map[string]string{
				"level":   "info",
				"message": "hello, world",
				"caller":  `~testing.go:\d+`,
			},
		},
	}

	for _, test := range tests {
		tt := test

		t.Run(tt.description, func(t *testing.T) {
			t.Parallel()

			// Use a buffer for our output.
			logWriter := &bytes.Buffer{}

			zerologLogger := zerolog.New(logWriter)

			if tt.zerologLevel != zerolog.PanicLevel {
				zerologLogger.Level(tt.zerologLevel)
			}

			// Send the created logger to the test case to invoke desired
			// logging.
			if tt.reportCaller {
				zerologLogger = zerologLogger.With().Caller().Logger()
			}

			if tt.assertions == nil {
				assert.Equal(t, logWriter.Len(), 0)
				return
			}

			logger := New(&zerologLogger)

			if tt.defaultName != nil {
				logger = logger.WithName(strings.Join(tt.defaultName, NameSeparator))
			}

			tt.logFunc(logger)

			var loggedLine map[string]string
			b := logWriter.Bytes()
			err := json.Unmarshal(b, &loggedLine)

			require.NoError(t, err)

			for k, v := range tt.assertions {
				field, ok := loggedLine[k]

				// Annotate negative tests with a minus. To ensure `key` is
				// *not* in the output, name the assertion `-key`.
				if strings.HasPrefix(k, "-") {
					assert.False(t, ok)
					assert.Empty(t, field)

					continue
				}

				// Annotate regexp matches with the value starting with a tilde
				// (~). The tilde will be dropped and used to compile a regexp to
				// match the field.
				if strings.HasPrefix(v, "~") {
					assert.Regexp(t, regexp.MustCompile(v[1:]), field)
					continue
				}

				assert.True(t, ok)
				assert.NotEmpty(t, field)
				assert.Equal(t, v, field)
			}
		})
	}
}

func TestLogSink_Enabled(t *testing.T) {
	t.Parallel()

	type fields struct {
		verbosity zerolog.Level
	}

	type args struct {
		level int
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name:   "output level 0 is enabled with verbosity info",
			fields: fields{verbosity: zerolog.InfoLevel},
			args:   args{level: 0},
			want:   true,
		},
		{
			name:   "output level 1 is not enabled with verbosity info",
			fields: fields{verbosity: zerolog.InfoLevel},
			args:   args{level: 1},
			want:   false,
		},
		{
			name:   "output level 0 is enabled with verbosity debug",
			fields: fields{verbosity: zerolog.DebugLevel},
			args:   args{level: 0},
			want:   true,
		},
		{
			name:   "output level 1 is enabled with verbosity debug",
			fields: fields{verbosity: zerolog.DebugLevel},
			args:   args{level: 1},
			want:   true,
		},
		{
			name:   "output level 2 is not enabled with verbosity debug",
			fields: fields{verbosity: zerolog.DebugLevel},
			args:   args{level: 2},
			want:   false,
		},
		{
			name:   "output level 0 is enabled with verbosity trace",
			fields: fields{verbosity: zerolog.TraceLevel},
			args:   args{level: 0},
			want:   true,
		},
		{
			name:   "output level 1 is enabled with verbosity trace",
			fields: fields{verbosity: zerolog.TraceLevel},
			args:   args{level: 1},
			want:   true,
		},
		{
			name:   "output level 2 is enabled with verbosity trace",
			fields: fields{verbosity: zerolog.TraceLevel},
			args:   args{level: 2},
			want:   true,
		},
		{
			name:   "output level 3 is not enabled with verbosity trace",
			fields: fields{verbosity: zerolog.TraceLevel},
			args:   args{level: 3},
			want:   false,
		},
	}

	for _, test := range tests {
		tt := test

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger := zerolog.New(nil).Level(tt.fields.verbosity)

			ls := &LogSink{
				l: &logger,
			}

			assert.Equalf(t, tt.want, ls.Enabled(tt.args.level), "Enabled(%v)", tt.args.level)
		})
	}
}

func Test_zerologLevel(t *testing.T) {
	t.Parallel()

	type args struct {
		level int
	}

	tests := []struct {
		name string
		args args
		want zerolog.Level
	}{
		{
			name: "info",
			args: args{level: 0},
			want: zerolog.InfoLevel,
		},
		{
			name: "debug",
			args: args{level: 1},
			want: zerolog.DebugLevel,
		},
		{
			name: "trace",
			args: args{level: 2},
			want: zerolog.TraceLevel,
		},
		{
			name: "beyond",
			args: args{level: 3},
			want: zerolog.Level(-2),
		},
	}

	for _, test := range tests {
		tt := test

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equalf(t, tt.want, zerologLevel(tt.args.level), "zerologLevel(%v)", tt.args.level)
		})
	}
}
