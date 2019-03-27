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

	if agent, has := entry.Data["agent"]; has {
		b.WriteString(agent.(string))
		b.WriteString(" ")
		delete(entry.Data, "agent")
	} else if entry.HasCaller() {
		if entry.Caller.Function == "github.com/kataras/golog.integrateStdLogger.func1" {
			entry.Caller.Function = "iris"
			entry.Caller.File = "iris"
		} else {
			if idx := strings.LastIndex(entry.Caller.Function, "/"); idx > 0 {
				entry.Caller.Function = entry.Caller.Function[idx+1:]
			}
		}
		if idx := strings.LastIndex(entry.Caller.File, "/"); idx > 0 {
			entry.Caller.File = entry.Caller.File[idx+1:]
		}

		if entry.Level == logrus.ErrorLevel {
			b.WriteString(fmt.Sprintf("%s(%s:%d) ",
				entry.Caller.Function, entry.Caller.File, entry.Caller.Line))
		}
	}
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
