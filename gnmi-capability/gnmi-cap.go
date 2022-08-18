package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"flag"

	"github.com/azyablov/gnmi-pg/gnmilib"
	"github.com/golang/protobuf/proto"
	"github.com/openconfig/gnmi/proto/gnmi"
	"google.golang.org/grpc"
	// log "github.com/golang/glog" // TODO: check to implement, if relevant
)

var (
	// Certificate/key and rootca
	rootCA = flag.String("rootCA", "", "CA certificate file in PEM format.")
	cert   = flag.String("cert", "", "Client certificate file in PEM format.")
	key    = flag.String("key", "", "Client private key file.")

	// Credentials
	username = flag.String("username", "admin", "The username to authenticate against target.")
	password = flag.String("password", "admin", "The password to authenticate against target.")

	// gNMI target connectivity options
	targetHostname = flag.String("hostname", "", "The target hostname used to verify the hostname returned by TLS handshake.")
	targetAddr     = flag.String("addr", "", "The target address in the format of host[:port], by default port is 57400.")

	// Connection options
	insecConn  = flag.Bool("insecure", false, "Insecure connection.")
	skipVerify = flag.Bool("skip_verify", false, "Diable certificate validation during TLS session ramp-up.")
	timeout    = flag.Duration("timeout", 10*time.Second, "Connection timeout.")
)

const (
	targetPort = 57400
)

func main() {

	// Parsing flags
	flag.Parse()
	if len(*targetAddr) == 0 {
		flag.Usage()
		log.Fatalf("addr is manadatory to provide")
	}
	// Setting up grpc options
	dOpts, err := gnmilib.SetupGNMISecureTransport(
		gnmilib.TLSInit{
			InsecConn:      *insecConn,
			SkipVerify:     *skipVerify,
			TargetHostname: *targetHostname,
			RootCA:         *rootCA,
			Cert:           *cert,
			Key:            *key,
		})
	if err != nil {
		log.Fatal(err)
	}
	// Set up a connection to the server.
	var t string // target address to connect using grpc.Dial()
	if strings.Contains(*targetAddr, ":") {
		t = *targetAddr
	} else {
		t = fmt.Sprintf("%s:%v", *targetAddr, targetPort)
	}

	// Dialing ...
	gRPCconn, err := grpc.Dial(t, *dOpts...)
	if err != nil {
		log.Fatalf("can't not connect to the host %s due to %s", t, err)
	}
	defer gRPCconn.Close()

	// Creating context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	// Attaching credentiasl to the context
	uc := gnmilib.UserCredentials{
		Username: *username,
		Password: *password,
	}
	ctx, err = gnmilib.PopulateMDCredentials(ctx, uc)
	if err != nil {
		log.Fatalln(err)
	}

	gNMIC := gnmi.NewGNMIClient(gRPCconn)

	resp, err := gNMIC.Capabilities(ctx, &gnmi.CapabilityRequest{})
	if err != nil {
		log.Fatalf("can't get capabilities: %s", err)
	}

	// Printing message in text form into STDOUT
	fmt.Printf("Capabilities Response:\n%v", proto.MarshalTextString(resp))

}
