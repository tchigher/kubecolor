package printer

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

	"github.com/dty1er/kubecolor/color"
	"github.com/dty1er/kubecolor/kubectl"
)

type GetPrinter struct {
	Writer         io.Writer
	WithHeader     bool
	FormatOpt      kubectl.FormatOption
	DarkBackground bool

	isFirstLine bool
	inString    bool
}

func (gp *GetPrinter) Print(outReader io.Reader) {
	gp.isFirstLine = true
	scanner := bufio.NewScanner(outReader)
	for scanner.Scan() {
		line := scanner.Text()
		switch gp.FormatOpt {
		case kubectl.Json:
			gp.PrintJson(line)
		case kubectl.Yaml:
			gp.PrintYaml(line)
		default:
			gp.PrintTable(line)
		}
	}
}

func (gp *GetPrinter) PrintTable(line string) {
	if gp.isHeader() {
		fmt.Fprintf(gp.Writer, "%s\n", color.Apply(line, getHeaderColorByBackground(gp.DarkBackground)))
		gp.isFirstLine = false
		return
	}

	printLineAsTableFormat(gp.Writer, line, gp.DarkBackground, gp.DecideColor)
}

func (gp *GetPrinter) PrintJson(line string) {
	w := gp.Writer
	indentCnt := gp.findIndent(line)
	trimmedLine := strings.TrimLeft(line, " ")

	if strings.HasPrefix(trimmedLine, "{") ||
		strings.HasPrefix(trimmedLine, "}") ||
		strings.HasPrefix(trimmedLine, "]") {
		// when coming here, it must not be starting with key.
		// that patterns are:
		// {
		// }
		// },
		// ]
		// ],
		// note: it must not be "[" because it will be always after key
		// in this case, just write it without color
		fmt.Fprintf(w, "%s", toSpaces(indentCnt))
		fmt.Fprintf(w, "%s", trimmedLine)
		fmt.Fprintf(w, "\n")
		return
	}

	// when coming here:
	// "key": {
	// "key": [
	// "key": value
	// "key": value,
	trimmed := strings.SplitN(trimmedLine, ": ", 2) // if key contains ": " this works in a wrong way but it's unlikely to happen

	if len(trimmed) == 1 {
		// when coming here, it will be a value in an array
		if strings.HasSuffix(trimmed[0], ",") {
			// when coming here, it must be `value,`
			ss := strings.TrimRight(trimmed[0], ",") // this is a value; it might be double-quoted or not
			if strings.HasPrefix(ss, `"`) && strings.HasSuffix(ss, `"`) {
				ss = strings.TrimLeft(ss, `"`)
				ss = strings.TrimRight(ss, `"`)
				fmt.Fprintf(w, "%s", toSpaces(indentCnt))
				fmt.Fprintf(w, `"`)
				fmt.Fprintf(w, "%s", color.Apply(ss, gp.colorByIndent(indentCnt)))
				fmt.Fprintf(w, `",`)
				fmt.Fprintf(w, "\n")
			} else {
				fmt.Fprintf(w, "%s", toSpaces(indentCnt))
				fmt.Fprintf(w, "%s", color.Apply(ss, getColorByValueType(ss, gp.DarkBackground)))
				fmt.Fprintf(w, "\n")
			}
		} else {
			ss := trimmed[0]
			// when coming here, it must be `value`
			if strings.HasPrefix(ss, `"`) && strings.HasSuffix(ss, `"`) {
				ss = strings.TrimLeft(ss, `"`)
				ss = strings.TrimRight(ss, `"`)
				fmt.Fprintf(w, "%s", toSpaces(indentCnt))
				fmt.Fprintf(w, `"`)
				fmt.Fprintf(w, "%s", color.Apply(ss, gp.colorByIndent(indentCnt)))
				fmt.Fprintf(w, `"`)
				fmt.Fprintf(w, "\n")
			} else {
				fmt.Fprintf(w, "%s", toSpaces(indentCnt))
				fmt.Fprintf(w, "%s", color.Apply(ss, getColorByValueType(ss, gp.DarkBackground)))
				fmt.Fprintf(w, "\n")
			}
		}
		return
	}

	key := trimmed[0]
	key = strings.TrimLeft(key, `"`)
	key = strings.TrimRight(key, `"`)

	if strings.HasSuffix(trimmedLine, "{") {
		// trim double quotation and colon, bracket
		fmt.Fprintf(w, "%s", toSpaces(indentCnt))
		fmt.Fprintf(w, `"`)
		fmt.Fprintf(w, "%s", color.Apply(key, gp.colorByIndent(indentCnt)))
		fmt.Fprintf(w, `": {`)
		fmt.Fprintf(w, "\n")
	} else if strings.HasSuffix(trimmedLine, "[") {
		// trim double quotation and colon, bracket
		fmt.Fprintf(w, "%s", toSpaces(indentCnt))
		fmt.Fprintf(w, `"`)
		fmt.Fprintf(w, "%s", color.Apply(key, gp.colorByIndent(indentCnt)))
		fmt.Fprintf(w, `": [`)
		fmt.Fprintf(w, "\n")
	} else if strings.HasSuffix(trimmed[1], ",") {
		// when coming here, it must be `"key": value,`
		ss := strings.TrimRight(trimmed[1], ",") // this is a value; it might be double-quoted or not
		if strings.HasPrefix(ss, `"`) && strings.HasSuffix(ss, `"`) {
			ss = strings.TrimLeft(ss, `"`)
			ss = strings.TrimRight(ss, `"`)
			fmt.Fprintf(w, "%s", toSpaces(indentCnt))
			fmt.Fprintf(w, `"`)
			fmt.Fprintf(w, "%s", color.Apply(key, gp.colorByIndent(indentCnt)))
			fmt.Fprintf(w, `": "`)
			fmt.Fprintf(w, "%s", color.Apply(ss, getColorByValueType(ss, gp.DarkBackground)))
			fmt.Fprintf(w, `",`)
			fmt.Fprintf(w, "\n")
		} else {
			fmt.Fprintf(w, "%s", toSpaces(indentCnt))
			fmt.Fprintf(w, `"`)
			fmt.Fprintf(w, "%s", color.Apply(key, gp.colorByIndent(indentCnt)))
			fmt.Fprintf(w, `": `)
			fmt.Fprintf(w, "%s", color.Apply(ss, getColorByValueType(ss, gp.DarkBackground)))
			fmt.Fprintf(w, `,`)
			fmt.Fprintf(w, "\n")
		}
	} else {
		// when coming here, it must be `"key": value`
		ss := trimmed[1]
		if strings.HasPrefix(ss, `"`) && strings.HasSuffix(ss, `"`) {
			ss = strings.TrimLeft(ss, `"`)
			ss = strings.TrimRight(ss, `"`)
			fmt.Fprintf(w, "%s", toSpaces(indentCnt))
			fmt.Fprintf(w, `"`)
			fmt.Fprintf(w, "%s", color.Apply(key, gp.colorByIndent(indentCnt)))
			fmt.Fprintf(w, `": "`)
			fmt.Fprintf(w, "%s", color.Apply(ss, getColorByValueType(ss, gp.DarkBackground)))
			fmt.Fprintf(w, `"`)
			fmt.Fprintf(w, "\n")
		} else {
			fmt.Fprintf(w, "%s", toSpaces(indentCnt))
			fmt.Fprintf(w, `"`)
			fmt.Fprintf(w, "%s", color.Apply(key, gp.colorByIndent(indentCnt)))
			fmt.Fprintf(w, `": `)
			fmt.Fprintf(w, "%s", color.Apply(ss, getColorByValueType(ss, gp.DarkBackground)))
			fmt.Fprintf(w, "\n")
		}
	}
}

