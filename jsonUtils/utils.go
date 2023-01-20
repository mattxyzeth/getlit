package jsonUtils

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/crcls/lit-go-sdk/auth"
	"github.com/crcls/lit-go-sdk/conditions"
)

type marshalable interface {
	auth.AuthSig | conditions.EvmContractCondition | conditions.AccessControlCondition | conditions.AbiMember
}

func PrintJSON[T marshalable](data T) {
	b, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	var out bytes.Buffer
	json.Indent(&out, b, "", "\t")
	out.WriteTo(os.Stdout)
	fmt.Println("")
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