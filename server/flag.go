package server

import "flag"

type Flags struct {
	Port             int
	Host             string
	IsHTTPS          bool
	certFile         string
	keyFile          string
	IdleTimeoutSecs  int
	ReadTimeoutSecs  int
	WriteTimeoutSecs int
}

func RegisterServerFlags() *Flags {
	flags := Flags{}
	flag.StringVar(&flags.Host, "host", "127.0.0.1", "Host to listen on")
	flag.IntVar(&flags.Port, "port", 27000, "Port to listen on")
	flag.BoolVar(&flags.IsHTTPS, "https", false, "Enable HTTPS")
	flag.StringVar(&flags.certFile, "cert-file", "", "Path to the certificate file. Ignored if --https is not set.")
	flag.StringVar(&flags.keyFile, "key-file", "", "Path to the key file. Ignored if --https is not set.")
	flag.IntVar(&flags.IdleTimeoutSecs, "idle-timeout", 60, "Idle timeout in seconds")
	flag.IntVar(&flags.ReadTimeoutSecs, "read-timeout", 60, "Read timeout in seconds")
	flag.IntVar(&flags.WriteTimeoutSecs, "write-timeout", 60, "Write timeout in seconds")
	return &flags
}
