package microProxy

import (
	"flag"

	"github.com/nunoOliveiraqwe/micro-proxy/log"
	"github.com/nunoOliveiraqwe/micro-proxy/server"
)

type Flags struct {
	LogFlags    *log.Flags
	ServerFlags *server.Flags
}

func RegisterFlags() *Flags {
	return &Flags{
		LogFlags:    log.RegisterLogFlags(),
		ServerFlags: server.RegisterServerFlags(),
	}
}

func (f *Flags) ParseFlags() {
	flag.Parse()
}
