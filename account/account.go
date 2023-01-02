package account

import (
	"fmt"
	"os"
	"path/filepath"

	"getlit/config"

	"github.com/crcls/lit-go-sdk/auth"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/wallet"
)

type Account struct {
	Address ethgo.Address
	AuthSig *auth.AuthSig
	Wallet  *wallet.Key
}

func New(c *config.Config) (*Account, error) {
	if c.PrivateKey == "" {
		return nil, fmt.Errorf("Private key missing. Please call init first.")
	}

	privKey, err := crypto.HexToECDSA(c.PrivateKey)
	if err != nil {
		return nil, err
	}

	keyfile, err := os.Create(filepath.Join(c.WorkingDir, ".getlit", "keyfile"))
	if err != nil {
		return nil, err
	}
	keyfile.Close()

	if err := crypto.SaveECDSA(filepath.Join(c.WorkingDir, ".getlit", "keyfile"), privKey); err != nil {
		return nil, err
	}
	wallet := wallet.NewKey(privKey)

	return &Account{
		Address: wallet.Address(),
		Wallet:  wallet,
	}, nil
}
