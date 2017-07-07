// Copyright 2011 Google Inc. All Rights Reserved.
// This file is available under the Apache license.

package vm

import (
	"regexp"
	"testing"
	"time"

	"github.com/go-test/deep"
	"github.com/google/mtail/metrics"
)

var instructions = []struct {
	name          string
	i             instr
	re            []*regexp.Regexp
	str           []string
	reversedStack []interface{} // stack is inverted to be pushed onto vm stack

	expectedStack  []interface{}
	expectedThread thread
}{
	// Composite literals require too many explicit conversions.
	{"inc",
		instr{inc, nil},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{0},
		[]interface{}{},
		thread{pc: 0, matches: map[int][]string{}},
	},
	{"inc by int",
		instr{inc, 2},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{0, 1}, // first is metric 0 "foo", second is the inc val.
		[]interface{}{},
		thread{pc: 0, matches: map[int][]string{}},
	},
	{"inc by string",
		instr{inc, 2},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{0, "1"}, // first is metric 0 "foo", second is the inc val.
		[]interface{}{},
		thread{pc: 0, matches: map[int][]string{}},
	},
	{"set int",
		instr{iset, nil},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{1, 2}, // set metric 1 "bar"
		[]interface{}{},
		thread{pc: 0, matches: map[int][]string{}},
	},
	{"set str",
		instr{iset, nil},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{1, "2"},
		[]interface{}{},
		thread{pc: 0, matches: map[int][]string{}},
	},
	{"match",
		instr{match, 0},
		[]*regexp.Regexp{regexp.MustCompile("a*b")},
		[]string{},
		[]interface{}{},
		[]interface{}{},
		thread{match: true, pc: 0, matches: map[int][]string{0: {"aaaab"}}},
	},
	{"cmp lt",
		instr{cmp, -1},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{1, "2"},
		[]interface{}{},
		thread{pc: 0, match: true, matches: map[int][]string{}}},
	{"cmp eq",
		instr{cmp, 0},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{"2", "2"},
		[]interface{}{},
		thread{pc: 0, match: true, matches: map[int][]string{}}},
	{"cmp gt",
		instr{cmp, 1},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{2, 1},
		[]interface{}{},
		thread{pc: 0, match: true, matches: map[int][]string{}}},
	{"cmp le",
		instr{cmp, 1},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{2, "2"},
		[]interface{}{},
		thread{pc: 0, match: false, matches: map[int][]string{}}},
	{"cmp ne",
		instr{cmp, 0},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{"1", "2"},
		[]interface{}{},
		thread{pc: 0, match: false, matches: map[int][]string{}}},
	{"cmp ge",
		instr{cmp, -1},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{2, 2},
		[]interface{}{},
		thread{pc: 0, match: false, matches: map[int][]string{}}},
	{"cmp gt float float",
		instr{cmp, 1},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{"2.0", "1.0"},
		[]interface{}{},
		thread{pc: 0, match: true, matches: map[int][]string{}}},
	{"cmp gt float int",
		instr{cmp, 1},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{"1.0", "2"},
		[]interface{}{},
		thread{pc: 0, match: false, matches: map[int][]string{}}},
	{"cmp gt int float",
		instr{cmp, 1},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{"1", "2.0"},
		[]interface{}{},
		thread{pc: 0, match: false, matches: map[int][]string{}}},
	{"cmp eq string string false",
		instr{cmp, 0},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{"abc", "def"},
		[]interface{}{},
		thread{pc: 0, match: false, matches: map[int][]string{}}},
	{"cmp eq string string true",
		instr{cmp, 0},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{"abc", "abc"},
		[]interface{}{},
		thread{pc: 0, match: true, matches: map[int][]string{}}},
	{"jnm",
		instr{jnm, 37},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{},
		[]interface{}{},
		thread{pc: 37, matches: map[int][]string{}}},
	{"jm",
		instr{jm, 37},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{},
		[]interface{}{},
		thread{pc: 0, matches: map[int][]string{}}},
	{"jmp",
		instr{jmp, 37},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{},
		[]interface{}{},
		thread{pc: 37, matches: map[int][]string{}}},
	{"strptime",
		instr{strptime, 0},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{"2012/01/18 06:25:00", "2006/01/02 15:04:05"},
		[]interface{}{},
		thread{pc: 0, time: time.Date(2012, 1, 18, 6, 25, 0, 0, time.UTC),
			matches: map[int][]string{}}},
	{"iadd",
		instr{iadd, 0},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{2, 1},
		[]interface{}{int64(3)},
		thread{pc: 0, matches: map[int][]string{}}},
	{"isub",
		instr{isub, 0},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{2, 1},
		[]interface{}{int64(1)},
		thread{pc: 0, matches: map[int][]string{}}},
	{"imul",
		instr{imul, 0},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{2, 1},
		[]interface{}{int64(2)},
		thread{pc: 0, matches: map[int][]string{}}},
	{"idiv",
		instr{idiv, 0},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{4, 2},
		[]interface{}{int64(2)},
		thread{pc: 0, matches: map[int][]string{}}},
	{"imod",
		instr{imod, 0},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{4, 2},
		[]interface{}{int64(0)},
		thread{pc: 0, matches: map[int][]string{}}},
	{"imod 2",
		instr{imod, 0},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{3, 2},
		[]interface{}{int64(1)},
		thread{pc: 0, matches: map[int][]string{}}},
	{"tolower",
		instr{tolower, 0},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{"mIxeDCasE"},
		[]interface{}{"mixedcase"},
		thread{pc: 0, matches: map[int][]string{}}},
	{"length",
		instr{length, 0},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{"1234"},
		[]interface{}{4},
		thread{pc: 0, matches: map[int][]string{}}},
	{"length 0",
		instr{length, 0},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{""},
		[]interface{}{0},
		thread{pc: 0, matches: map[int][]string{}}},
	{"shl",
		instr{shl, 0},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{2, 1},
		[]interface{}{int64(4)},
		thread{pc: 0, matches: map[int][]string{}}},
	{"shr",
		instr{shr, 0},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{2, 1},
		[]interface{}{int64(1)},
		thread{pc: 0, matches: map[int][]string{}}},
	{"and",
		instr{and, 0},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{2, 1},
		[]interface{}{int64(0)},
		thread{pc: 0, matches: map[int][]string{}}},
	{"or",
		instr{or, 0},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{2, 1},
		[]interface{}{int64(3)},
		thread{pc: 0, matches: map[int][]string{}}},
	{"xor",
		instr{xor, 0},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{2, 1},
		[]interface{}{int64(3)},
		thread{pc: 0, matches: map[int][]string{}}},
	{"xor 2",
		instr{xor, 0},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{2, 3},
		[]interface{}{int64(1)},
		thread{pc: 0, matches: map[int][]string{}}},
	{"xor 3",
		instr{xor, 0},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{-1, 3},
		[]interface{}{int64(^3)},
		thread{pc: 0, matches: map[int][]string{}}},
	{"not",
		instr{not, 0},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{0},
		[]interface{}{int64(-1)},
		thread{pc: 0, matches: map[int][]string{}}},
	{"pow",
		instr{ipow, 0},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{2, 2},
		[]interface{}{int64(4)},
		thread{pc: 0, matches: map[int][]string{}}},
	{"strtol",
		instr{strtol, 2},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{"ff", 16},
		[]interface{}{int64(255)},
		thread{pc: 0, matches: map[int][]string{}}},
	{"settime",
		instr{settime, 0},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{int64(0)},
		[]interface{}{},
		thread{pc: 0, time: time.Unix(0, 0).UTC(), matches: map[int][]string{}}},
	{"push int",
		instr{push, 1},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{},
		[]interface{}{1},
		thread{pc: 0, matches: map[int][]string{}}},
	{"push float",
		instr{push, 1.0},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{},
		[]interface{}{1.0},
		thread{pc: 0, matches: map[int][]string{}}},
	{"setmatched false",
		instr{setmatched, false},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{},
		[]interface{}{},
		thread{matched: false, pc: 0, matches: map[int][]string{}}},
	{"setmatched true",
		instr{setmatched, true},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{},
		[]interface{}{},
		thread{matched: true, pc: 0, matches: map[int][]string{}}},
	{"otherwise",
		instr{otherwise, nil},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{},
		[]interface{}{},
		thread{match: true, pc: 0, matches: map[int][]string{}}},
	{"fadd",
		instr{fadd, nil},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{1.0, 2.0},
		[]interface{}{3.0},
		thread{match: false, pc: 0, matches: map[int][]string{}}},
	{"fsub",
		instr{fsub, nil},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{1.0, 2.0},
		[]interface{}{-1.0},
		thread{pc: 0, matches: map[int][]string{}}},
	{"fmul",
		instr{fmul, nil},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{1.0, 2.0},
		[]interface{}{2.0},
		thread{pc: 0, matches: map[int][]string{}}},
	{"fdiv",
		instr{fdiv, nil},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{1.0, 2.0},
		[]interface{}{0.5},
		thread{pc: 0, matches: map[int][]string{}}},
	{"fmod",
		instr{fmod, nil},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{1.0, 2.0},
		[]interface{}{1.0},
		thread{pc: 0, matches: map[int][]string{}}},
	{"fpow",
		instr{fpow, nil},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{2.0, 2.0},
		[]interface{}{4.0},
		thread{pc: 0, matches: map[int][]string{}}},
	{"fset",
		instr{fset, nil},
		[]*regexp.Regexp{},
		[]string{},
		[]interface{}{2, 2.0}, // quux set to 2.
		[]interface{}{},
		thread{pc: 0, matches: map[int][]string{}}},
}

