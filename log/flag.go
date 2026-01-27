package log

import "flag"

type Flags struct {
	Debug    bool
	LogPath  string
	LogLevel string
}

func RegisterLogFlags() *Flags {
	logFlags := &Flags{}
	flag.BoolVar(&logFlags.Debug, "debug", false, "enable debug logging")
	flag.StringVar(&logFlags.LogPath, "log-path", "", "path to log file")
	flag.StringVar(&logFlags.LogLevel, "log-level", "INFO", "log level")
	return logFlags
}
