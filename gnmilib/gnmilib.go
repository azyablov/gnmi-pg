package gnmilib

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type TLSInit struct {
	InsecConn      bool
	SkipVerify     bool
	TargetHostname string
	RootCA         string
	Cert           string
	Key            string
}

type UserCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func SetupGNMISecureTransport(t TLSInit) (*[]grpc.DialOption, error) {
	dOpts := []grpc.DialOption{}
	if t.InsecConn {
		// Insecure connection setup
		dOpts = append(dOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		return &dOpts, nil
	} else {
		tlsConfig := tls.Config{}
		// Applying skipVerify
		tlsConfig.InsecureSkipVerify = t.SkipVerify
		if !t.SkipVerify {
			tlsConfig.ServerName = t.TargetHostname
			if len(t.RootCA) == 0 || len(t.Cert) == 0 || len(t.Key) == 0 {
				return &dOpts, fmt.Errorf("one of more files for rootCA / certificate / key are not specified")
			}

			// Populating root CA certificates pool
			fh, err := os.Open(t.RootCA)
			if err != nil {
				return nil, fmt.Errorf("populating root CA certificates pool: %s", err)
			}
			bs, err := ioutil.ReadAll(fh)
			if err != nil {
				return nil, fmt.Errorf("reading root CA cert: %s", err)
			}

			certCAPool := x509.NewCertPool()
			if !certCAPool.AppendCertsFromPEM(bs) {
				return nil, errors.New("can't load PEM file for rootCAt")
			}
			tlsConfig.RootCAs = certCAPool

			// Loading certificate
			certTLS, err := tls.LoadX509KeyPair(t.Cert, t.Key)
			if err != nil {
				return nil, fmt.Errorf("can't load certificate keypair: %s", err)
			}
			// Leaf is the parsed form of the leaf certificate, which may be initialized
			// using x509.ParseCertificate to reduce per-handshake processing.
			certTLS.Leaf, err = x509.ParseCertificate(certTLS.Certificate[0])
			if err != nil {
				return nil, fmt.Errorf("cert parsing error: %s", err)
			}
			tlsConfig.Certificates = []tls.Certificate{certTLS}

			// Setting minimum version for TLS1.2 in accordance with specification
			tlsConfig.MinVersion = tls.VersionTLS12

		}
		dOpts = append(dOpts, grpc.WithTransportCredentials(credentials.NewTLS(&tlsConfig)))
		return &dOpts, nil
	}
}

func PopulateMDCredentials(ctx context.Context, uc UserCredentials) (context.Context, error) {
	// Specification - https://github.com/openconfig/reference/blob/master/rpc/gnmi/gnmi-authentication.md
	if len(uc.Username) == 0 {
		return ctx, errors.New("populateCredentials: username must be provided")

	}

	// Retrieve MD from outgoing context
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}

	// Populating
	md.Set("username", uc.Username)
	if len(uc.Password) != 0 {
		md.Set("password", uc.Password)
	}

	// Creating new context with new MD or updated MD attached to it.
	return metadata.NewOutgoingContext(ctx, md), nil
}
