package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"getlit/account"
	"getlit/config"
	"io"
	"os"
	"path/filepath"
	"strings"
	"encoding/hex"
	"context"

	"github.com/crcls/lit-go-sdk/auth"
	"github.com/crcls/lit-go-sdk/client"
	"github.com/crcls/lit-go-sdk/conditions"
	"github.com/crcls/lit-go-sdk/crypto"
)

var CMDS = map[string]string{
	"init":    "Initialize the CLI.",
	"encrypt": "Encrypt data and store on IPFS. Pipe from a file or type your message and end the input with 'q'.",
	"decrypt": "Retrieve and decrypt data.",
	"help":    "Prints the help context.",
}

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
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "help":
			fmt.Println("")
			fmt.Println("Usage: bui command [OPTIONS]")
			fmt.Println("")

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
			if _, err := account.New(c); err != nil {
				panic(err)
			}

			fmt.Println("\nGreat! You're all set up. Your settings have been saved to .getlit")
			fmt.Println("Call help to see how to encrypt and decrypt your data using the Lit Protocol.")
		case "encrypt":
			if !initialized() {
				fmt.Fprintf(os.Stderr, "Not initialized. Please run `init` first.")
				return
			}

			// These will load the config and account from the saved values from init
			c := config.Load()
			account, err := account.New(c)
			if err != nil {
				panic(err)
			}

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

			// Collect the condition details
			fmt.Printf("Enter an address for a contract that will perform the verification: ")
			address, err := reader.ReadString('\n')
			if err != nil {
				panic(err)
			}

			fmt.Printf("Enter the name of the method to call on this contract: ")
			method, err := reader.ReadString('\n')
			if err != nil {
				panic(err)
			}

			fmt.Printf("Enter the method arguments to call the verification method (CSV): ")
			args, err := reader.ReadString('\n')
			if err != nil {
				panic(err)
			}

			fmt.Printf("Enter the function ABI json: ")
			abi := conditions.AbiMember{}
			if err := CaptureJSON(&abi); err != nil {
				panic(err)
			}

			fmt.Printf("What is the name of the key from the return values that should be used for comparison (return for unnamed key): ")
			compKey, err := reader.ReadString('\n')
			if err != nil {
				panic(err)
			}

			fmt.Printf("How should the return value be compared(=, >, >=, <=, <): ")
			comparator, err := reader.ReadString('\n')
			if err != nil {
				panic(err)
			}

			fmt.Printf("What value should be used to compare the return value against: ")
			compValue, err := reader.ReadString('\n')
			if err != nil {
				panic(err)
			}

			// Generate the AuthSig. Currently, this is not part of the SDK.
			// I figured that the consuming application would want to implement
			// it in a specific way. If you disagree, please open an issue in:
			// https://github.com/crcls/lit-go-sdk/issues
			authSig, err := account.Siwe(c.ChainId, "")
			if err != nil {
				panic(err)
			}
			fmt.Println("AuthSig:")
			PrintJSON(authSig)

			condition := conditions.EvmContractCondition{
				ContractAddress: strings.TrimSpace(address),
				FunctionName:    strings.TrimSpace(method),
				FunctionParams:  strings.Split(strings.TrimSpace(args), ","),
				FunctionAbi:     abi,
				Chain:           c.Network,
				ReturnValueTest: conditions.ReturnValueTest{
					Key:        strings.TrimSpace(compKey),
					Comparator: strings.TrimSpace(comparator),
					Value:      strings.TrimSpace(compValue),
				},
			}

			fmt.Println("EvmContractCondition:")
			PrintJSON(condition)

			symmetricKey := crypto.Prng(32)
			ciphertext := crypto.AesEncrypt(symmetricKey, data)

			fmt.Printf("SymmetricKey: %s\n", hex.EncodeToString(symmetricKey))
			fmt.Printf("CipherText: %s\n", hex.EncodeToString(ciphertext))

			// Create an instance of the Lit Client
			litClient, err := client.New(context.Background(), c.LitConfig)
			if err != nil {
				fmt.Fprintf(os.Stderr, err.Error())
				return
			}

			encryptedKey, err := litClient.SaveEncryptionKey(
				context.Background(),
				symmetricKey,
				authSig,
				[]conditions.EvmContractCondition{condition},
				c.ChainId,
				true,
			)
			if err != nil {
				fmt.Fprintf(os.Stderr, err.Error())
			}

			fmt.Printf("\nEncryptedKey: %s\n", encryptedKey)
		case "decrypt":
			if !initialized() {
				fmt.Fprintf(os.Stderr, "Not initialized. Please run `init` first.")
				return
			}

			// These will load the config and account from the saved values from init
			c := config.Load()
			account, err := account.New(c)
			if err != nil {
				panic(err)
			}

			fmt.Println("Enter the EvmContractCondition JSON object:")
			cond := conditions.EvmContractCondition{}
			if err := CaptureJSON(&cond); err != nil {
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

			authSig, err := account.Siwe(c.ChainId, "")
			if err != nil {
				panic(err)
			}

			// Create an instance of the Lit Client
			litClient, err := client.New(context.Background(), c.LitConfig)
			if err != nil {
				fmt.Fprintf(os.Stderr, err.Error())
				return
			}

			// Build the request params to decrypt the symmetric key.
			keyParams := client.EncryptedKeyParams{
				AuthSig:               authSig,
				Chain:                 c.Network,
				EvmContractConditions: []conditions.EvmContractCondition{cond},
				ToDecrypt:             strings.TrimSpace(encryptedKey),
			}

			// Send the request to the Lit network
			// A context can be used here to set the response timeout or to manually cancel the request..
			symmetricKey, err := litClient.GetEncryptionKey(context.Background(), &keyParams)
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

type marshalable interface {
	auth.AuthSig | conditions.EvmContractCondition | conditions.AbiMember
}

func PrintJSON[T marshalable](data T) {
	b, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	var out bytes.Buffer
	json.Indent(&out, b, "", "\t")
	out.WriteTo(os.Stdout)
	fmt.Println("\n")
}

func CaptureJSON[T marshalable](s *T) error {
	data := make([]byte, 0)
	brackets := make([]byte, 0)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(bufio.ScanBytes)
	for scanner.Scan() {
		b := scanner.Bytes()

		if err := scanner.Err(); err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
		data = append(data, b...)

		str := string(b)

		if str == "{" || str == "[" {
			brackets = append(brackets, b[0])
		}

		if str == "}" || str == "]" {
			brackets = brackets[:len(brackets)-1]
		}

		if len(brackets) == 0 {
			break
		}
	}

	if err := json.Unmarshal(data, s); err != nil {
		return err
	}

	return nil
}
