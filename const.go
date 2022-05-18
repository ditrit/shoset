package shoset

// empty string
const (
	VOID string = ""
)

// protocol
const (
	CONNECTION_TYPE string = "tcp"

	// join
	PROTOCOL_JOIN     string = "join"
	AKNOWLEDGE_JOIN   string = "aknowledge_join"
	UNAKNOWLEDGE_JOIN string = "unaknowledge_join"
	MEMBER            string = "member"

	// link
	PROTOCOL_LINK string = "link"
	BROTHERS      string = "brothers"

	// bye
	PROTOCOL_EXIT string = "bye"
	DELETE        string = "delete"
)

// direction
const (
	OUT string = "out"
	IN  string = "in"
	ME  string = "me" // create conn without real direction - see it as a fake connection between 2 shosets from the same type in order to get them know each other
)

// IP
const (
	DEFAULT_IP string = "0.0.0.0:"  // analyze all network traffic - "no specific address"
	LOCALHOST  string = "127.0.0.1" // loopback
)

// key
const (
	ALL      string = "all" // doesn't restrict key
	RELATIVE string = "me"  // restrict keys to relatives from fake connection between 2 shosets from the same type
)

// encode type
const (
	CERTIFICATE     string = "CERTIFICATE"
	RSA_PRIVATE_KEY string = "RSA PRIVATE KEY"
)

// path
const (
	PATH_CA_CERT        string = "/cert/CAcert.crt"
	PATH_CERT           string = "/cert/cert.crt"
	PATH_CA_PRIVATE_KEY string = "/cert/privateCAKey.key"
	PATH_PRIVATE_KEY    string = "/cert/privateKey.key"
	PATH_CONFIG_DIR     string = "/config/"
	PATH_CERT_DIR       string = "/cert/"
)

// TLS
const (
	// TLS double way communication - established protocols
	TLS_DOUBLE_WAY_TEST_WRITE string = "testTLSdoubleWayWrite"
	TLS_DOUBLE_WAY_PKI_EVT    string = "pkievt_TLSdoubleWay"

	// TLS single way communication - certificate request
	TLS_SINGLE_WAY_PKI_EVT        string = "pkievt_TLSsingleWay"
	TLS_SINGLE_WAY_RETURN_PKI_EVT string = "return_pkievt_TLSsingleWay"
)

const (
	// CERT_FILE_ENV is the environment variable which identifies where to locate
	// the SSL certificate file. If set this overrides the system default.
	CERT_FILE_ENV = "SSL_CERT_FILE"
)

// viper
const (
	CONFIG_TYPE string = "yaml"
)

// logger
const (
	INFO  string = "info"
	TRACE string = "trace"
	DEBUG string = "debug"
	WARN  string = "warn"
	ERROR string = "error"
)
