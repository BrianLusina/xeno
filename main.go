package main

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/creack/pty"
)

// MaxBufferSize sets the size limit for the command output buffer
const MaxBufferSize = 16

func main() {
	a := app.New()
	window := a.NewWindow("Xeno")

	textGrid := widget.NewTextGrid() // Create a new TextGrid

	os.Setenv("TERM", "dumb")

	command := exec.Command("/bin/bash") // Create a new command
	p, err := pty.Start(command)         // Start the command

	if err != nil {
		fyne.LogError("Failed to open PTY", err)
		os.Exit(1)
	}

	defer command.Process.Kill()

	onTypedKey := func(key *fyne.KeyEvent) {
		if key.Name == fyne.KeyReturn || key.Name == fyne.KeyEnter {
			_, _ = p.Write([]byte{'\r'})
		}
	}

	onTypedRune := func(r rune) {
		_, _ = p.WriteString(string(r))
	}

	window.Canvas().SetOnTypedKey(onTypedKey)
	window.Canvas().SetOnTypedRune(onTypedRune)

	buffer := [][]rune{}
	reader := bufio.NewReader(p)

	// Goroutine that reads from pty
	go func() {
		line := []rune{}
		buffer = append(buffer, line)

		for {
			r, _, err := reader.ReadRune()

			if err != nil {
				if err == io.EOF {
					return
				}

				os.Exit(0)
			}

			line = append(line, r)
			buffer[len(buffer)-1] = line

			if r == '\r' {
				// buffer is at capacity
				if len(buffer) > MaxBufferSize {
					// pop first line
					buffer = buffer[1:]
				}

				line = []rune{}
				buffer = append(buffer, line)
			}
		}
	}()

	// Goroutine that renders to UI
	go func() {
		for {
			time.Sleep(100 * time.Second)
			textGrid.SetText("")
			var lines string

			for _, line := range buffer {
				lines = lines + string(line)
			}
			textGrid.SetText(string(lines))
		}
	}()

	// Create a new container with a wrapped layout
	// set the layout width to 900, height to 325
	window.SetContent(
		container.New(
			layout.NewGridWrapLayout(fyne.NewSize(900, 325)),
			textGrid,
		),
	)

	window.ShowAndRun()
}
