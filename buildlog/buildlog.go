package buildlog

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

var readInterval time.Duration = 10 * time.Millisecond

// Logger outputs logs
type Logger interface {
	Run()
	Stop()
}

type log struct {
	ctx    context.Context
	file   io.Reader
	writer io.Writer
	cancel context.CancelFunc
	done   chan<- struct{}
}

type logLine struct {
	Time     int64  `json:"t"`
	Message  string `json:"m"`
	Line     int    `json:"n"`
	StepName string `json:"s"`
}

type parseError struct{}

func (e *parseError) Error() string { return "Parse Error" }

// New creates new Logger interface.
func New(filepath string, writer io.Writer, done chan<- struct{}) (Logger, error) {
	log := log{
		writer: writer,
		done:   done,
	}

	var err error
	log.file, err = os.OpenFile(filepath, os.O_RDONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return log, fmt.Errorf("failed to open raw build log file: %w", err)
	}

	log.ctx, log.cancel = context.WithCancel(context.Background())

	return log, nil
}

func (l log) Stop() {
	l.cancel()
}

func (l log) Run() {
	reader := bufio.NewReader(l.file)
	buildDone := false

	for {
		if !buildDone {
			select {
			case <-l.ctx.Done():
				buildDone = true
			default:
			}
		}

		readDone, err := l.output(reader)
		parseErr := &parseError{}
		if errors.As(err, &parseErr) {
			continue
		}

		if err != nil {
			logrus.Errorf("failed to run logger: %v\n", err)
			logrus.Info("But build is still running")
			close(l.done)
			break
		}

		if buildDone && readDone {
			close(l.done)
			break
		}
		time.Sleep(readInterval)
	}
}

func readln(prefix []byte, r *bufio.Reader) ([]byte, error) {
	line, isPrefix, err := r.ReadLine()
	if err != nil {
		return []byte{}, err
	}

	if isPrefix {
		return readln(append(prefix, line...), r)
	}

	return append(prefix, line...), err
}

func (l log) output(reader *bufio.Reader) (bool, error) {
	line, err := readln([]byte{}, reader)
	if err != nil {
		if err == io.EOF {
			return true, nil
		}
		return false, fmt.Errorf("failed to read logfile: %w", err)
	}

	parsedLog, err := parse(line)
	if err != nil {
		fmt.Fprintln(l.writer, "\\e[33mwaring: parsed error\\e[m")
		return false, &parseError{}
	}

	fmt.Fprintln(l.writer, parsedLog)
	return false, nil
}

func parse(rawLog []byte) (string, error) {
	ll := &logLine{}
	err := json.Unmarshal(rawLog, ll)
	if err != nil {
		return "", fmt.Errorf("failed to parse raw log: %w", err)
	}

	return fmt.Sprintf("%s: %s", ll.StepName, ll.Message), nil
}
