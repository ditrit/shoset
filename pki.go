package shoset

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/square/certstrap/depot"
	"github.com/square/certstrap/pkix"
	"github.com/urfave/cli"
)

type User struct {
	Login    string `json:"login"`
	Password int64  `json:"password"`
}
type Register struct {
	Secret       string `json:"secret"`
	Cert_request string `json:"cert_request"`
}

func (c *Shoset) Init() error {
	// elle sort immédiatement si :
	if c.GetBindAddress() == "" { // il n'y a pas encore eu de bind (bindadress est vide)
		return errors.New("shoset not bound")
	} else if c.ConnsByName.Len() != 0 { // j'ai déjà fait un link ou un join ou j'ai un fichier de configuration (ce qui veut dire que j'ai des connsbyname)
		return errors.New("a protocol already happened on this shoset")
	} else if c.GetIsInit() { // il y a eu déjà un init ou j'ai déjà un certificat (mon certificat existe déjà)
		return errors.New("shoset already initialized")
	}

	c.SetIsInit(true)

	// elle réalise les actions suivantes :
	// 1. La CA
	// créer clef privée et publique CA
	key, err := pkix.CreateRSAKey(4096)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Create RSA Key error:", err)
		os.Exit(1)
	}

	keybytes, _ := key.ExportPrivate()
	private_key, err := os.Create("") ////////////////////////////////////
	if err != nil {
		panic(err)
	}
	_, err = private_key.WriteString(string(keybytes))
	if err != nil {
		panic(err)
	}

	// request de certificat pour la CA (à partir de ces clefs)
	csr, _ := pkix.CreateCertificateSigningRequest(key, "", nil, nil, nil, "", "", "", "", "client") //////////////////
	csrBytes, err := csr.Export()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Print certificate request error:", err)
		os.Exit(1)
	}
	cert_request, err := os.Create("") //////////////////////// fichier temporaire
	if err != nil {
		panic(err)
	}
	cert_request.WriteString(string(csrBytes))

	// autosignature du certificat de la CA
	var crtOut *pkix.Certificate
	var csr *pkix.CertificateSigningRequest

	//We sign the certificate request

	crtbytes, _ := ioutil.ReadFile("./out/ExempleCA.crt")
	crt, _ := pkix.NewCertificateFromPEM(crtbytes)

	keybytes, _ := ioutil.ReadFile("./out/ExempleCA.key")
	key, _ := pkix.NewKeyFromPrivateKeyPEM(keybytes)

	csr, err := pkix.NewCertificateSigningRequestFromPEM([]byte(reg.Cert_request))
	if err != nil {
		fmt.Printf("err")
	}
	expire_time, _ := time.Parse("020106 150405", "220902 050316")
	crtOut, err = pkix.CreateCertificateHost(crt, key, csr, expire_time)

	crtBytes, _ := crtOut.Export()

	// 2. Le certificat de la shoset
	// récupérer le certificat de la CA
	// génération des clefs privée, publique et request pour la shoset
	// création du certificat signé avec la clef privée de la CA

	// 3. Elle associe le rôle 'pki' au nom logique de la shoset

	return nil
}

// pour les shoset ayant le rôle 'pki' :
// 1. Service activé pour les deux fonctions
// getsecret(login, password) => { secret }
func (c *Shoset) GetSecret(login, password string) string {
	if c.GetIsInit() {
		// utiliser login et password
		return uuid.New().String()
	}
	return ""
}

// getCAcert() => { certificat de la CA }
func (c *Shoset) GetCAcert() string {
	if c.GetIsInit() {

	}
	return ""
}

// getCert(certRequest) => { certificat }
func (c *Shoset) GetCert() string {
	if c.GetIsInit() {

	}
	return ""
}

