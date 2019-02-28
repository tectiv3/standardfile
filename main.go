package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"syscall"

	"github.com/sevlyar/go-daemon"

	"github.com/heetch/confita"
	"github.com/heetch/confita/backend"
	"github.com/heetch/confita/backend/file"
	"github.com/heetch/confita/backend/flags"
)

type Config struct {
	DB         string `config:"db"`
	Port       int    `config:"port"`
	Socket     string `config:"socket"`
	NoReg      bool   `config:"noreg"`
	Debug      bool   `config:"debug"`
	Foreground bool   `config:"foreground"`
	UseCORS    bool   `config:"cors"`
}

var cfg = Config{
	DB:         "sf.db",
	Port:       8888,
	Debug:      false,
	NoReg:      false,
	Foreground: false,
	UseCORS:    false,
}

var (
	signal  = flag.Bool("stop", false, `shutdown server`)
	migrate = flag.Bool("migrate", false, `perform DB migrations`)
	ver     = flag.Bool("v", false, `show version`)
	run     = make(chan bool)
)

// Version string will be set by linker
var Version = "dev"

// BuildTime string will be set by linker
var BuildTime = ""

func init() {
	confBackends := []backend.Backend{}

	configs := []string{"standardfile.json", "standardfile.toml", "standardfile.yaml"}
	for _, c := range configs {
		if _, err := os.Stat(c); err == nil {
			confBackends = append(confBackends, file.NewBackend(c))
		}
	}
	confBackends = append(confBackends, flags.NewBackend())
	loader := confita.NewLoader(confBackends...)
	err := loader.Load(context.Background(), &cfg)
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	flag.Parse()

	if *ver {
		fmt.Println(`        Version:         ` + Version + `
        Built:           ` + BuildTime + `
        Go version:      ` + runtime.Version() + `
        OS/Arch:         ` + runtime.GOOS + "/" + runtime.GOARCH)

		return
	}

	if *migrate {
		Migrate()
		return
	}

	if cfg.Port == 0 {
		cfg.Port = 8888
	}

	if cfg.Foreground {
		worker()
		return
	}

	daemon.AddCommand(daemon.BoolFlag(signal), syscall.SIGTERM, termHandler)

	cntxt := &daemon.Context{
		PidFileName: "pid",
		PidFilePerm: 0644,
		LogFileName: "log",
		LogFilePerm: 0640,
		WorkDir:     "./",
		Umask:       027,
		Args:        nil,
	}

	if len(daemon.ActiveFlags()) > 0 {
		d, err := cntxt.Search()
		if err != nil {
			log.Fatalln("Unable send signal to the daemon:", err)
		}
		log.Println("Stopping server")
		daemon.SendCommands(d)
		return
	}

	d, err := cntxt.Reborn()
	if err != nil {
		log.Fatalln(err)
	}
	if d != nil {
		return
	}
	defer cntxt.Release()

	go worker()

	if err := daemon.ServeSignals(); err != nil {
		log.Println("Error:", err)
	}
}

func termHandler(sig os.Signal) error {
	close(run)
	return daemon.ErrStop
}
