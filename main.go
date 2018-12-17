package main

import (
  "github.com/oaktown/calliope/cmd"
  "io"
  "log"
  "os"
)

type Logger struct {
  Filename string
  Writer io.Writer
}

func NewLogger(filename string) Logger {
  f, _ := os.OpenFile(filename, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
  return Logger{
    Filename: filename,
    Writer: f,
  }
}

func (c Logger) Write(p []byte) (n int, err error) {
  os.Stdout.Write(p)
  f, err := c.Writer.Write(p)
  return f, err
}

func main() {
  logger := NewLogger("calliope.log")
  log.SetOutput(logger)
  cmd.Execute()
}
