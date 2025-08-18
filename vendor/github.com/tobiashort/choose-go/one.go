package choose

import (
	"fmt"
	"os"
	"syscall"

	"github.com/tobiashort/ansi-go"
	"github.com/tobiashort/cfmt-go"
	. "github.com/tobiashort/utils-go/must"

	"golang.org/x/term"
)

func One(prompt string, options []string) (int, string, bool) {
	oldState := Must2(term.MakeRaw(int(os.Stdin.Fd())))
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	ok := false
	selectedIndex := 0

draw:
	fmt.Printf("%s\r\n", prompt)
	for index, option := range options {
		if index == selectedIndex {
			cfmt.Printf(" #yB{> %s}\r\n", option)
		} else {
			fmt.Printf("   %s\r\n", option)
		}
	}

	buf := make([]byte, 3)
	for {
		n := Must2(os.Stdin.Read(buf))
		switch string(buf[:n]) {
		case "j":
			fallthrough
		case ansi.InputKeyDown:
			if selectedIndex < len(options)-1 {
				selectedIndex++
			} else {
				selectedIndex = 0
			}
			goto redraw
		case "k":
			fallthrough
		case ansi.InputKeyUp:
			if selectedIndex > 0 {
				selectedIndex--
			} else {
				selectedIndex = len(options) - 1
			}
			goto redraw
		case ansi.InputCR:
			fallthrough
		case ansi.InputLF:
			fallthrough
		case ansi.InputCRLF:
			ok = true
			goto done
		case "q":
			fallthrough
		case ansi.InputEscape:
			ok = false
			goto done
		case ansi.InputCtrlC:
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
			goto done
		}
	}

redraw:
	fmt.Print(ansi.CursorMoveUp(len(options) + 1))
	goto draw

done:
	return selectedIndex, options[selectedIndex], ok
}
