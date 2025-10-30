package redishook

import (
	"fmt"
	"strings"
)

func commandArgsToString(args []interface{}) string {
	var b strings.Builder

	for i, arg := range args {
		if i > 0 {
			b.WriteString(" ")
		}

		b.WriteString(fmt.Sprintf("%v", arg))
	}

	return b.String()
}
