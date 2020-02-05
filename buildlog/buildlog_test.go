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
			`{"t": 1580198209, "m": "test 1", "n": 0, "s": "main"}` + "\n",
			`{"t": 1580198222, "m": "test 2", "n": 1, "s": "main"}` + "\n",
		}
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

		expected := []string{
			"2020-01-28 16:56:49 +0900 JST: test 1\n",
			"2020-01-28 16:57:02 +0900 JST: test 2\n",
		}
		assert.Equal(t, strings.Join(expected, ""), writer.String())
	})
}

type compareableLog struct {
	ctx    context.Context
	file   string
	writer io.Writer
	cancel context.CancelFunc
}

func TestNew(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tmpFile, err := ioutil.TempFile("", "")
		if err != nil {
			t.Fatal(err)
		}

		writer := bytes.NewBuffer(nil)

		logger, err := New(context.Background(), tmpFile.Name(), writer)
		if err != nil {
			t.Fatal(err)
		}

		log, ok := logger.(log)
		if !ok {
			t.Fatal("Failed to convert Logger to log")
		}

		file, ok := log.file.(*os.File)
		if !ok {
			t.Fatal("Failed to convert Reader to File")
		}

		assert.Equal(t, tmpFile.Name(), file.Name())
		assert.Equal(t, writer, log.writer)
	})

}
