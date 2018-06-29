package main

import (
	"flag"
	"github.com/sevlyar/go-daemon"
	"log"
	"os"
	"syscall"
)

var (
	signal     = flag.Bool("stop", false, `shutdown server`)
	migrate    = flag.Bool("migrate", false, `perform DB migrations`)
	port       = flag.Int("p", 8888, `port to listen on`)
	dbpath     = flag.String("db", "sf.db", `db file location`)
	debug      = flag.Bool("debug", false, `enable debug output`)
	foreground = flag.Bool("foreground", false, `run in foreground`)
	run        = make(chan bool)
)

//VERSION is server version
const VERSION = "v0.3.0"

func main() {
	flag.Parse()

	if *migrate {
		Migrate(*dbpath)
		return
	}

	if *foreground {
		worker(*port, *dbpath)
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

	go worker(*port, *dbpath)

	if err := daemon.ServeSignals(); err != nil {
		log.Println("Error:", err)
	}
}

func termHandler(sig os.Signal) error {
	close(run)
	return daemon.ErrStop
}
