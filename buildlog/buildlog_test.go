package buildlog

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	intervalTime = 500
)

var (
	testInputs = []string{
		`{"t": 1581662022394, "m": "test 1", "n": 0, "s": "main"}` + "\n",
		`{"t": 1581662022395, "m": "test 2", "n": 1, "s": "main"}` + "\n",
	}
)

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
		defer tmpFile.Close()

		go write(t, tmpFile.Name(), testInputs)

		parent, cancel := context.WithCancel(context.Background())
		writer := bytes.NewBuffer(nil)
		l := log{
			file:   tmpFile,
			writer: writer,
			ctx:    parent,
			cancel: cancel,
		}

		go l.Run()

		time.Sleep(intervalTime * time.Millisecond)
		l.Stop()

		expected := "main: test 1\nmain: test 2\n"
		assert.Equal(t, expected, writer.String())
	})

	t.Run("failure by parsing error", func(t *testing.T) {
		tmpFile, err := ioutil.TempFile("", "")
		if err != nil {
			t.Fatal(err)
		}
		defer tmpFile.Close()

		testInvalidInputs := []string{
			`{"t": 1581662022394, "m": "test 1", "n": 0, "s": "main"}` + "\n",
			`{` + "\n",
		}
		go write(t, tmpFile.Name(), testInvalidInputs)

		parent, cancel := context.WithCancel(context.Background())
		writer := bytes.NewBuffer(nil)
		l := log{
			file:   tmpFile,
			writer: writer,
			ctx:    parent,
			cancel: cancel,
		}

		go l.Run()

		time.Sleep(intervalTime * time.Millisecond)
		l.Stop()

		expected := "main: test 1\n"
		assert.Equal(t, expected, writer.String())
	})
}

func TestStop(t *testing.T) {
	t.Run("success, confirm not to write log after stopped", func(t *testing.T) {
		tmpFile, err := ioutil.TempFile("", "")
		if err != nil {
			t.Fatal(err)
		}
		defer tmpFile.Close()

		testInputsNotWritten := []string{
			`{test}` + "\n",
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

		go l.Run()

		time.Sleep(intervalTime * time.Millisecond)
		l.Stop()

		go write(t, tmpFile.Name(), testInputsNotWritten)
		time.Sleep(intervalTime * time.Millisecond)

		expected := "main: test 1\nmain: test 2\n"
		assert.Equal(t, expected, writer.String())
	})
}

func TestNew(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tmpFile, err := ioutil.TempFile("", "")
		if err != nil {
			t.Fatal(err)
		}
		defer tmpFile.Close()

		writer := bytes.NewBuffer(nil)

		logger, err := New(tmpFile.Name(), writer)
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

	t.Run("failure", func(t *testing.T) {
		writer := bytes.NewBuffer(nil)

		logger, err := New("/", writer)
		if err == nil {
			t.Fatal("failure err is nil")
		}

		expected := log{
			writer: writer,
			file:   (*os.File)(nil),
		}

		msg := err.Error()
		assert.Equal(t, expected, logger)
		assert.Equal(t, 0, strings.Index(msg, "failed to open raw build log file: "), fmt.Sprintf("expected error is `failed to open raw build log file: ...`, actual: `%v`", msg))
	})
}
