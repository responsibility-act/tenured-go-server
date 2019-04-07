package main

import (
	"bytes"
	"fmt"
	"github.com/ihaiker/tenured-go-server/commons"
	"io"
	"log"
	"regexp"
	"strings"
	"text/template"
	"unicode"
)

type TenuredReader struct {
	lines   []string
	readIdx int
}

func fmto(ftl string, out io.Writer, data ...interface{}) {
	_, _ = fmt.Fprintf(out, ftl, data...)
}

func ftl(ftl string, data interface{}, out io.Writer) {
	t := template.Must(template.New("letter").Parse(ftl))
	if err := t.Execute(out, data); err != nil {
		log.Panic(err)
	}
}

func ftlc(ftlstr string, data interface{}) []byte {
	out := new(bytes.Buffer)
	ftl(ftlstr, data, out)
	return out.Bytes()
}

func lowerName(name string) string {
	newName := []rune(name)
	newName[0] = unicode.ToLower(newName[0])
	return string(newName)
}

func UpperName(name string) string {
	newName := []rune(name)
	newName[0] = unicode.ToUpper(newName[0])
	return string(newName)
}

func match(regexp *regexp.Regexp, line string) ([]string, error) {
	if !regexp.MatchString(line) {
		return nil, NotMatch
	}
	return regexp.FindStringSubmatch(line), nil
}

func body(lines []string) []string {
	return lines[1 : len(lines)-1]
}

func comment(lines []string) (string, []string) {
	startIdx := 0
	for _, line := range lines {
		if strings.HasPrefix(line, "//") {
			startIdx++
		} else {
			break
		}
	}
	if startIdx == 0 {
		return "", lines
	} else {
		c := strings.Join(lines[0:startIdx], "\n")
		return c, lines[startIdx:]
	}
}

func trim(line string) string {
	line = strings.TrimLeftFunc(line, unicode.IsSpace)
	line = strings.TrimRightFunc(line, unicode.IsSpace)

	b := new(bytes.Buffer)
	first := true
	for _, v := range line {
		if unicode.IsSpace(v) {
			if first {
				b.WriteRune(' ')
			}
			first = false
		} else {
			first = true
			b.WriteRune(v)
		}
	}
	return b.String()
}

func (this *TenuredReader) next() bool {
	for i := this.readIdx; i < len(this.lines); i++ {
		if trim(this.lines[i]) != "" {
			return true
		}
	}
	return false
}

func (this *TenuredReader) line() (string, error) {
	if !(this.readIdx < len(this.lines)) {
		return "", io.EOF
	}
	for i := this.readIdx; i < len(this.lines); i++ {
		this.readIdx++
		line := trim(this.lines[i])
		if line != "" {
			return line, nil
		}
	}
	return "", nil
}

func (this *TenuredReader) Read() ([]string, error) {
	if !this.next() {
		return nil, io.EOF
	}
	lines := make([]string, 0)
	for {
		line, err := this.line()
		lines = append(lines, line)
		if err != nil {
			return lines, err
		} else if line == "}" {
			break
		}
	}
	return lines, nil
}

func NewReader(file string) (*TenuredReader, error) {
	f := commons.NewFile(file)
	if lines, err := f.Lines(); err != nil {
		return nil, err
	} else {
		return &TenuredReader{
			lines: lines, readIdx: 0,
		}, nil
	}
}
