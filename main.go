package main

import (
	"flag"
	"github.com/sevlyar/go-daemon"
	"log"
	"os"
	"syscall"
)

const DEBUG = true

var (
	signal = flag.String("s", "", `stop — shutdown
reload — reloading the configuration file`)
	run = make(chan bool)
)

func main() {
	flag.Parse()
	daemon.AddCommand(daemon.StringFlag(signal, "stop"), syscall.SIGTERM, termHandler)
	daemon.AddCommand(daemon.StringFlag(signal, "reload"), syscall.SIGHUP, reloadHandler)

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

func reloadHandler(sig os.Signal) error {
	log.Println("configuration reloaded")
	return nil
}
