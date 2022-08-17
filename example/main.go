package main

import (
	"crypto/rand"
	"crypto/x509"
	"embed"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"mime"
	"net"
	"os"
	"strconv"
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

type CIDR struct {
	IP  net.IP
	Net *net.IPNet
}

func (c *CIDR) UnmarshalText(txt []byte) error {
	ip, net, err := net.ParseCIDR(string(txt))
	c.IP = ip
	c.Net = net
	return err
}

type MIME struct {
	Mediatype string
	Params    map[string]string
}

func (x *MIME) UnmarshalText(txt []byte) error {
	mt, params, err := mime.ParseMediaType(string(txt))
	x.Mediatype = mt
	x.Params = params
	return err
}

type Cert struct {
	Cert *x509.Certificate
}

type CryptoPrime struct {
	Prime *big.Int
}

func (p *CryptoPrime) UnmarshalText(txt []byte) error {
	bits, err := strconv.Atoi(string(txt))
	if err != nil {
		return fmt.Errorf("invalid bit count %q: %w", txt, err)
	}
	prime, err := rand.Prime(rand.Reader, bits)
	if err != nil {
		return fmt.Errorf("failed to create prime with %d bits: %w", bits, err)
	}
	p.Prime = prime
	return nil
}

func (x *Cert) UnmarshalText(txt []byte) error {
	// get cert from file
	file, err := os.OpenFile(string(txt), os.O_RDONLY, 0)
	if err != nil {
		crt, _ := pem.Decode(txt)
		if crt == nil {
			return fmt.Errorf("failed parsing PEM block %q", txt)
		}
		cert, err := x509.ParseCertificate(crt.Bytes)
		x.Cert = cert
		return err
	}
	defer file.Close()
	bytes, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed reading content of %q: %w", txt, err)
	}

	crt, _ := pem.Decode(bytes)
	if crt == nil {
		return fmt.Errorf("failed parsing PEM block: %q", bytes)
	}
	cert, err := x509.ParseCertificate(crt.Bytes)
	x.Cert = cert
	return err
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
