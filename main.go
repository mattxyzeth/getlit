package main

import (
	"bufio"
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"getlit/account"
	"getlit/config"
	"getlit/ethereum"
	"io"
	"os"
	"path/filepath"
	"strings"

	"getlit/jsonUtils"

	"github.com/crcls/lit-go-sdk/client"
	"github.com/crcls/lit-go-sdk/conditions"
	"github.com/crcls/lit-go-sdk/crypto"
)

var CMDS = map[string]string{
	"init":    "Initialize the CLI and wallet account.",
	"encrypt": "Encrypt data and store the Lit conditions with the network.",
	"decrypt": "Retrieve the symmetric key and decrypt the ciphertext.",
	"help":    "Prints the help context.",
}

var (
	conditionType = flag.String("type", "evmcontract", "Choose the condition type you'd like to use.")
)

func initialized() bool {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	var hasPk bool
	var hasConfig bool

	if _, err := os.Stat(filepath.Join(wd, ".getlit", "keyfile")); err == nil {
		hasPk = true
	}

	if _, err := os.Stat(filepath.Join(wd, ".getlit", "config.yml")); err == nil {
		hasConfig = true
	}

	return hasPk && hasConfig
}

func main() {
	flag.Parse()

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "help":
			fmt.Println("")
			fmt.Println("Usage: bui command [OPTIONS]")
			fmt.Println("Options:")
			fmt.Println("  -type\t\t*required* Set the condition type you'd like to use.\n\t\tValid options are:\n\t\t  accesscontrol\n\t\t  evmcontract\n\t\t  solrpc\n\t\t  unified")
			fmt.Println("")
			fmt.Println("Commands:")

			for cmd, description := range CMDS {
				fmt.Printf("  %s\t%s\n", cmd, description)
			}
			fmt.Println("")
		case "init":
			if initialized() {
				fmt.Println("Already initialized.")
				return
			}

			reader := bufio.NewReader(os.Stdin)
			fmt.Println("\nWelcome to GetLit 🔥")
			fmt.Println("\nLet's set up your wallet so you can authenticate with the Lit network.")
			fmt.Println("\nWhat network would you like to encrypt with?")
			fmt.Println("(Please choose from one of the networks compatible with Lit Protocol.)")
			fmt.Println("https://developer.litprotocol.com/support/supportedChains")

			network, err := reader.ReadString('\n')
			if err != nil {
				panic(err)
			}

			fmt.Printf("\nEnter the private key for an EVM compatible wallet: ")

			pk, err := reader.ReadString('\n')
			if err != nil {
				panic(err)
			}

			// Creates and saves the config to the current directory
			c := config.New(strings.TrimSpace(network), strings.TrimSpace(pk))
			if err := c.Save(); err != nil {
				panic(err)
			}

			// Creates and saves the keyfile for the account
			if _, err := account.NewEthereum(c); err != nil {
				panic(err)
			}

			fmt.Println("\nGreat! You're all set up. Your settings have been saved to .getlit")
			fmt.Println("Call help to see how to encrypt and decrypt your data using the Lit Protocol.")
		case "encrypt":
			if !initialized() {
				fmt.Fprintf(os.Stderr, "Not initialized. Please run `init` first.")
				return
			}
			fmt.Printf("\x1b[1mCondition type set to: %s (Change with the -type option)\n\x1b[0m", *conditionType)
			// These will load the config and account from the saved values from init
			c := config.Load()
			a, err := account.New(c, *conditionType)

			data := make([]byte, 0)
			reader := bufio.NewReader(os.Stdin)
			// Any content can be entered with mulitple lines. Maybe not the best method to capture content but for this it's fine.
			fmt.Println("Enter the content you'd like to encrypt. Close the input by typing `q` on a new line and pressing enter.")
			for {
				line, err := reader.ReadBytes('\n')
				if err != nil {
					if err == io.EOF {
						break
					} else {
						panic(err)
					}
				}

				if strings.TrimSpace(string(line)) == "q" {
					break
				}

				data = append(data, line...)
			}

			symmetricKey := crypto.Prng(32)
			ciphertext := crypto.AesEncrypt(symmetricKey, data)

			// Create an instance of the Lit Client
			litClient, err := client.New(context.Background(), c.LitConfig)
			if err != nil {
				fmt.Fprintf(os.Stderr, err.Error())
				return
			}

			var encryptedKey string

			switch *conditionType {
			case "evmcontract":
				conds := ethereum.CreateEvmContractCondition(a.Chain)
				fmt.Println("\x1b[1mEvmContractCondition:\x1b[0m")
				jsonUtils.PrintJSON(conds[0])

				encryptedKey, err = client.SaveEncryptionKey(
					litClient,
					context.Background(),
					symmetricKey,
					a.AuthSig,
					conds,
					a.Chain,
					true,
				)
				if err != nil {
					fmt.Fprintf(os.Stderr, err.Error())
					os.Exit(1)
					return
				}
			case "accesscontrol":
				conds := ethereum.CreateAccessControlCondition()
				fmt.Println("Condition:")
				jsonUtils.PrintJSON(conds[0])

				encryptedKey, err = client.SaveEncryptionKey(
					litClient,
					context.Background(),
					symmetricKey,
					a.AuthSig,
					conds,
					a.Chain,
					true,
				)
				if err != nil {
					fmt.Fprintf(os.Stderr, err.Error())
					os.Exit(1)
					return
				}
			}

			fmt.Printf("\x1b[1m\nSymmetricKey:\x1b[0m %s\n", hex.EncodeToString(symmetricKey))
			fmt.Printf("\x1b[1mCipherText:\x1b[0m %s\n", hex.EncodeToString(ciphertext))
			fmt.Printf("\x1b[1mEncryptedKey:\x1b[0m %s\n", encryptedKey)
		case "decrypt":
			if !initialized() {
				fmt.Fprintf(os.Stderr, "Not initialized. Please run `init` first.")
				return
			}
			fmt.Printf("\x1b[1mCondition type set to: %s\n\n\x1b[0m", *conditionType)

			// These will load the config and account from the saved values from init
			c := config.Load()
			a, err := account.NewEthereum(c)
			if err != nil {
				panic(err)
			}

			fmt.Println("Enter the EvmContractCondition JSON object:")
			cond := conditions.EvmContractCondition{}
			if err := jsonUtils.CaptureJSON(&cond); err != nil {
				panic(err)
			}

			reader := bufio.NewReader(os.Stdin)

			fmt.Printf("Enter the encrypted key value: ")
			encryptedKey, err := reader.ReadString('\n')
			if err != nil {
				panic(err)
			}

			fmt.Printf("Enter the ciphertext value: ")
			ciphertextHex, err := reader.ReadString('\n')
			if err != nil {
				panic(err)
			}

			// Create an instance of the Lit Client
			litClient, err := client.New(context.Background(), c.LitConfig)
			if err != nil {
				fmt.Fprintf(os.Stderr, err.Error())
				return
			}

			// Send the request to the Lit network
			// A context can be used here to set the response timeout or to manually cancel the request..
			symmetricKey, err := client.GetEncryptionKey(
				litClient,
				context.Background(),
				a.AuthSig,
				[]conditions.EvmContractCondition{cond},
				a.Chain,
				strings.TrimSpace(encryptedKey),
			)
			if err != nil {
				fmt.Fprintf(os.Stderr, err.Error())
				fmt.Println("\nFailed to decrypt the message.")
				return
			}

			// Convert the ciphertext back to bytes
			ciphertext, err := hex.DecodeString(strings.TrimSpace(ciphertextHex))
			if err != nil {
				panic(err)
			}

			// lit-go-sdk provides the AES methods and includes padding.
			// The IV is prepended to the ciphertext. Not sure if this is how
			// the JS SDK does it.
			plaintext := crypto.AesDecrypt(symmetricKey, ciphertext)
			fmt.Printf("\nDecrypted message: %s\n", string(plaintext))
		}
	}
}
