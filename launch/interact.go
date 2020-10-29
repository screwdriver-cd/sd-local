package launch

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/creack/pty"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
)

type Interact interface {
	Run(c *exec.Cmd, commands [][]string) error
}

type InteractImpl struct {
}

func (d *InteractImpl) Run(c *exec.Cmd, commands [][]string) error {
	ptmx, tty, err := pty.Open()
	c.Stdin = tty
	c.Stdout = tty
	c.Stderr = tty
	if err != nil {
		logrus.Warn(fmt.Errorf("failed: %s", err))
		return err
	}

	// Make sure to close the pty at the end.
	defer func() {
		_ = ptmx.Close()
		_ = tty.Close()
	}()

	// Handle pty size.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			if err := pty.InheritSize(os.Stdin, ptmx); err != nil {
				logrus.Warn(fmt.Errorf("error resizing pty: %s", err))
			}
		}
	}()
	ch <- syscall.SIGWINCH // Initial resize.

	// Set stdin in raw mode.
	oldState, err := terminal.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer func() { _ = terminal.Restore(int(os.Stdin.Fd()), oldState) }() // Best effort.

	// Start the command
	c.Start()

	// Copy stdin to the pty and the pty to stdout.
	go func() {
		for _, v := range commands {
			v = append(v, "\n")
			_, _ = io.Copy(ptmx, strings.NewReader(strings.Join(v, " ")))
		}
		_, _ = io.Copy(ptmx, os.Stdin)
	}()
	go func() {
		_, _ = io.Copy(os.Stdout, ptmx)
	}()

	c.Wait()

	return nil
}
