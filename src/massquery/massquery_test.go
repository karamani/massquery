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

func TestValidateArgs(t *testing.T) {

	var err error

	connectionStringArg = ""
	err = validateArgs()
	if err == nil || err.Error() != "'cnn' arg is required" {
		t.Errorf("validateArgs: 'cnn' is empty")
	}

	connectionStringArg = "validcnn"

	queryArg, execArg = "", ""
	err = validateArgs()
	if err == nil || err.Error() != "it should be one of the arguments 'query' or 'exec'" {
		t.Errorf("validateArgs: 'query' & 'exec' is empty")
	}

	queryArg, execArg = "validquery", ""
	err = validateArgs()
	if err != nil {
		t.Errorf("validateArgs: valid 'query' fail")
	}

	queryArg, execArg = "", "validexec"
	err = validateArgs()
	if err != nil {
		t.Errorf("validateArgs: valid 'exec' fail")
	}

	queryArg, execArg = "validquery", "validexec"
	err = validateArgs()
	if err == nil || err.Error() != "it should be only one of the arguments 'query' or 'exec'" {
		t.Errorf("validateArgs: given both args 'exec' & 'query'")
	}
}