// NewInitCommand sets up an "init" command to initialize a new CA
func NewInitCommand() cli.Command {
	return cli.Command{
		Name:        "init",
		Usage:       "Create Certificate Authority",
		Description: "Create Certificate Authority, including certificate, key and extra information file.",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "passphrase",
				Usage: "Passphrase to encrypt private key PEM block",
			},
			cli.IntFlag{
				Name:  "key-bits",
				Value: 4096,
				Usage: "Size (in bits) of RSA keypair to generate (example: 4096)",
			},
			cli.IntFlag{
				Name:   "years",
				Hidden: true,
			},
			cli.StringFlag{
				Name:  "expires",
				Value: "18 months",
				Usage: "How long until the certificate expires (example: 1 year 2 days 3 months 4 hours)",
			},
			cli.StringFlag{
				Name:  "organization, o",
				Usage: "Sets the Organization (O) field of the certificate",
			},
			cli.StringFlag{
				Name:  "organizational-unit, ou",
				Usage: "Sets the Organizational Unit (OU) field of the certificate",
			},
			cli.StringFlag{
				Name:  "country, c",
				Usage: "Sets the Country (C) field of the certificate",
			},
			cli.StringFlag{
				Name:  "common-name, cn",
				Usage: "Sets the Common Name (CN) field of the certificate",
			},
			cli.StringFlag{
				Name:  "province, st",
				Usage: "Sets the State/Province (ST) field of the certificate",
			},
			cli.StringFlag{
				Name:  "locality, l",
				Usage: "Sets the Locality (L) field of the certificate",
			},
			cli.StringFlag{
				Name:  "key",
				Usage: "Path to private key PEM file (if blank, will generate new key pair)",
			},
			cli.BoolFlag{
				Name:  "stdout",
				Usage: "Print certificate to stdout in addition to saving file",
			},
			cli.StringSliceFlag{
				Name:  "permit-domain",
				Usage: "Create a CA restricted to subdomains of this domain (can be specified multiple times)",
			},
		},
		Action: initAction,
	}
}

func initAction(c *cli.Context) {
	if !c.IsSet("common-name") {
		fmt.Println("Must supply Common Name for CA")
		os.Exit(1)
	}

	formattedName := strings.Replace(c.String("common-name"), " ", "_", -1)

	if depot.CheckCertificate(d, formattedName) || depot.CheckPrivateKey(d, formattedName) {
		fmt.Fprintf(os.Stderr, "CA with specified name \"%s\" already exists!\n", formattedName)
		os.Exit(1)
	}

	var err error
	expires := c.String("expires")
	if years := c.Int("years"); years != 0 {
		expires = fmt.Sprintf("%s %d years", expires, years)
	}

	// Expiry parsing is a naive regex implementation
	// Token based parsing would provide better feedback but
	expiresTime, err := parseExpiry(expires)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid expiry: %s\n", err)
		os.Exit(1)
	}

	var passphrase []byte
	if c.IsSet("passphrase") {
		passphrase = []byte(c.String("passphrase"))
	} else {
		passphrase, err = createPassPhrase()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	var key *pkix.Key
	if c.IsSet("key") {
		keyBytes, err := ioutil.ReadFile(c.String("key"))
		if err != nil {
			fmt.Fprintln(os.Stderr, "Read Key error:", err)
			os.Exit(1)
		}

		key, err = pkix.NewKeyFromPrivateKeyPEM(keyBytes)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Read Key error:", err)
			os.Exit(1)
		}
		fmt.Printf("Read %s\n", c.String("key"))
	} else {
		key, err = pkix.CreateRSAKey(c.Int("key-bits"))
		if err != nil {
			fmt.Fprintln(os.Stderr, "Create RSA Key error:", err)
			os.Exit(1)
		}
		if len(passphrase) > 0 {
			fmt.Printf("Created %s/%s.key (encrypted by passphrase)\n", depotDir, formattedName)
		} else {
			fmt.Printf("Created %s/%s.key\n", depotDir, formattedName)
		}
	}

	crt, err := pkix.CreateCertificateAuthority(key, c.String("organizational-unit"), expiresTime, c.String("organization"), c.String("country"), c.String("province"), c.String("locality"), c.String("common-name"), c.StringSlice("permit-domain"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Create certificate error:", err)
		os.Exit(1)
	}
	fmt.Printf("Created %s/%s.crt\n", depotDir, formattedName)

	if c.Bool("stdout") {
		crtBytes, err := crt.Export()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Print CA certificate error:", err)
			os.Exit(1)
		} else {
			fmt.Println(string(crtBytes))
		}
	}

	if err = depot.PutCertificate(d, formattedName, crt); err != nil {
		fmt.Fprintln(os.Stderr, "Save certificate error:", err)
	}
	if len(passphrase) > 0 {
		if err = depot.PutEncryptedPrivateKey(d, formattedName, key, passphrase); err != nil {
			fmt.Fprintln(os.Stderr, "Save encrypted private key error:", err)
		}
	} else {
		if err = depot.PutPrivateKey(d, formattedName, key); err != nil {
			fmt.Fprintln(os.Stderr, "Save private key error:", err)
		}
	}

	// Create an empty CRL, this is useful for Java apps which mandate a CRL.
	crl, err := pkix.CreateCertificateRevocationList(key, crt, expiresTime)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Create CRL error:", err)
		os.Exit(1)
	}
	if err = depot.PutCertificateRevocationList(d, formattedName, crl); err != nil {
		fmt.Fprintln(os.Stderr, "Save CRL error:", err)
		os.Exit(1)
	}
	fmt.Printf("Created %s/%s.crl\n", depotDir, formattedName)
}
