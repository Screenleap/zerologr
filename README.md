# Zerologr

[![Go Reference](https://pkg.go.dev/badge/github.com/go-logr/zerologr.svg)](https://pkg.go.dev/github.com/go-logr/zerologr)
![test](https://github.com/go-logr/zerologr/workflows/test/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-logr/zerologr)](https://goreportcard.com/report/github.com/go-logr/zerologr)

A [logr](https://github.com/go-logr/logr) LogSink implementation using [Zerolog](https://github.com/rs/zerolog).

## Usage

```go
import (
    "os"

    "github.com/go-logr/logr"
    "github.com/go-logr/zerologr"
    "github.com/rs/zerolog"
)

func main() {
    zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
    zerologr.Verbosity = 2
    zerologr.NameFieldName = "logger"
    zerologr.NameSeparator = "/"

    zl := zerolog.New(os.Stderr)
    zl = zl.With().Caller().Timestamp().Logger()
    var log logr.Logger = zerologr.New(&zl)

    log.Info("Logr in action!", "the answer", 42)
}
```

## Implementation Details

~~For the most part, concepts in Zerolog correspond directly with those in logr.~~

~~Levels in logr correspond to custom debug levels in Zerolog. Any given level
in logr is represents by `zerologLevel = 1 - logrLevel`.~~

~~For example `V(2)` is equivalent to Zerolog's `TraceLevel`, while `V(1)` is
equivalent to Zerolog's `DebugLevel`.~~

Only `info` and `error` levels are supported. You should control the verbosity of info through V-levels.

For example `V(2)` will output `{"level": "info", "v": 2}`.