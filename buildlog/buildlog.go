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
}

type logLine struct {
	Time     int64  `json:"t"`
	Message  string `json:"m"`
	Line     int    `json:"n"`
	StepName string `json:"s"`
}

// New creates new Logger interface.
func New(ctx context.Context, filepath string, writer io.Writer) (Logger, error) {
	log := log{
		writer: writer,
	}

	var err error
	log.file, err = os.OpenFile(filepath, os.O_RDONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return log, fmt.Errorf("failed to open raw build log file: %w", err)
	}

	log.ctx, log.cancel = context.WithCancel(ctx)

	return log, nil
}

func (l log) Stop() {
	l.cancel()
	return
}

func (l log) Run() {
	reader := bufio.NewReader(l.file)

	for {
		select {
		case <-l.ctx.Done():
			break
		default:
			err := l.output(reader)
			if err != nil {
				fmt.Printf("failed to run logger: %w\n", err)
				l.cancel()
			}
		}
		time.Sleep(readInterval)
	}
}

func (l log) output(reader *bufio.Reader) error {
	line, _, err := reader.ReadLine()
	if err != nil {
		if err != io.EOF {
			return fmt.Errorf("failed to read logfile: %w", err)
		}
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

	ISOTime := time.Unix(0, ll.Time*int64(time.Millisecond))

	return fmt.Sprintf("%s: %s", ISOTime.String(), ll.Message), nil
}