// TestInstrs tests that each instruction behaves as expected through one
// instruction cycle.
func TestInstrs(t *testing.T) {
	for _, tc := range instructions {
		var m []*metrics.Metric
		m = append(m,
			metrics.NewMetric("foo", "test", metrics.Counter, metrics.Int),
			metrics.NewMetric("bar", "test", metrics.Counter, metrics.Int),
			metrics.NewMetric("quux", "test", metrics.Gauge, metrics.Float))
		obj := &object{re: tc.re, str: tc.str, m: m, prog: []instr{tc.i}}
		v := New(tc.name, obj, true)
		v.t = new(thread)
		v.t.stack = make([]interface{}, 0)
		for _, item := range tc.reversedStack {
			v.t.Push(item)
		}
		v.t.matches = make(map[int][]string, 0)
		v.input = "aaaab"
		v.execute(v.t, tc.i)
		if v.terminate {
			t.Fatalf("Execution failed, see info log.")
		}

		if diff := deep.Equal(tc.expectedStack, v.t.stack); diff != nil {
			t.Errorf("%s: unexpected virtual machine stack state.\n%s", tc.name, diff)
		}
		// patch in the thread stack because otherwise the test table is huge
		tc.expectedThread.stack = tc.expectedStack

		if diff := deep.Equal(v.t, &tc.expectedThread); diff != nil {
			t.Errorf("%s: unexpected virtual machine thread state.\n%s", tc.name, diff)
		}
	}
}
