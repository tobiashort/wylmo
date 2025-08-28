package choose

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	. "github.com/tobiashort/utils-go/must"
)

const (
	DEFAULT_NONE = iota
	DEFAULT_NO
	DEFAULT_YES
)

func YesNo(prompt string, default_ int) bool {
	b := strings.Builder{}
	b.WriteString(prompt)
	b.WriteString(" (")
	if default_ == DEFAULT_YES {
		b.WriteString("Y")
	} else {
		b.WriteString("y")
	}
	b.WriteString("/")
	if default_ == DEFAULT_NO {
		b.WriteString("N")
	} else {
		b.WriteString("n")
	}
	b.WriteString(") ")

	r := bufio.NewReader(os.Stdin)

ask:
	fmt.Print(b.String())
	ans := strings.TrimSpace(string(Must2(r.ReadBytes('\n'))))
	switch ans {
	case "":
		switch default_ {
		case DEFAULT_NONE:
			goto ask
		case DEFAULT_NO:
			return false
		default:
			return true
		}
	case "y":
		fallthrough
	case "Y":
		return true
	case "n":
		fallthrough
	case "N":
		return false
	default:
		goto ask
	}
}
