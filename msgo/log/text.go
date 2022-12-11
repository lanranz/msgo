package log

import (
	"fmt"
	"strings"
	"time"
)

type TextFormatter struct {
}

func (f *TextFormatter) Format(param *LoggingFormatParam) string {
	now := time.Now()
	var builderField strings.Builder
	var fieldsString = ""
	if param.LoggerFields != nil {
		fieldsString = "| fields: "
		num := len(param.LoggerFields)
		count := 0
		for k, v := range param.LoggerFields {
			fmt.Fprintf(&builderField, "%s=%v", k, v)
			if count < num-1 {
				fmt.Fprintf(&builderField, ",")
				count++
			}
		}
		fieldsString = builderField.String()
	}
	var msgInfo = "\n msg: "
	if param.Level == LevelError {
		msgInfo = "\n Error Cause By:"
	}
	if param.IsColor {
		//要带颜色  error的颜色 为红色 info为绿色 debug为蓝色
		levelColor := f.LevelColor(param.Level)
		msgColor := f.MsgColor(param.Level)
		return fmt.Sprintf("%s [msgo] %s %s%v%s | level= %s %s %s %s %s %#v %s  %s",
			yellow, reset, blue, now.Format("2006/01/02 - 15:04:05"), reset,
			levelColor, param.Level.Level(), reset, msgColor, msgInfo, param.Msg, reset, fieldsString,
		)
	}
	return fmt.Sprintf("[msgo] %v | level=%s %s %#v  %s ",
		now.Format("2006/01/02 - 15:04:05"),
		param.Level.Level(), msgInfo, param.Msg, fieldsString,
	)
}

func (f *TextFormatter) LevelColor(level LoggerLevel) string {
	switch level {
	case LevelDebug:
		return blue
	case LevelInfo:
		return green
	case LevelError:
		return red
	default:
		return cyan
	}
}

func (f *TextFormatter) MsgColor(level LoggerLevel) string {
	switch level {
	case LevelDebug:
		return ""
	case LevelInfo:
		return ""
	case LevelError:
		return red
	default:
		return cyan
	}
}
