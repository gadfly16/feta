package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/otiai10/copy"
)

type testCase struct {
	name    string
	command string
	want    string
}

func initTest(t *testing.T) {
	err := os.RemoveAll("/tmp/feta_test_tree")
	if err != nil {
		t.Fatalf("Couldn't remove test dir: %s", err)
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Couldn't get working dir: %s", err)
	}
	err = copy.Copy(wd+"/feta_test_tree", "/tmp/feta_test_tree")
	if err != nil {
		t.Fatalf("Couldn't copy test dir: %s", err)
	}
	err = os.Chdir("/tmp/feta_test_tree")
	if err != nil {
		t.Fatalf("Couldn't cd to test dir: %s", err)
	}
}

func toString(b io.Writer) string {
	return b.(*bytes.Buffer).String()
}

func toArgs(command string) []string {
	return strings.Split("feta -u -S /tmp/feta_test_tree "+command, " ")
}

func TestSelectors(t *testing.T) {
	initTest(t)
	tests := []testCase{
		{
			name:    "Simple name",
			command: "get dir_a",
			want:    `[{"Obj":"/dir_a/"}]`,
		},
		{
			name:    "Simple multi",
			command: "get *",
			want:    `[{"Obj":"/dir_a/"},{"Obj":"/file_a"}]`,
		},
		{
			name:    "All meta",
			command: "get @",
			want:    `[{"Obj":"/","Result":{"data":{"subdata_a":12,"subdata_b":"thing"}}}]`,
		},
		{
			name:    "All meta from file below",
			command: "get dir_a/file_b@",
			want:    `[{"Obj":"/dir_a/file_b","Result":{"User":"Bob"}}]`,
		},
		{
			name:    "Invalid relative reference",
			command: "get ..",
			want:    `[{"Error":"Invalid relative reference from /"}]`,
		},
		{
			name:    "Precedence order",
			command: "get @1+2*3*(4+5)-data.subdata_a-1==42&&data.subdata_b==\"thing\"",
			want:    `[{"Obj":"/","Result":true}]`,
		},
		{
			name:    "Relative site path",
			command: "-S /tmp/feta_test_tree/../feta_test_tree get file_a@",
			want:    `[{"Obj":"/file_a","Result":{"User":"Alice"}}]`,
		},
		{
			name:    "Recurse from root",
			command: "get /**/file*@User",
			want:    `[{"Obj":"/dir_a/file_b","Result":"Bob"},{"Obj":"/file_a","Result":"Alice"}]`,
		},
		{
			name:    "Relative",
			command: "get dir_a/../file_a",
			want:    `[{"Obj":"/file_a"}]`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			os.Args = toArgs(tc.command)
			out = bytes.NewBuffer(nil)
			main()
			if got := toString(out); got != tc.want+"\n" {
				t.Errorf("Want: %s  Got: %s", tc.want, got)
			}
		})
	}
}
