package log

import (
	"fmt"
	"io"
	"log"
	"msgo/internal/msstrings"
	"os"
	"path"
	"strings"
	"time"
)

const (
	greenBg   = "\033[97;42m"
	whiteBg   = "\033[90;47m"
	yellowBg  = "\033[90;43m"
	redBg     = "\033[97;41m"
	blueBg    = "\033[97;44m"
	magentaBg = "\033[97;45m"
	cyanBg    = "\033[97;46m"
	green     = "\033[32m"
	white     = "\033[37m"
	yellow    = "\033[33m"
	red       = "\033[31m"
	blue      = "\033[34m"
	magenta   = "\033[35m"
	cyan      = "\033[36m"
	reset     = "\033[0m"
)

// 级别
type LoggerLevel int

func (l LoggerLevel) Level() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelError:
		return "ERROR"
	default:
		return ""
	}
}

const (
	LevelDebug LoggerLevel = iota
	LevelInfo
	LevelError
)

type Fields map[string]any

type Logger struct {
	Formatter      LoggingFormatter
	Level          LoggerLevel
	Outs           []*LoggerWriter
	LoggerFields   Fields
	LogPath        string
	LoggerFileSize int64
}

type LoggerWriter struct {
	Level LoggerLevel
	Out   io.Writer
}

type LoggingFormatter interface {
	Format(param *LoggingFormatParam) string
}

type LoggingFormatParam struct {
	IsColor      bool
	Level        LoggerLevel
	Msg          any
	LoggerFields Fields
}

type LoggerFormatter struct {
	Level        LoggerLevel
	IsColor      bool
	LoggerFields Fields
}

func New() *Logger {
	return &Logger{}
}

func Default() *Logger {
	logger := New()
	logger.Level = LevelDebug
	w := &LoggerWriter{
		Level: LevelDebug,
		Out:   os.Stdout,
	}
	logger.Outs = append(logger.Outs, w)
	logger.Formatter = &TextFormatter{}
	return logger
}

func (l *Logger) Info(msg any) {
	l.Print(LevelInfo, msg)
}

func (l *Logger) Debug(msg any) {
	l.Print(LevelDebug, msg)
}

func (l *Logger) Error(msg any) {
	l.Print(LevelError, msg)
}

func (l *Logger) Print(level LoggerLevel, msg any) {
	//级别不满足 不打印日志
	if l.Level > level {
		return
	}
	param := &LoggingFormatParam{
		Level:        level,
		LoggerFields: l.LoggerFields,
		Msg:          msg,
	}
	str := l.Formatter.Format(param)
	for _, out := range l.Outs {
		if out.Out == os.Stdout {
			param.IsColor = true
			str = l.Formatter.Format(param)
		}
		if out.Level == -1 || level == out.Level {
			fmt.Fprintln(out.Out, str)
			l.CheckFileSize(out)
		}
	}
}

func (f *LoggerFormatter) formatter(msg any) string {
	now := time.Now()
	if f.IsColor {
		//要带颜色  error的颜色 为红色 info为绿色 debug为蓝色
		levelColor := f.LevelColor()
		msgColor := f.MsgColor()
		return fmt.Sprintf("%s [msgo] %s %s%v%s | level= %s %s %s | msg=%s %#v %s | fields=%#v",
			yellow, reset, blue, now.Format("2006/01/02 - 15:04:05"), reset,
			levelColor, f.Level.Level(), reset, msgColor, msg, reset, f.LoggerFields,
		)
	}
	return fmt.Sprintf("[msgo] %v | level=%s | msg= %#v | fields=%#v",
		now.Format("2006/01/02 - 15:04:05"),
		f.Level.Level(), msg, f.LoggerFields,
	)
}

func (f *LoggerFormatter) LevelColor() string {
	switch f.Level {
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

func (f *LoggerFormatter) MsgColor() string {
	switch f.Level {
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

func (l *Logger) WithFields(fields Fields) *Logger {
	return &Logger{
		Formatter:    l.Formatter,
		Outs:         l.Outs,
		Level:        l.Level,
		LoggerFields: fields,
	}
}

func (l *Logger) SetLogPath(logPath string) {
	l.LogPath = logPath
	l.Outs = append(l.Outs, &LoggerWriter{
		Level: -1,
		Out:   FileWriter(path.Join(logPath, "all.log")),
	})
	l.Outs = append(l.Outs, &LoggerWriter{
		Level: LevelDebug,
		Out:   FileWriter(path.Join(logPath, "debug.log")),
	})
	l.Outs = append(l.Outs, &LoggerWriter{
		Level: LevelInfo,
		Out:   FileWriter(path.Join(logPath, "info.log")),
	})
	l.Outs = append(l.Outs, &LoggerWriter{
		Level: LevelError,
		Out:   FileWriter(path.Join(logPath, "error.log")),
	})
}

// 判断文件大小
func (l *Logger) CheckFileSize(w *LoggerWriter) {
	logFile := w.Out.(*os.File)
	if logFile != nil {
		stat, err := logFile.Stat()
		if err != nil {
			log.Println(err)
			return
		}
		size := stat.Size()
		if l.LoggerFileSize <= 0 {
			//默认100M
			l.LoggerFileSize = 100 << 20
		}
		if size >= l.LoggerFileSize {
			_, fileName := path.Split(stat.Name())
			//去掉文件名中的后缀
			name := fileName[0:strings.Index(fileName, ".")]
			writer := FileWriter(path.Join(l.LogPath, msstrings.JoinStrings(name, ".", time.Now().UnixMilli(), ".log")))
			if err != nil {
				log.Println("logger checkFileSize error info :", err)
				return
			}
			w.Out = writer
		}
	}
}

func FileWriter(name string) io.Writer {
	w, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	return w
}
