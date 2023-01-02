package main

import (
	"bufio"
	"fmt"
	"getlit/account"
	"getlit/config"
	"os"
	"path/filepath"
	"strings"
)

var CMDS = map[string]string{
	"init":    "Initialize the CLI.",
	"encrypt": "Encrypt data and store on IPFS.",
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

			c := config.New(strings.TrimSpace(network), strings.TrimSpace(pk))
			if err := c.Save(); err != nil {
				panic(err)
			}

			if _, err := account.New(c); err != nil {
				panic(err)
			}

			fmt.Println("\nGreat! You're all set up. Your settings have been saved to .getlit")
			fmt.Println("Call help to see how to encrypt and decrypt your data using the Lit Protocol.")
		}
	}
}