func (gp *GetPrinter) PrintYaml(line string) {
	w := gp.Writer
	indentCnt := gp.findIndent(line)
	trimmedLine := strings.TrimLeft(line, " ")

	if strings.HasPrefix(trimmedLine, "-") {
		// when coming here, it must be "- key: value" or "- value"
		trimmed := strings.TrimLeft(trimmedLine, "- ")
		if strings.Contains(trimmed, ": ") && unicode.IsLetter(rune(trimmed[0])) {
			// when coming here, it must be "- key: value"
			ss := strings.SplitN(trimmed, ": ", 2) // assuming key must not contain ": " while value might do
			k, v := ss[0], ss[1]
			fmt.Fprintf(w, "%s", toSpaces(indentCnt))
			fmt.Fprintf(w, "- ")
			fmt.Fprintf(w, "%s", color.Apply(k, gp.colorByIndent(indentCnt+2))) // add length of "- "
			fmt.Fprintf(w, ": ")
			fmt.Fprintf(w, "%s", color.Apply(v, getColorByValueType(v, gp.DarkBackground)))
			fmt.Fprintf(w, "\n")
		} else {
			// when coming here, it must be "- value"
			fmt.Fprintf(w, "%s", toSpaces(indentCnt))
			fmt.Fprintf(w, "- ")
			fmt.Fprintf(w, "%s", color.Apply(trimmed, getColorByValueType(trimmed, gp.DarkBackground)))
			fmt.Fprintf(w, "\n")
		}
		return
	}

	// when coming here, "key:" or "key: value" or "value"
	if strings.Contains(trimmedLine, ": ") && unicode.IsLetter(rune(trimmedLine[0])) {
		// when coming here, it must be "key: value"
		ss := strings.SplitN(trimmedLine, ": ", 2) // assuming key must not contain ": " while value might do
		k, v := ss[0], ss[1]
		fmt.Fprintf(w, "%s", toSpaces(indentCnt))
		fmt.Fprintf(w, "%s", color.Apply(k, gp.colorByIndent(indentCnt)))
		fmt.Fprintf(w, ": ")
		fmt.Fprintf(w, "%s", color.Apply(v, getColorByValueType(v, gp.DarkBackground)))
		fmt.Fprintf(w, "\n")
	} else if strings.HasSuffix(trimmedLine, ":") && unicode.IsLetter(rune(trimmedLine[0])) {
		// when coming here, it must be "key:"
		trimmed := strings.TrimRight(trimmedLine, ":")
		fmt.Fprintf(w, "%s", toSpaces(indentCnt))
		fmt.Fprintf(w, "%s", color.Apply(trimmed, gp.colorByIndent(indentCnt)))
		fmt.Fprintf(w, ":")
		fmt.Fprintf(w, "\n")
	} else {
		// when coming here, it must be just a "value"
		// when a string was too long, the line can be broken and come here
		fmt.Fprintf(w, "%s", toSpaces(indentCnt))
		fmt.Fprintf(w, "%s", color.Apply(trimmedLine, getColorByValueType(trimmedLine, gp.DarkBackground)))
		fmt.Fprintf(w, "\n")
	}
}

func (gp *GetPrinter) colorByIndent(indent int) color.Color {
	switch indent / 4 % 2 {
	case 1:
		return color.White
	default:
		return color.Yellow
	}
}

func (gp *GetPrinter) findIndent(line string) int {
	return len(line) - len(strings.TrimLeft(line, " "))
}

func (gp *GetPrinter) isHeader() bool {
	return gp.WithHeader && gp.isFirstLine
}

func (gp *GetPrinter) DecideColor(_ int, column string) (color.Color, bool) {
	if column == "CrashLoopBackOff" {
		return color.Red, true
	}

	// When Readiness is "n/m" then yellow
	if strings.Count(column, "/") == 1 {
		if arr := strings.Split(column, "/"); arr[0] != arr[1] {
			_, e1 := strconv.Atoi(arr[0])
			_, e2 := strconv.Atoi(arr[1])
			if e1 == nil && e2 == nil { // check both is number
				return color.Yellow, true
			}
		}

	}

	return 0, false
}
