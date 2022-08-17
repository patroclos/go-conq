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
)

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
