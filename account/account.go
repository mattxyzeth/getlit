package account

import (
	"fmt"
	"getlit/config"
	"os"
	"path/filepath"

	"github.com/crcls/lit-go-sdk/auth"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/wallet"
)

type Account struct {
	Address    ethgo.Address
	AuthSig    auth.AuthSig
	Chain      string
	PrivateKey string
}

func New(c *config.Config, conditionType string) (*Account, error) {
	var a *Account
	var err error
	if conditionType == "accesscontrol" || conditionType == "evmcontract" {
		a, err = NewEthereum(c)
		if err != nil {
			return nil, err
		}

	}
	return a, err
}

func NewEthereum(c *config.Config) (*Account, error) {
	if c.PrivateKey == "" {
		return nil, fmt.Errorf("Private key missing. Please call init first.")
	}

	privKey, err := crypto.HexToECDSA(c.PrivateKey)
	if err != nil {
		return nil, err
	}
	wallet := wallet.NewKey(privKey)

	_, derr := os.Stat(filepath.Join(c.WorkingDir, ".getlit", "keyfile"))
	if os.IsNotExist(derr) {
		keyfile, err := os.Create(filepath.Join(c.WorkingDir, ".getlit", "keyfile"))
		if err != nil {
			return nil, err
		}
		keyfile.Close()

		if err := crypto.SaveECDSA(filepath.Join(c.WorkingDir, ".getlit", "keyfile"), privKey); err != nil {
			return nil, err
		}
	}

	fmt.Printf("\x1b[1mEthereum wallet loaded: %s\x1b[0m\n\n", wallet.Address())

	// Generate the AuthSig. Currently, this is not part of the SDK.
	// I figured that the consuming application would want to implement
	// it in a specific way. If you disagree, please open an issue in:
	// https://github.com/crcls/lit-go-sdk/issues
	authSig, err := Siwe(wallet, c.ChainId, "")
	if err != nil {
		panic(err)
	}

	return &Account{
		Address:    wallet.Address(),
		AuthSig:    authSig,
		Chain:      c.Network,
		PrivateKey: c.PrivateKey,
	}, nil
}
