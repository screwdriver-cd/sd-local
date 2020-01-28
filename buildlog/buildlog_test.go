package buildlog

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type writeCloser struct {
	io.Writer
}

func (wc writeCloser) Close() error {
	return nil
}

func write(tb testing.TB, filepath string, inputs []string) {
	tb.Helper()

	for _, input := range inputs {
		file, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			tb.Fatal(err)
		}

		_, err = file.Write([]byte(input))
		if err != nil {
			tb.Fatal(err)
		}
		//time.Sleep(1 * time.Second)

		err = file.Close()
		if err != nil {
			tb.Fatal(err)
		}
	}
}

func TestRun(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tmpFile, err := ioutil.TempFile("", "")
		if err != nil {
			t.Fatal(err)
		}

		testInputs := []string{
			"test1\n",
			"test2\n",
		}
		t.Log(tmpFile.Name())
		go write(t, tmpFile.Name(), testInputs)

		parent, cancel := context.WithCancel(context.Background())
		writer := bytes.NewBuffer(nil)
		l := log{
			file:   tmpFile,
			writer: writer,
			ctx:    parent,
			cancel: cancel,
		}

		errChan := l.Run()
		_ = errChan

		time.Sleep(3 * time.Second)
		l.Stop()

		assert.Equal(t, strings.Join(testInputs, ""), writer.String())

		//t.Log(<-errChan)
	})
}
