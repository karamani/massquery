package main

import (
	"testing"
)

func TestParametrizedString(t *testing.T) {

	cases := []struct {
		inString, inTpl string
		inParams        []string
		want            string
	}{
		{"{0}|{1}", "{%d}", []string{"param0", "param1"}, "param0|param1"},
		{"{res0}|{res1}", "{res%d}", []string{"param0", "param1"}, "param0|param1"},
		{"{res1}|{res1}", "{res%d}", []string{"param0", "param1"}, "param1|param1"},
	}

	for _, c := range cases {
		got := parameterizedString(c.inString, c.inTpl, c.inParams)
		if got != c.want {
			t.Errorf("parametrizedString(%q, %q, %q) == %q, want %q", c.inString, c.inTpl, c.inParams, got, c.want)
		}
	}
}

func TestFormatRes(t *testing.T) {

	cases := []struct {
		inFormat, inInput, inCnn, inStatus string
		inValues                           []string
		want                               string
	}{
		{"", "", "", "success", []string{"param0", "param1"}, "param0\tparam1"},
		{"{status}|{res}", "", "", "success", []string{"param0", "param1"}, "success|param0\tparam1"},
		{"{res2};{res1};{res0}", "", "", "", []string{"param0", "param1"}, "{res2};param1;param0"},
		{"{input}|{res}|{cnn}|{status}|{res0}", "i", "cnn", "success", []string{"r0", "r1"}, "i|r0\tr1|cnn|success|r0"},
	}

	for _, c := range cases {
		got := formatRes(c.inFormat, c.inInput, c.inCnn, c.inStatus, c.inValues)
		if got != c.want {
			t.Errorf("formatRes(%q, %q, %q, %q, %q) == %q, want %q",
				c.inFormat, c.inInput, c.inCnn, c.inStatus, c.inValues, got, c.want)
		}
	}
}
