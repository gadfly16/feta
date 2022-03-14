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

func teardownTest(t *testing.T) {
	err := os.RemoveAll("/tmp/feta_test_tree")
	if err != nil {
		t.Fatalf("Couldn't remove test dir: %s", err)
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
			name:    "Invalid relative reference",
			command: "get ..",
			want:    `[{"Error":"Invalid relative reference from /"}]`,
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
	teardownTest(t)
}
