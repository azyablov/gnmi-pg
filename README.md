# Experimental implementation of gNMI utilities



## gNMI Capability
```sh
Usage of ./gnmi-cap:
  -addr string
        The target address in the format of host[:port], by default port is 57400.
  -cert string
        Client certificate file in PEM format.
  -hostname string
        The target hostname used to verify the hostname returned by TLS handshake.
  -insecure
        Insecure connection.
  -key string
        Client private key file.
  -password string
        The password to authenticate against target. (default "admin")
  -rootCA string
        CA certificate file in PEM format.
  -skip_verify
        Diable certificate validation during TLS session ramp-up.
  -timeout duration
        Connection timeout. (default 10s)
  -username string
        The username to authenticate against target. (default "admin")
```
## gNMI Get
```sh
Usage of ./gnmi-get:
  -addr string
        The target address in the format of host[:port], by default port is 57400.
  -cert string
        Client certificate file in PEM format.
  -dtype int
        GetRequest_ALL    GetRequest_DataType = 0 // All data elements.
                GetRequest_CONFIG GetRequest_DataType = 1 // Config (rw) only elements.
                GetRequest_STATE  GetRequest_DataType = 2 // State (ro) only elements.
                // Data elements marked in the schema as operational. This refers to data
                // elements whose value relates to the state of processes or interactions
                // running on the device.
                GetRequest_OPERATIONAL GetRequest_DataType = 3)
  -encoding int
        the encoding that the target should utilise to serialise the subtree of the data tree requested. Possible values:
                Encoding_JSON      Encoding = 0 // JSON encoded text.
                Encoding_BYTES     Encoding = 1 // Arbitrarily encoded bytes.
                Encoding_PROTO     Encoding = 2 // Encoded according to scalar values of TypedValue.
                Encoding_ASCII     Encoding = 3 // ASCII text of an out-of-band agreed format.
                Encoding_JSON_IETF Encoding = 4 // JSON encoded text as per RFC7951. (default 4)
  -hostname string
        The target hostname used to verify the hostname returned by TLS handshake.
  -insecure
        Insecure connection.
  -key string
        Client private key file.
  -password string
        The password to authenticate against target. (default "admin")
  -prefix string
        The prefix is applied to all paths within the GetRequest message.
  -rootCA string
        CA certificate file in PEM format.
  -skip_verify
        Diable certificate validation during TLS session ramp-up.
  -timeout duration
        Connection timeout. (default 10s)
  -username string
        The username to authenticate against target. (default "admin")
  -xpath value
        The prefix is applied to all paths within the GetRequest message.
```