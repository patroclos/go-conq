package main

import (
	"embed"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/patroclos/go-conq"
	"github.com/patroclos/go-conq/aid"
	"github.com/patroclos/go-conq/aid/cmdhelp"
	"github.com/patroclos/go-conq/commander"
	_ "github.com/patroclos/go-conq/example/internal/translations"
	"github.com/patroclos/go-conq/example/unansi"
	"github.com/patroclos/go-conq/getopt"
	"github.com/posener/complete"
)

//go:embed help/*
var helpFs embed.FS

func main() {
	ctx := conq.OSContext()
	root := New()
	com := commander.New(getopt.New(), aid.DefaultHelp)

	err := com.Execute(root, ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

var optPath = conq.ReqOpt[string](conq.O{
	Name:    "path",
	Predict: complete.PredictSet("good", "bad", "ugly"),
})

// the default parser injected by the conq.Opt type supports types implementing encoding.TextUnmarshaler
var OptConfig = conq.Opt[AppConfigFile]{Name: "config"}
var optAddr = conq.Opt[net.IP]{Name: "addr"}
var optCidr = conq.Opt[CIDR]{Name: "cidr"}
var optMime = conq.Opt[MIME]{Name: "mime"}
var optCert = conq.Opt[Cert]{Name: "cert"}
var optCfg = conq.Opt[AppConfigFile]{Name: "config"}
var optPrime = conq.ReqOpt[CryptoPrime]{Name: "prime"}
var optMac = conq.Opt[net.HardwareAddr]{Name: "mac"}

var envDebug = conq.Opt[string]{Name: "CONQ_DEBUG"}

func New() *conq.Cmd {
	helpCmd := cmdhelp.New(helpFs)
	hCmd := cmdhelp.New(helpFs)
	hCmd.Name = "-h"
	return &conq.Cmd{
		Name: "example",
		Opts: []conq.Opter{optPath, optAddr, optCidr, optMime, optCert, optCfg, optPrime, optMac},
		Env:  conq.Opts{envDebug},
		Commands: []*conq.Cmd{
			helpCmd,
			hCmd,
			commander.CmdCompletion,
			{Name: "foo", Commands: []*conq.Cmd{{Name: "baz"}}},
			{Name: "bar"},
			unansi.New(),
		},
		Run: run,
	}
}

func run(c conq.Ctx) error {
	if cfg, err := OptConfig.Get(c); err == nil {
		CurrentConfig = &cfg
	}
	path := optPath.Get(c)
	c.Printer.Fprintf(c.Err, "example: %s: %s\n", optPath.Name, path)

	c.Printer.Fprintf(c.Err, "Configuration has %d profiles.\n", len(CurrentConfig.Config.Profiles))

	days := time.Hour * 25 * 5
	c.Printer.Fprintf(c.Err, "today is: %s\n", days)
	c.Printer.Fprintf(c.Err, "prime: %s\n", optPrime.Get(c).Prime)
	if mac, err := optMac.Get(c); err == nil {
		c.Printer.Fprintf(c.Err, "%v: %#v\n", mac, mac)
	}

	if ip, err := optAddr.Get(c); err == nil {
		c.Printer.Fprintf(c.Err, "IP: %v (private? %v)\n", ip, ip.IsPrivate())
	}

	if debug, err := envDebug.Get(c); err == nil {
		c.Printer.Fprintf(c.Err, "CONQ_DEBUG: %q\n", debug)
	}

	config, _ := optCfg.Get(c)
	for profname := range config.Config.Profiles {
		c.Printer.Fprintf(c.Err, "config profile: %v\n", profname)
	}

	if mime, err := optMime.Get(c); err == nil {
		c.Printer.Fprintf(c.Out, "MIME: %s (%v)\n", mime.Mediatype, mime.Params)
	}

	if cidr, err := optCidr.Get(c); err == nil {
		c.Printer.Fprintf(c.Err, "Network: %v\n", cidr.Net)
		addrs, err := net.InterfaceAddrs()
		if err != nil {
			return fmt.Errorf("failed getting interface addresses: %w", err)
		}
		for _, addr := range addrs {
			ip, _, err := net.ParseCIDR(addr.String())
			if err != nil {
				return fmt.Errorf("failed parsing ip %q: %w", addr.String(), err)
			}
			c.Printer.Fprintf(c.Out, "%v (contained in %v? %v)\n", ip, cidr.Net, cidr.Net.Contains(ip))
			if cidr.Net.Contains(ip) {
			}
		}
	}

	if cert, err := optCert.Get(c); err == nil {
		format := "2006-01-02"
		nb, na := cert.Cert.NotBefore.Format(format), cert.Cert.NotAfter.Format(format)
		c.Printer.Fprintf(c.Out, "Cert %s %v-%v issuer:%s\n", cert.Cert.Subject.CommonName, nb, na, cert.Cert.Issuer.CommonName)
	}
	return nil
}
