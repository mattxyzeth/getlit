package ethereum

import (
	"bufio"
	"fmt"
	"getlit/jsonUtils"
	"os"
	"strings"

	"github.com/crcls/lit-go-sdk/conditions"
)

func CreateEvmContractCondition(chain string) []conditions.EvmContractCondition {
	reader := bufio.NewReader(os.Stdin)
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
	if err := jsonUtils.CaptureJSON(&abi); err != nil {
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

	condition := conditions.EvmContractCondition{
		ContractAddress: strings.TrimSpace(address),
		FunctionName:    strings.TrimSpace(method),
		FunctionParams:  strings.Split(strings.TrimSpace(args), ","),
		FunctionAbi:     abi,
		Chain:           chain,
		ReturnValueTest: conditions.ReturnValueTest{
			Key:        strings.TrimSpace(compKey),
			Comparator: strings.TrimSpace(comparator),
			Value:      strings.TrimSpace(compValue),
		},
	}

	return []conditions.EvmContractCondition{condition}
}