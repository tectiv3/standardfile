package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/sevlyar/go-daemon"

	"github.com/heetch/confita"
	"github.com/heetch/confita/backend"
	"github.com/heetch/confita/backend/env"
	"github.com/heetch/confita/backend/file"
	"github.com/heetch/confita/backend/flags"
)

type config struct {
	DB         string `config:"db"`
	Port       int    `config:"port"`
	Socket     string `config:"socket"`
	NoReg      bool   `config:"noreg"`
	Debug      bool   `config:"debug"`
	Foreground bool   `config:"foreground"`
	UseCORS    bool   `config:"cors" json:"cors" yaml:"cors" toml:"cors"`
}

var cfg = config{
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
	cfgPath = flag.String("c", ".", `config file location`)
	run     = make(chan bool)
)

var loadedConfig = "using flags"

// Version string will be set by linker
var Version = "dev"

// BuildTime string will be set by linker
var BuildTime = "N/A"

func init() {
	confBackends := []backend.Backend{}
	cfgPath = getConfigFlag()

	if *cfgPath == "env" {
		confBackends = append(confBackends, env.NewBackend())
		loadedConfig = "using environment"
	} else {
		configs := []string{"standardfile.json", "standardfile.toml", "standardfile.yaml"}
		path := strings.TrimRight(*cfgPath, string(os.PathSeparator))
		fileName := ""
		if info, err := os.Stat(path); err == nil {
			if !info.IsDir() {
				fileName = filepath.Base(path)
			}
		}
		if len(fileName) > 1 {
			path = filepath.Dir(path)
			configs = append([]string{fileName}, configs...)
		}
		for _, c := range configs {
			if _, err := os.Stat(path + "/" + c); err == nil {
				loadedConfig = path + "/" + c
				confBackends = append(confBackends, file.NewBackend(loadedConfig))
				break
			}
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
		socket := "no"
		if len(cfg.Socket) > 0 {
			socket = cfg.Socket
		}
		fmt.Println(`        Version:           ` + Version + `
        Built:             ` + BuildTime + `
        Go Version:        ` + runtime.Version() + `
        OS/Arch:           ` + runtime.GOOS + "/" + runtime.GOARCH + `
        Loaded Config:     ` + loadedConfig + `
        No Registrations:  ` + strconv.FormatBool(cfg.NoReg) + `
        CORS Enabled:      ` + strconv.FormatBool(cfg.UseCORS) + `
        Run in Foreground: ` + strconv.FormatBool(cfg.Foreground) + `
        Webserver Port:    ` + strconv.Itoa(cfg.Port) + `
        Socket:            ` + socket + `
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

func getConfigFlag() *string {
	cfg := "."
	args := os.Args[1:]
	if len(args) == 0 {
		return &cfg
	}
	s := args[0]
	if len(s) < 2 || s[0] != '-' {
		return &cfg
	}
	numMinuses := 1
	if s[1] == '-' {
		numMinuses++
		if len(s) == 2 { // "--" terminates the flags
			return &cfg
		}
	}

	name := s[numMinuses:]
	if len(name) == 0 || name[0] == '-' || name[0] == '=' {
		log.Fatalf("bad flag syntax: %s", s)
		return &cfg
	}

	args = args[1:]
	hasValue := false
	value := ""

	// possible to set it as -config=value or -c=value
	for i := 1; i < len(name); i++ {
		if name[i] == '=' {
			value = name[i+1:]
			hasValue = true
			name = name[0:i]
			break
		}
	}

	//check if flag name is config
	if name != "c" && name != "config" {
		return &cfg
	}

	if !hasValue && len(args) > 0 {
		// value is the next arg
		hasValue = true
		value = args[0]
	}

	cfg = value
	return &cfg
}
