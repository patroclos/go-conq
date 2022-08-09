package main

import (
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"mime"
	"net"
	"os"
	"strconv"

	"github.com/patroclos/go-conq"
	"github.com/patroclos/go-conq/aid"
	"github.com/patroclos/go-conq/commander"
	"github.com/patroclos/go-conq/getopt"
	"github.com/posener/complete"
	"gopkg.in/yaml.v2"
)

type O = conq.O

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

type AppConfig struct {
	Bind      string   `yaml:"bind"`
	Endpoints []net.IP `yaml:"endpoints"`
}

func (c *AppConfig) UnmarshalText(txt []byte) error {
	file, err := os.OpenFile(string(txt), os.O_RDONLY, 0)
	if err != nil {
		return fmt.Errorf("failed opening file: %w", err)
	}
	defer file.Close()
	bytes, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed reading bytes from file: %w", err)
	}

	return yaml.Unmarshal(bytes, c)
}

func main() {
	optPath := conq.ReqOpt[string](conq.O{
		Name:    "path",
		Predict: complete.PredictSet("good", "bad", "ugly"),
	})
	// the default parser injected by the conq.Opt type supports types implementing encoding.TextUnmarshaler
	optAddr := conq.Opt[net.IP](O{Name: "addr"})
	optCidr := conq.Opt[CIDR](O{Name: "cidr"})
	optMime := conq.Opt[MIME](O{Name: "mime"})
	optCert := conq.Opt[Cert](O{Name: "cert"})
	optCfg := conq.Opt[AppConfig](O{Name: "config"})
	optPrime := conq.ReqOpt[CryptoPrime](O{Name: "prime"})
	optMac := conq.Opt[net.HardwareAddr](O{Name: "mac"})
	testCmd := &conq.Cmd{
		Name: "example",
		Opts: []conq.Opter{optPath, optAddr, optCidr, optMime, optCert, optCfg, optPrime, optMac},
		Run: func(c conq.Ctx) error {
			path := optPath.Get(c)
			fmt.Fprintf(c.Err, "example: %s: %s\n", optPath.Name, path)

			fmt.Fprintf(c.Err, "prime: %s\n", optPrime.Get(c))
			if mac, err := optMac.Get(c); err == nil {
				fmt.Fprintf(c.Err, "%v: %#v\n", mac, mac)
			}

			if ip, err := optAddr.Get(c); err == nil {
				fmt.Fprintf(c.Err, "IP: %v (private? %v)\n", ip, ip.IsPrivate())
			}

			config, _ := optCfg.Get(c)
			for _, ep := range config.Endpoints {
				fmt.Fprintf(c.Err, "config endoint: %v (private? %v)\n", ep, ep.IsPrivate())
			}

			if mime, err := optMime.Get(c); err == nil {
				fmt.Fprintf(c.Out, "MIME: %s (%v)\n", mime.Mediatype, mime.Params)
			}

			if cidr, err := optCidr.Get(c); err == nil {
				fmt.Fprintf(c.Err, "Network: %v\n", cidr.Net)
				addrs, err := net.InterfaceAddrs()
				if err != nil {
					return fmt.Errorf("failed getting interface addresses: %w", err)
				}
				for _, addr := range addrs {
					ip, _, err := net.ParseCIDR(addr.String())
					if err != nil {
						return fmt.Errorf("failed parsing ip %q: %w", addr.String(), err)
					}
					fmt.Fprintf(c.Out, "%v (contained in %v? %v)\n", ip, cidr.Net, cidr.Net.Contains(ip))
					if cidr.Net.Contains(ip) {
					}
				}
			}

			if cert, err := optCert.Get(c); err == nil {
				format := "2006-01-02"
				nb, na := cert.Cert.NotBefore.Format(format), cert.Cert.NotAfter.Format(format)
				fmt.Fprintf(c.Out, "Cert %s %v-%v issuer:%s\n", cert.Cert.Subject.CommonName, nb, na, cert.Cert.Issuer.CommonName)
			}
			return nil
		},
		Commands: []*conq.Cmd{
			aid.HelpCommand(),
			commander.CmdCompletion,
			{Name: "foo", Commands: []*conq.Cmd{{Name: "baz"}}},
			{Name: "bar"},
		},
	}
	err := commander.New(getopt.New(), nil).Execute(testCmd, conq.OSContext())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
