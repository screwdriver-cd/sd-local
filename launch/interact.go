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
	"golang.org/x/term"
)

// Interacter wraps up the interactive process
type Interacter interface {
	Run(c *exec.Cmd, commands [][]string) error
}

// Interact takes interactive processing
type Interact struct {
	flagVerbose bool
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
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }() // Best effort.

	// Start the command
	c.Start()

	// Copy stdin to the pty and the pty to stdout.
	go func() {
		if !d.flagVerbose {
			// Clear entire screen
			_, _ = io.Copy(os.Stdout, strings.NewReader("\x1b[2J"))
		}

		for _, v := range commands {
			func() {
				if !d.flagVerbose {
					// Moves the cursor to row 1, column 1, and Clear from cursor to end of screen.
					_, _ = io.Copy(os.Stdout, strings.NewReader("\x1b[1;H\x1b[0Jplease wait while sd-local setup process...\r\n"))
					// Decreased intensity
					_, _ = io.Copy(os.Stdout, strings.NewReader("\x1b[2m"))
					// Normal intensity
					defer io.Copy(os.Stdout, strings.NewReader("\x1b[22m"))
				}

				v := append(v, "\n")
				command := strings.Join(v, " ")
				for {
					if len(command) <= maxByte {
						io.Copy(ptmx, strings.NewReader(command[:]))
						// Wait send the command.
						time.Sleep(time.Millisecond * 300)
						break
					} else {
						io.Copy(ptmx, strings.NewReader(command[:maxByte]))
						// Wait send the command.
						time.Sleep(time.Millisecond * 300)
						command = command[maxByte:]
					}
				}
			}()
		}
		// wait Launcher setup
		time.Sleep(time.Second * 1)
		if !d.flagVerbose {
			// Moves the cursor to row 1, column 1, and Clear from cursor to end of screen.
			_, _ = io.Copy(os.Stdout, strings.NewReader("\x1b[1;H\x1b[0J"))
		} else {
			// spacer
			_, _ = io.Copy(os.Stdout, strings.NewReader("\r\n"))
		}
		_, _ = io.Copy(os.Stdout, strings.NewReader("Welcome to sd-local interactive mode. To exit type 'exit'\n"))
		_, _ = io.Copy(ptmx, strings.NewReader("\n"))
		_, _ = io.Copy(ptmx, os.Stdin)
	}()
	go func() {
		_, _ = io.Copy(os.Stdout, ptmx)
	}()

	c.Wait()

	return nil
}
