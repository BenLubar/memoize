package memoize_test

import (
	"testing"

	"github.com/BenLubar/memoize"
)

func TestPanic(t *testing.T) {
	count := 0
	f := func(i int) {
		count++
		if count%2 == 1 {
			panic(count)
		}
	}
	f = memoize.Memoize(f).(func(int))

	expect := func(p interface{}, i int) {
		defer func() {
			if r := recover(); p != r {
				t.Errorf("for input %d:\nexpected: %v\nactual: %v", i, p, r)
			}
		}()

		f(i)
	}

	expect(1, 1)
	expect(1, 1)
	expect(nil, 2)
	expect(nil, 2)
	expect(1, 1)
	expect(3, 100)
}

func TestVariadic(t *testing.T) {
	count := 0
	var concat func(string, ...string) string
	concat = func(s0 string, s1 ...string) string {
		count++

		if len(s1) == 0 {
			return s0
		}
		return concat(s0+s1[0], s1[1:]...)
	}
	concat = memoize.Memoize(concat).(func(string, ...string) string)

	expect := func(actual, expected string, n int) {
		if actual != expected || n != count {
			t.Errorf("expected: %q\nactual: %q\nexpected count: %d\nactual count: %d", expected, actual, n, count)
		}
	}

	expect("", "", 0)
	expect(concat("string"), "string", 1)
	expect(concat("string", "one"), "stringone", 3)
	expect(concat("string", "one"), "stringone", 3)
	expect(concat("string", "two"), "stringtwo", 5)
	expect(concat("string", "one"), "stringone", 5)
	expect(concat("stringone", "two"), "stringonetwo", 7)
	expect(concat("string", "one", "two"), "stringonetwo", 8)
}
