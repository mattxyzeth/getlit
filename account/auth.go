package account

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/crcls/lit-go-sdk/auth"
	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/wallet"
)

const PREFIX_191 = "\x19Ethereum Signed Message:\n"

func EIP191(msg string) (msgBytes []byte) {
	msgLen := len(msg)
	msgBytes = append(msgBytes, []byte(PREFIX_191)...)
	msgBytes = append(msgBytes, []byte(strconv.FormatInt(int64(msgLen), 10))...)
	msgBytes = append(msgBytes, []byte(msg)...)
	return
}

func EIP4361(address ethgo.Address, msg, chain, nonce, date string) string {
	return fmt.Sprintf(`GetLit wants you to sign in with your Ethereum account:
%s

%s
URI: http://example.com
Version: 1
Chain ID: %s
Nonce: %s
Issued At: %s`, address, msg, chain, nonce, date)
}

func (a *Account) Siwe(chain, msg string) (*auth.AuthSig, error) {
	if a.AuthSig != nil {
		return a.AuthSig, nil
	}

	date := time.Now()

	eip4361 := EIP4361(a.Address, msg, chain, strconv.FormatInt(date.Unix(), 10), date.Format(time.RFC3339))
	msgBytes := EIP191(eip4361)

	sig, err := a.Wallet.SignMsg(msgBytes)
	if err != nil {
		return nil, err
	}

	authSig := &auth.AuthSig{
		Address:       a.Address.String(),
		DerivedVia:    "ethgo.Key.SignMsg",
		SignedMessage: eip4361,
		Sig:           "0x" + hex.EncodeToString(sig),
	}

	a.AuthSig = authSig

	return authSig, nil
}

func RecoverAddress(signature, plaintext string) (addr ethgo.Address, err error) {
	var sig []byte
	sig, err = hex.DecodeString(signature[2:] /* Remove 0x */)
	if err != nil {
		fmt.Println("AuthSig failed to hex decode sig", err)
		return
	}

	if sig[len(sig)-1] == 28 {
		sig[len(sig)-1] = 1
	}

	// fmt.Printf("%s\n", string(EIP191(plaintext)))

	addr, err = wallet.EcrecoverMsg(EIP191(plaintext), sig)
	if err != nil {
		fmt.Println("AuthSig verify failed", err)
		return
	}

	return
}
