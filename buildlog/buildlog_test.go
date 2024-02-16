package buildlog

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
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
		tmpFile, err := os.CreateTemp("", "*")
		if err != nil {
			t.Fatal(err)
		}
		defer tmpFile.Close()

		go write(t, tmpFile.Name(), testInputs)

		parent, cancel := context.WithCancel(context.Background())
		writer := bytes.NewBuffer(nil)
		done := make(chan struct{})
		l := log{
			file:   tmpFile,
			writer: writer,
			ctx:    parent,
			cancel: cancel,
			done:   done,
		}

		go l.Run()

		time.Sleep(intervalTime * time.Millisecond)
		l.Stop()
		timeout := time.After(5 * time.Second)

		select {
		case <-done:
			expected := "main: test 1\r\nmain: test 2\r\n"
			assert.Equal(t, expected, writer.String())
		case <-timeout:
			assert.Fail(t, "timeout stop buildlog")
		}
	})

	t.Run("success with long JSON output", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "*")
		if err != nil {
			t.Fatal(err)
		}
		defer tmpFile.Close()

		longBuffer := make([]byte, 4096, 4096)
		for i := 0; i < 4096; i++ {
			longBuffer[i] = '*'
		}
		longInputs := []string{
			`{"t": 1581662022394, "m": "test 1", "n": 0, "s": "main"}` + "\n",
			`{"t": 1581662022395, "m": "long input ` + string(longBuffer) + `", "n": 1, "s": "main"}` + "\n",
			`{"t": 1581662022394, "m": "test 3", "n": 2, "s": "main"}` + "\n",
		}

		go write(t, tmpFile.Name(), longInputs)

		parent, cancel := context.WithCancel(context.Background())
		writer := bytes.NewBuffer(nil)
		done := make(chan struct{})
		l := log{
			file:   tmpFile,
			writer: writer,
			ctx:    parent,
			cancel: cancel,
			done:   done,
		}

		go l.Run()

		time.Sleep(intervalTime * time.Millisecond)
		l.Stop()

		timeout := time.After(5 * time.Second)

		select {
		case <-done:
			expected := "main: test 1\r\nmain: long input " + string(longBuffer) + "\r\nmain: test 3\r\n"
			assert.Equal(t, expected, writer.String())
		case <-timeout:
			assert.Fail(t, "timeout stop buildlog")
		}
	})

	t.Run("continue builds with parsing error", func(t *testing.T) {
		defer func() {
			logrus.SetOutput(os.Stderr)
		}()
		tmpFile, err := os.CreateTemp("", "*")
		if err != nil {
			t.Fatal(err)
		}
		defer tmpFile.Close()

		testInvalidInputs := []string{
			`{"t": 1581662022394, "m": "test 1", "n": 0, "s": "main"}` + "\n",
			`{` + "\n",
			`{"t": 1581662022394, "m": "test 3", "n": 0, "s": "main"}` + "\n",
		}
		go write(t, tmpFile.Name(), testInvalidInputs)

		parent, cancel := context.WithCancel(context.Background())
		writer := bytes.NewBuffer(nil)
		done := make(chan struct{})
		l := log{
			file:   tmpFile,
			writer: writer,
			ctx:    parent,
			cancel: cancel,
			done:   done,
		}
		textFormatter := new(logrus.TextFormatter)
		textFormatter.PadLevelText = true
		logrus.SetFormatter(textFormatter)
		logrus.SetOutput(writer)

		go l.Run()

		time.Sleep(intervalTime * time.Millisecond)
		l.Stop()

		timeout := time.After(5 * time.Second)

		select {
		case <-done:
			assert.Contains(t, writer.String(), "main: test 1")
			assert.Contains(t, writer.String(), "Parsed error. If you want to check see sd-artifacts/builds.log:2")
			assert.Contains(t, writer.String(), "main: test 3")
		case <-timeout:
			assert.Fail(t, "timeout stop buildlog")
		}
	})
}

func TestStop(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	l := log{
		ctx:    parent,
		cancel: cancel,
		done:   make(chan struct{}),
	}

	timeout := time.After(5 * time.Second)
	l.Stop()

	select {
	case v := <-l.ctx.Done():
		assert.Equal(t, struct{}{}, v)
	case <-timeout:
		assert.Fail(t, "timeout stop buildlog")
	}
}

func TestNew(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "*")
		if err != nil {
			t.Fatal(err)
		}
		defer tmpFile.Close()

		writer := bytes.NewBuffer(nil)

		loggerDone := make(chan struct{})
		logger, err := New(tmpFile.Name(), writer, loggerDone)
		if err != nil {
			t.Fatal(err)
		}

		log, ok := logger.(*log)
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

		loggerDone := make(chan struct{})
		logger, err := New("/", writer, loggerDone)
		if err == nil {
			t.Fatal("failure err is nil")
		}

		expected := &log{
			writer: writer,
			file:   (*os.File)(nil),
			done:   loggerDone,
		}

		msg := err.Error()
		assert.Equal(t, expected, logger)
		assert.Equal(t, 0, strings.Index(msg, "failed to open raw build log file: "), fmt.Sprintf("expected error is `failed to open raw build log file: ...`, actual: `%v`", msg))
	})
}
