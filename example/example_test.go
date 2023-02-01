//go:build !binary_log
// +build !binary_log

/*
Copyright 2021 The logr Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package zerologr_test

import (
	"os"

	"github.com/go-logr/logr"
	"github.com/go-logr/zerologr"
	"github.com/rs/zerolog"
)

type E struct {
	str string
}

func (e E) Error() string {
	return e.str
}

func Helper(log logr.Logger, msg string) {
	helper2(log, msg)
}

func helper2(log logr.Logger, msg string) {
	log.WithCallDepth(2).Info(msg)
}

func ExampleNew() {
	zerologr.SetMaxV(1)
	zl := zerolog.New(os.Stdout)
	log := zerologr.New(&zl)
	log = log.WithName("MyName")
	log = log.WithValues("module", "example")

	log.Info("hello", "val1", 1, "val2", map[string]int{"k": 1})
	log.V(1).Info("you should see this")
	log.V(1).V(1).Info("you should NOT see this")
	log.Error(nil, "uh oh", "trouble", true, "reasons", []float64{0.1, 0.11, 3.14})
	log.Error(E{"an error occurred"}, "goodbye", "code", -1)
	Helper(log, "thru a helper")

	// Output:
	// {"level":"info","module":"example","v":0,"logger":"MyName","val1":1,"val2":{"k":1},"message":"hello"}
	// {"level":"debug","module":"example","v":1,"logger":"MyName","message":"you should see this"}
	// {"level":"error","module":"example","logger":"MyName","trouble":true,"reasons":[0.1,0.11,3.14],"message":"uh oh"}
	// {"level":"error","module":"example","error":"an error occurred","logger":"MyName","code":-1,"message":"goodbye"}
	// {"level":"info","module":"example","v":0,"logger":"MyName","message":"thru a helper"}
}
