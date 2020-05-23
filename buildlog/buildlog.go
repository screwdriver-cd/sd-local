package buildlog

import (
	"bufio"
	"context"
	"encoding/json"
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
	Done() <-chan interface{}
}

type log struct {
	ctx    context.Context
	file   io.Reader
	writer io.Writer
	cancel context.CancelFunc
	done   chan interface{}
}

type logLine struct {
	Time     int64  `json:"t"`
	Message  string `json:"m"`
	Line     int    `json:"n"`
	StepName string `json:"s"`
}

// New creates new Logger interface.
func New(filepath string, writer io.Writer) (Logger, error) {
	log := log{
		writer: writer,
		done:   make(chan interface{}),
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

loop:
	for {
		select {
		case <-l.ctx.Done():
			buildDone = true
		default:
			readDone, err := l.output(reader)
			if err != nil {
				logrus.Errorf("failed to run logger: %v\n", err)
				logrus.Info("But build is not stopping")
				l.cancel()
			}

			if buildDone && readDone {
				close(l.done)
				break loop
			}
		}
		time.Sleep(readInterval)
	}
}

func (l log) Done() <-chan interface{} {
	return l.done
}

func (l log) output(reader *bufio.Reader) (bool, error) {
	line, _, err := reader.ReadLine()
	if err != nil {
		if err == io.EOF {
			return true, nil
		}
		return false, fmt.Errorf("failed to read logfile: %w", err)
	}

	parsedLog, err := parse(line)
	if err != nil {
		return false, fmt.Errorf("failed to output log: %w", err)
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
