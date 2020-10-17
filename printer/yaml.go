package printer

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/dty1er/kubecolor/color"
)

type YamlPrinter struct {
	DarkBackground bool

	inString bool
}

func (yp *YamlPrinter) Print(r io.Reader, w io.Writer) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		yp.printLineAsYamlFormat(line, w, yp.DarkBackground)
	}
}

func (yp *YamlPrinter) printLineAsYamlFormat(line string, w io.Writer, dark bool) {
	indentCnt := findIndent(line) // can be 0
	indent := toSpaces(indentCnt) // so, can be empty
	trimmedLine := strings.TrimLeft(line, " ")

	if yp.inString {
		// fmt.Println(trimmedLine)
		// if inString is true, the line must be a part of a string which is broken into several lines
		fmt.Fprintf(w, "%s%s\n", indent, yp.toColorizedStringValue(trimmedLine, dark))
		yp.inString = !yp.isStringClosed(trimmedLine)
		return
	}

	splitted := strings.SplitN(trimmedLine, ": ", 2) // assuming key does not contain ": " while value might do

	if len(splitted) == 2 {
		// key: value
		key, val := splitted[0], splitted[1]
		fmt.Fprintf(w, "%s%s: %s\n", indent, yp.toColorizedYamlKey(key, indentCnt, 2, dark), yp.toColorizedYamlValue(val, dark))
		yp.inString = yp.isStringOpenedButNotClosed(val)
		return
	}

	// when coming here, the line is just a "key:" or an element of an array
	if strings.HasSuffix(splitted[0], ":") {
		// key:
		fmt.Fprintf(w, "%s%s\n", indent, yp.toColorizedYamlKey(splitted[0], indentCnt, 2, dark))
		return
	}

	fmt.Fprintf(w, "%s%s\n", indent, yp.toColorizedYamlValue(splitted[0], dark))
}

func (yp *YamlPrinter) isValYamlKey(s string) bool {
	// key must end with :
	if !strings.HasSuffix(s, ":") {
		return false
	}

	// even if it ends with :, if it's in a string value, it's not a key
	if yp.inString {
		return false
	}

	return true
}

func (yp *YamlPrinter) toColorizedYamlKey(key string, indentCnt, basicWidth int, dark bool) string {
	hasColon := strings.HasSuffix(key, ":")
	hasLeadingDash := strings.HasPrefix(key, "- ")
	key = strings.TrimRight(key, ":")
	key = strings.TrimLeft(key, "- ")

	format := "%s"
	if hasColon {
		format += ":"
	}

	if hasLeadingDash {
		format = "- " + format
		indentCnt += 2
	}

	return fmt.Sprintf(format, color.Apply(key, getColorByKeyIndent(indentCnt, basicWidth, dark)))
}

func (yp *YamlPrinter) toColorizedYamlValue(value string, dark bool) string {
	if value == "{}" {
		return "{}"
	}

	hasLeadingDash := strings.HasPrefix(value, "- ")
	value = strings.TrimLeft(value, "- ")

	isDoubleQuoted := strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`)
	trimmedValue := strings.TrimRight(strings.TrimLeft(value, `"`), `"`)

	var format string
	switch {
	case hasLeadingDash && isDoubleQuoted:
		format = `- "%s"`
	case hasLeadingDash:
		format = "- %s"
	case isDoubleQuoted:
		format = `"%s"`
	default:
		format = "%s"
	}

	return fmt.Sprintf(format, color.Apply(trimmedValue, getColorByValueType(value, dark)))
}

func (yp *YamlPrinter) toColorizedStringValue(value string, dark bool) string {
	c := StringColorForLight
	if dark {
		c = StringColorForDark
	}

	isDoubleQuoted := strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`)
	trimmedValue := strings.TrimRight(strings.TrimLeft(value, `"`), `"`)

	var format string
	switch {
	case isDoubleQuoted:
		format = `"%s"`
	default:
		format = "%s"
	}
	return fmt.Sprintf(format, color.Apply(trimmedValue, c))
}

func (yp *YamlPrinter) isStringClosed(line string) bool {
	return strings.HasSuffix(line, "'") || strings.HasSuffix(line, `"`)
}

func (yp *YamlPrinter) isStringOpenedButNotClosed(line string) bool {
	return (strings.HasPrefix(line, "'") && !strings.HasSuffix(line, "'")) ||
		(strings.HasPrefix(line, `"`) && !strings.HasSuffix(line, `"`))
}
