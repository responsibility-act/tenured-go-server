package logs

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
)

type TextFormatter struct {
}

func (f *TextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b = &bytes.Buffer{}
	if entry.Buffer != nil {
		_, _ = entry.Buffer.WriteTo(b)
	}

	b.WriteString(entry.Time.Format("2006-01-02 15:04:05 "))
	b.WriteString("[")
	b.WriteString(entry.Level.String())
	b.WriteString("] ")

	agent, hasAgent := entry.Data["agent"]
	if hasAgent {
		b.WriteString("(")
		b.WriteString(agent.(string))
		b.WriteString(") ")
		delete(entry.Data, "agent")
	}

	if (uint32(entry.Level) < uint32(logrus.InfoLevel) || !hasAgent) && entry.HasCaller() {
		if entry.Caller.Function == "github.com/kataras/golog.integrateStdLogger.func1" {
			entry.Caller.Function = "iris"
			entry.Caller.File = "iris"
		} else {
			if idx := strings.LastIndex(entry.Caller.Function, "/"); idx > 0 {
				entry.Caller.Function = entry.Caller.Function[idx+1:]
			}
			if idx := strings.Index(entry.Caller.File, "/src/"); idx > 0 {
				entry.Caller.File = entry.Caller.File[idx+5:]
			}
		}

		b.WriteString(fmt.Sprintf("%s:%d %s",
			entry.Caller.File, entry.Caller.Line, entry.Caller.Function))
	}
	b.WriteString(": ")
	if len(entry.Data) > 0 {
		b.WriteString("{ ")
		for k, v := range entry.Data {
			_, _ = fmt.Fprintf(b, "%s=%v ", k, v)
		}
		b.WriteString("} ")
	}
	b.WriteString(entry.Message)
	b.WriteString("\n")

	return b.Bytes(), nil
}
