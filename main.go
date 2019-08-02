package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"syscall"

	"github.com/sevlyar/go-daemon"

	"github.com/heetch/confita"
	"github.com/heetch/confita/backend"
	"github.com/heetch/confita/backend/env"
	"github.com/heetch/confita/backend/file"
	"github.com/heetch/confita/backend/flags"
)

type Config struct {
	DB         string `config:"db,required"`
	Port       int    `config:"port,required"`
	Socket     string `config:"socket"`
	NoReg      bool   `config:"noreg"`
	Debug      bool   `config:"debug"`
	Foreground bool   `config:"foreground"`
	UseCORS    bool   `config:"cors" json:"cors" yaml:"cors" toml:"cors"`
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

	// this is a workaround for a couple of issues with the config library being
	// used (https://github.com/heetch/confita):
	// 1.- Confita doesn't allow setting prefixes to env vars: since the config
	// variables are too generic (Port, Debug) there is a posibility that an existing
	// ENV var overwrite a config file or flag. This is supposed to be fixed here:
	// https://github.com/heetch/confita/pull/33
	// 2.- Right now confita is not following the right order when loading values
	// from multiple backends, the current behavior is broken with what is supposelly they explain
	// in the docs (first match wins). There is an open issue for this:
	// https://github.com/heetch/confita/issues/64
	//
	// The only two posible solutions are: wait for both issues to be fixed, or
	// switch to a diferrent config library (https://github.com/spf13/viper)
	//
	// In the meantime having "STANDARDFILE_LOAD_CONF_ENV" as an  ENV var flag
	// will maintain BC while allowing to use ENV vars inside a docker container
	if _, ok := os.LookupEnv("STANDARDFILE_LOAD_CONF_ENV"); ok {
		confBackends = append(confBackends, env.NewBackend())
	} else {
		configs := []string{"standardfile.json", "standardfile.toml", "standardfile.yaml"}
		for _, c := range configs {
			if _, err := os.Stat(c); err == nil {
				confBackends = append(confBackends, file.NewBackend(c))
			}
		}
		confBackends = append(confBackends, flags.NewBackend())
	}
	loader := confita.NewLoader(confBackends...)
	err := loader.Load(context.Background(), &cfg)
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	flag.Parse()

	if *ver {
		fmt.Println(`        Version:           ` + Version + `
        Built:             ` + BuildTime + `
        Go Version:        ` + runtime.Version() + `
        OS/Arch:           ` + runtime.GOOS + "/" + runtime.GOARCH + `
        No Registrations:  ` + strconv.FormatBool(cfg.NoReg) + `
        CORS Enabled:      ` + strconv.FormatBool(cfg.UseCORS) + `
        Run in Foreground: ` + strconv.FormatBool(cfg.Foreground) + `
        Webserver Port:    ` + strconv.Itoa(cfg.Port) + `
        CORS Enabled:      ` + strconv.FormatBool(cfg.UseCORS) + `
        DB Path:           ` + cfg.DB + `
        Debug:             ` + strconv.FormatBool(cfg.Debug))
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
