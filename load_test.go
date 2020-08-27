package main

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"testing"
)

func TestLoadVarfile(t *testing.T) {
	want := map[string]interface{}{
		"int":       100,
		"str":       "a string",
		"array_int": []interface{}{1, 2, 3},
		"struct": map[string]interface{}{
			"int_field": 200,
			"str_field": "a string",
		},
	}

	for _, f := range []string{
		"./testfiles/varfile_testdata.json",
		"./testfiles/varfile_testdata.toml",
	} {
		got, err := loadVarfile(f)
		if err != nil {
			t.Fatalf("loadVarfile(%s) error: %s", f, err)
		}

		gotStr := fmt.Sprintf("%#v", got)
		wantStr := fmt.Sprintf("%#v", want)
		if gotStr != wantStr {
			t.Errorf("loadVarfile(%s) returns error\n\twant: %s\n\tgot:  %s", f, wantStr, gotStr)
		}
	}
}

func TestLoadTemplates(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"./testfiles/templates/file1", "file1"},
		{"./testfiles/templates/etc/file3", "file3"},
		{"./testfiles/templates", "etc/file3,etc/verify,file1,file2"},
		{"./testfiles/templates/", "etc/file3,etc/verify,file1,file2"},
	}

	for _, test := range tests {
		tmpl, err := loadTemplates(test.input)
		if err != nil {
			t.Fatalf("load templates %s error: %s", test.input, err)
		}

		tmpls := tmpl.Templates()
		names := make([]string, 0, len(tmpls))
		for _, t := range tmpls {
			names = append(names, t.Name())
		}
		sort.Strings(names)
		got := strings.Join(names, ",")
		if got != test.want {
			t.Errorf("loadTemplates(%s) returns error\n\twant: %s\n\tgot:  %s", test.input, test.want, got)
		}
	}

	var buf bytes.Buffer
	tmpls, _ := loadTemplates("./testfiles/templates/") // already tested, should be no error
	if err := tmpls.Lookup("etc/verify").Execute(&buf, nil); err != nil {
		t.Fatalf("execute verify error: %s", err)
	}
	if want, got := "verify\n", buf.String(); want != got {
		t.Errorf("execute verify:\n\twant: %s\n\tgot:  %s", want, got)
	}
}
