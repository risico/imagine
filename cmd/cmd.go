package cmd

import (
	"github.com/juju/errors"

	"github.com/risico/imagine/src/server"
)

var CLI struct {
	Start StartCMD `cmd:"" json:"start" help:"Starts the imagine application"`
}

type StartCMD struct {
	Hostname string `help:"sets the hostname for the server" json:"hostname"`
	Port     int    `help:"sets the application port" json:"port"`
}

func (r *StartCMD) Run() error {
	params := server.ServerParams{
		Hostname: r.Hostname,
		Port:     r.Port,
	}
	s := server.NewServer(params)
	return errors.Trace(s.Start())
}
