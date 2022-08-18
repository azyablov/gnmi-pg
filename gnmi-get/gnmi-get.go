package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/azyablov/gnmi-pg/gnmilib"
	"github.com/google/gnxi/utils/xpath"
	"github.com/openconfig/gnmi/proto/gnmi"
	"google.golang.org/grpc"
)

type multiFlag struct {
	flags []string
}

func (f *multiFlag) String() string {
	return fmt.Sprintf("%s", f.flags)
}

func (f *multiFlag) Set(s string) error {
	f.flags = append(f.flags, s)
	return nil
}

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

	// Get message params
	targetPrefix   = flag.String("prefix", "", "The prefix is applied to all paths within the GetRequest message.")
	xPathList      multiFlag // Flag to store all required xpaths to retrieve
	targetEncoding = flag.Int("encoding", 4,
		`the encoding that the target should utilise to serialise the subtree of the data tree requested. Possible values:
	Encoding_JSON      Encoding = 0 // JSON encoded text.
	Encoding_BYTES     Encoding = 1 // Arbitrarily encoded bytes.
	Encoding_PROTO     Encoding = 2 // Encoded according to scalar values of TypedValue.
	Encoding_ASCII     Encoding = 3 // ASCII text of an out-of-band agreed format.
	Encoding_JSON_IETF Encoding = 4 // JSON encoded text as per RFC7951.`)
	dataType = flag.Int("dtype", 0,
		`GetRequest_ALL    GetRequest_DataType = 0 // All data elements.
	GetRequest_CONFIG GetRequest_DataType = 1 // Config (rw) only elements.
	GetRequest_STATE  GetRequest_DataType = 2 // State (ro) only elements.
	// Data elements marked in the schema as operational. This refers to data
	// elements whose value relates to the state of processes or interactions
	// running on the device.
	GetRequest_OPERATIONAL GetRequest_DataType = 3)`)
	// TODO: add , path, encoding, data type
)

const (
	targetPort = 57400
)

func init() {
	flag.Var(&xPathList, "xpath", "The prefix is applied to all paths within the GetRequest message.")
}

func main() {

	// Parsing flags and validating inputs
	flag.Parse()
	if len(*targetAddr) == 0 {
		flag.Usage()
		log.Fatalf("addr is manadatory to provide")
	}

	if len(xPathList.flags) == 0 {
		flag.Usage()
		log.Fatalf("target xpath should be provided")
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

	// Verifing that prefix is following path encoding conventions
	prefix, err := xpath.ToGNMIPath(*targetPrefix)
	if err != nil {
		log.Fatalf("prefix is not valid xpath: %s", err)
	}

	// Looping over requested xpath list
	var pathList []*gnmi.Path // xpaths to add into GetRequest
	for _, xp := range xPathList.flags {
		gPath, err := xpath.ToGNMIPath(xp)
		if err != nil {
			log.Fatalf("can't parse provided xpath: %q", gPath)
		}
		pathList = append(pathList, gPath)
	}

	// Constructing Get Request
	getReq := gnmi.GetRequest{
		Prefix:   prefix,
		Path:     pathList,
		Type:     gnmi.GetRequest_DataType(*dataType),
		Encoding: gnmi.Encoding(*targetEncoding),
	}

	getResp, err := gNMIC.Get(ctx, &getReq)
	if err != nil {
		log.Fatalf("able to execute get request: %v", err)
	}

	// Printing message in text form into STDOUT
	fmt.Printf("Get Response getResp.String():\n%s\n\n", getResp.String())

	fmt.Printf("Get Response value getResp.Notification[0].Update[0].Val:\n%s\n\n", getResp.Notification[0].Update[0].Val)

	respVal := getResp.Notification[0].Update[0].Val.GetJsonIetfVal()
	fmt.Printf("Get Response value getResp.Notification[0].Update[0].Val.GetJsonIetfVal():\n%s\n\n", string(respVal))
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, respVal, "", "    "); err != nil {
		log.Fatalf("can't ident provided JSON: %s", err)
	}

	fmt.Println(prettyJSON.String())

}
