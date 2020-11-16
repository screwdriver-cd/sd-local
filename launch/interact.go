package launch

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/creack/pty"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
)

// Interacter wraps up the interactive process
type Interacter interface {
	Run(c *exec.Cmd, commands [][]string) error
}

// Interact takes interactive processing
type Interact struct {
}

// Run runs interactive process
func (d *Interact) Run(c *exec.Cmd, commands [][]string) error {

	// The maximum number of bytes to be written to /dev/pts
	const maxByte = 300

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

			v := append(v, "\n")
			command := strings.Join(v, " ")
			if len(command) > maxByte {
				for {
					if len(command) <= maxByte {
						io.Copy(ptmx, strings.NewReader(command[:]))
						// Wait send the command.
						time.Sleep(time.Second * 1)
						break
					} else {
						io.Copy(ptmx, strings.NewReader(command[:maxByte]))
						// Wait send the command.
						time.Sleep(time.Second * 1)
						command = command[maxByte:]
					}
				}
			} else {
				_, _ = io.Copy(ptmx, strings.NewReader(command))
			}
		}
		// wait Launcher setup
		time.Sleep(time.Second * 1)
		_, _ = io.Copy(os.Stdout, strings.NewReader("\r\nWelcome to sd-local interactive mode. If you exit type 'exit'\n"))
		_, _ = io.Copy(ptmx, strings.NewReader("\n"))
		_, _ = io.Copy(ptmx, os.Stdin)
	}()
	go func() {
		_, _ = io.Copy(os.Stdout, ptmx)
	}()

	c.Wait()

	return nil
}
