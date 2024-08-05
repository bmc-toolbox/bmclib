package executor

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func Test_Stdin(t *testing.T) {
	e := new(Execute)
	e.Cmd = "grep"
	e.Args = []string{"hello"}
	e.Stdin = bytes.NewReader([]byte("hello"))
	e.SetQuiet()

	result, err := e.ExecWithContext(context.Background())
	if err != nil {
		fmt.Println(err.Error())
	}

	assert.Equal(t, []byte("hello\n"), result.Stdout)
}

type checkBinTester struct {
	createFile  bool
	filePath    string
	expectedErr error
	fileMode    uint
	testName    string
}

func initCheckBinTests() []checkBinTester {
	return []checkBinTester{
		{
			false,
			"f",
			ErrBinLookupPath,
			0,
			"bin path lookup err test",
		},
		{
			false,
			"/tmp/f",
			ErrBinLstat,
			0,
			"bin exists err test",
		},
		{
			true,
			"/tmp/f",
			ErrBinNotExecutable,
			0o666,
			"bin exists with no executable bit test",
		},
		{
			true,
			"/tmp/j",
			nil,
			0o667,
			"bin with executable bit returns no error",
		},
		{
			true,
			"/tmp/k",
			nil,
			0o700,
			"bin with owner executable bit returns no error",
		},
		{
			true,
			"/tmp/l",
			nil,
			0o070,
			"bin with group executable bit returns no error",
		},
		{
			true,
			"/tmp/m",
			nil,
			0o007,
			"bin with other executable bit returns no error",
		},
	}
}

func Test_CheckExecutable(t *testing.T) {
	tests := initCheckBinTests()
	for _, c := range tests {
		if c.createFile {
			f, err := os.Create(c.filePath)
			if err != nil {
				t.Error(err)
			}

			// nolint:gocritic // test code
			defer os.Remove(c.filePath)

			if c.fileMode != 0 {
				err = f.Chmod(fs.FileMode(c.fileMode))
				if err != nil {
					t.Error(err)
				}
			}
		}

		e := new(Execute)
		e.Cmd = c.filePath
		err := e.CheckExecutable()
		assert.Equal(t, c.expectedErr, errors.Cause(err), c.testName)
	}
}
