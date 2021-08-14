package shoset

import (
	"errors"
	"fmt"

	// "io/ioutil"
	"os"

	"time"

	"github.com/google/uuid"
	"github.com/square/certstrap/pkix"
)

func (c *Shoset) Init() error {
	fmt.Println("enter init")
	// elle sort immédiatement si :
	if c.GetBindAddress() == "" { // il n'y a pas encore eu de bind (bindadress est vide)
		return errors.New("shoset not bound")
	} else if c.ConnsByName.Len() != 0 { // j'ai déjà fait un link ou un join ou j'ai un fichier de configuration (ce qui veut dire que j'ai des connsbyname)
		return errors.New("a protocol already happened on this shoset")
	} else if c.GetIsInit() { // il y a eu déjà un init ou j'ai déjà un certificat (mon certificat existe déjà)
		return errors.New("shoset already initialized")
	}
	fmt.Println("conditions ok")

	c.SetIsInit(true)

	// elle réalise les actions suivantes :
	// 1. La CA
	// créer clef privée CA
	CAkey, err := pkix.CreateRSAKey(4096)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Create RSA Key error:", err)
		os.Exit(1)
	}

	CAkeyBytes, _ := CAkey.ExportPrivate()
	dirname, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("gonna create file")
	CAkeyFile, err := os.Create(dirname + "/.shoset/" + c.ConnsByName.GetConfigName() + "/cert/privateCAKey.pem")
	if err != nil {
		panic(err)
	}
	_, err = CAkeyFile.Write(CAkeyBytes)
	if err != nil {
		panic(err)
	}

	// request de certificat pour la CA (à partir de ces clefs)
	var expires string
	if years := 10; years != 0 {
		expires = fmt.Sprintf("%s %d years", expires, years)
	}

	expiresTime, err := parseExpiry(expires)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid expiry: %s\n", err)
		os.Exit(1)
	}
	CAcrt, err := pkix.CreateCertificateAuthority(CAkey, "ditrit", expiresTime, "ditrit", "France", "", "Paris", "CA") // à voir pour des variables d'environnement plus tard
	if err != nil {
		fmt.Fprintln(os.Stderr, "Create certificate error:", err)
		os.Exit(1)
	}

	CAcrtBytes, err := CAcrt.Export()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Print CA certificate error:", err)
		os.Exit(1)
	}

	CAcertFile, err := os.Create(dirname + "/.shoset/" + c.ConnsByName.GetConfigName() + "/cert/CAcert.pem")
	if err != nil {
		panic(err)
	}
	CAcertFile.Write(CAcrtBytes)

	// 2. Le certificat de la shoset
	// récupérer le certificat de la CA
	// génération des clefs privée, publique et request pour la shoset
	// création du certificat signé avec la clef privée de la CA
	hostKey := c.CreateKey()
	hostCsr := c.CreateRequest(hostKey)
	c.SignRequest(CAcrt, hostCsr, hostKey)

	// 3. Elle associe le rôle 'pki' au nom logique de la shoset

	fmt.Println("finish init !!!!!!!!!!")
	return nil
}

// pour les shoset ayant le rôle 'pki' :
// 1. Service activé pour les deux fonctions
// getsecret(login, password) => { secret }
func (c *Shoset) GenerateSecret(login, password string) string {
	if c.GetIsInit() {
		// utiliser login et password
		return uuid.New().String()
	}
	return ""
}

// // getCAcert() => { certificat de la CA }
// func (c *Shoset) GetCAcert() {
// 	if c.GetIsInit() {

// 	}
// }

// // getCert(certRequest) => { certificat }
// func (c *Shoset) GetCert() {
// 	if c.GetIsInit() {
// 	}
// }

func (c *Shoset) CreateKey() *pkix.Key {
	key, err := pkix.CreateRSAKey(4096)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Create RSA Key error:", err)
		os.Exit(1)
	}

	keyBytes, _ := key.ExportPrivate()
	dirname, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err)
	}
	keyFile, err := os.Create(dirname + "/.shoset/" + c.ConnsByName.GetConfigName() + "/cert/privateKey.pem")
	if err != nil {
		panic(err)
	}
	_, err = keyFile.Write(keyBytes)
	if err != nil {
		panic(err)
	}
	return key
}

func (c *Shoset) CreateRequest(hostKey *pkix.Key) *pkix.CertificateSigningRequest {
	// var hostCsr *pkix.CertificateSigningRequest

	//We sign the certificate request
	hostKeyBytes, _ := hostKey.ExportPrivate()
	// hostKeyFile, err := os.Create("") ////////////////////////////////////
	// if err != nil {
	// 	panic(err)
	// }
	// _, err = hostKeyFile.WriteString(string(hostKeyBytes))
	// if err != nil {
	// 	panic(err)
	// }

	hostCsr, err := pkix.NewCertificateSigningRequestFromPEM(hostKeyBytes)
	if err != nil {
		fmt.Printf("err")
	}
	return hostCsr

}

func (c *Shoset) SignRequest(CAcrt *pkix.Certificate, hostCsr *pkix.CertificateSigningRequest, hostKey *pkix.Key) {
	// var hostCert *pkix.Certificate
	// var err error
	expire_time, _ := time.Parse("020106 150405", "220902 050316")
	fmt.Println("cacert", CAcrt, "hostkey", hostKey, "hostcsr", hostCsr, "expiretime", expire_time)
	hostCert, err := pkix.CreateCertificateHost(CAcrt, hostKey, hostCsr, expire_time)

	hostCrtBytes, _ := hostCert.Export()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Print certificate error:", err)
		os.Exit(1)
	}
	dirname, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err)
	}
	hostCertFile, err := os.Create(dirname + "/.shoset/" + c.ConnsByName.GetConfigName() + "/cert/cert.pem")
	if err != nil {
		panic(err)
	}
	hostCertFile.Write(hostCrtBytes)
}
