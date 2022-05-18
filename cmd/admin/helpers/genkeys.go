package helpers

import (
	"calc/common/config"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"os"
)

func NewGenKeysCmd(cfg *config.Auth) *cli.Command {
	return &cli.Command{
		Name:  "genkeys",
		Usage: "creates an x509 private/public keys for auth tokens",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "c",
				Required: false,
			},
		},
		Action: func(ctx *cli.Context) error {
			return GenKeys(cfg.PrivateKeyFile, cfg.PublicKeyFile)
		},
	}
}

// GenKeys creates an x509 private/public keys for auth tokens
func GenKeys(privateKeyFile, publicKeyFile string) error {
	// Generate a new private key.
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// Get a file for the private key information in PEM form.
	privateFile, err := os.Create(privateKeyFile)
	if err != nil {
		return errors.Wrap(err, "creating "+privateKeyFile)
	}

	defer func(privateFile *os.File) {
		_ = privateFile.Close()
	}(privateFile)

	// Construct a PEM block for the private key.
	privateBlock := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	// Write the private key to the private key file.
	if err := pem.Encode(privateFile, &privateBlock); err != nil {
		return errors.Wrap(err, "encoding to "+privateKeyFile)
	}

	// Marshal the public key from the private key to PKIX.
	asn1Bytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return errors.Wrap(err, "marshaling public key")
	}

	// Get a file for the public key information in PEM form.
	publicFile, err := os.Create(publicKeyFile)
	if err != nil {
		return errors.Wrap(err, "creating "+publicKeyFile)
	}

	defer func(privateFile *os.File) {
		_ = privateFile.Close()
	}(privateFile)

	// Construct a PEM block for the public key.
	publicBlock := pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: asn1Bytes,
	}

	// Write the public key to the private key file.
	if err := pem.Encode(publicFile, &publicBlock); err != nil {
		return errors.Wrap(err, "encoding to "+publicKeyFile)
	}

	log.Info().Msgf("private (%s) and public (%s) keys files generated", privateKeyFile, publicKeyFile)

	return nil
}
