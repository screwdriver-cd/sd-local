package buildlog

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

var readInterval time.Duration = 10 * time.Millisecond

type Logger interface {
	Run() chan error
	Stop()
}

type log struct {
	ctx    context.Context
	file   io.Reader
	writer io.Writer
	cancel context.CancelFunc
}

type logLine struct {
	Time     int64  `json:"t"`
	Message  string `json:"m"`
	Line     int    `json:"n"`
	StepName string `json:"s"`
}

func New(ctx context.Context, filepath string, writer io.Writer) (Logger, error) {
	log := log{
		writer: writer,
	}

	var err error
	log.file, err = os.OpenFile(filepath, os.O_RDONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return log, fmt.Errorf("failed to open raw build log file :%w", err)
	}

	log.ctx, log.cancel = context.WithCancel(ctx)

	return log, nil
}

func (l log) Run() chan error {
	errChan := make(chan error)
	go l.run(errChan)
	return errChan
}

func (l log) Stop() {
	l.cancel()
	return
}

func (l log) run(errChan chan error) {
	reader := bufio.NewReader(l.file)

	for {
		select {
		case <-l.ctx.Done():
			break
		default:
			l.output(reader)
		}
		time.Sleep(readInterval)
	}
}

func (l log) output(reader *bufio.Reader) error {
	line, _, err := reader.ReadLine()
	if err != nil {
		return nil
	}

	parsedLog, err := parse(line)
	if err != nil {
		return fmt.Errorf("failed to output log: %w", err)
	}

	fmt.Fprintln(l.writer, parsedLog)
	return nil
}

func parse(rawLog []byte) (string, error) {
	ll := &logLine{}
	err := json.Unmarshal(rawLog, ll)
	if err != nil {
		return "", fmt.Errorf("failed to parse raw log: %w", err)
	}

	ISOTime := time.Unix(ll.Time, 0)

	return fmt.Sprintf("%s: %s", ISOTime.String(), ll.Message), nil
}
