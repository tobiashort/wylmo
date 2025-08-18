package choose

import (
	"fmt"
	"os"
	"syscall"

	"golang.org/x/term"

	"github.com/tobiashort/ansi-go"
	"github.com/tobiashort/cfmt-go"
	"github.com/tobiashort/orderedmap-go"
	"github.com/tobiashort/utils-go/assert"

	. "github.com/tobiashort/utils-go/must"
)

func Many(prompt string, options []string) ([]string, bool) {
	return ManyN(prompt, options, len(options))
}

func ManyN(prompt string, options []string, n int) ([]string, bool) {
	assert.True(n >= 2, "n must be >= 2")
	assert.True(n <= len(options), "n must be <= len(options)")

	oldState := Must2(term.MakeRaw(int(os.Stdin.Fd())))
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	ok := false
	activeIndex := 0
	selectedCount := 0

	selection := orderedmap.NewOrderedMap[string, bool]()
	for _, option := range options {
		selection.Put(option, false)
	}

draw:
	fmt.Printf("%s (%d/%d)\r\n", prompt, selectedCount, n)
	for index, option := range options {
		if index == activeIndex {
			if selected, _ := selection.Get(option); selected {
				cfmt.Printf(" #yB{> [x] %s}\r\n", option)
			} else {
				cfmt.Printf(" #yB{> [ ] %s}\r\n", option)
			}
		} else {
			if selected, _ := selection.Get(option); selected {
				fmt.Printf("   [x] %s\r\n", option)
			} else {
				fmt.Printf("   [ ] %s\r\n", option)
			}
		}
	}

	buf := make([]byte, 3)
	for {
		c := Must2(os.Stdin.Read(buf))
		switch string(buf[:c]) {
		case "A":
			if n == len(options) {
				selectedCount = n
				for option := range selection.Iterate() {
					selection.Put(option, true)
				}
				goto redraw
			}
		case "N":
			selectedCount = 0
			for option := range selection.Iterate() {
				selection.Put(option, false)
			}
			goto redraw
		case "j":
			fallthrough
		case ansi.InputKeyDown:
			if activeIndex < len(options)-1 {
				activeIndex++
			} else {
				activeIndex = 0
			}
			goto redraw
		case "k":
			fallthrough
		case ansi.InputKeyUp:
			if activeIndex > 0 {
				activeIndex--
			} else {
				activeIndex = len(options) - 1
			}
			goto redraw
		case ansi.InputSpace:
			option := options[activeIndex]
			selected, _ := selection.Get(option)
			if selected {
				selectedCount--
				selection.Put(option, false)
			} else {
				if selectedCount < n {
					selectedCount++
					selection.Put(option, true)
				}
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
	fmt.Printf("%s\r", ansi.EraseEntireLine)
	selectedOptions := make([]string, 0)
	for option, selected := range selection.Iterate() {
		if selected {
			selectedOptions = append(selectedOptions, option)
		}
	}

	return selectedOptions, ok
}
