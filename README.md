# GetLit 🔥

This application's purpose is to showcase the use of
https://github.com/crcls/lit-go-sdk that I'm working on.

## Initialization Instructions

1. Make sure you have Go 1.18 or > installed with `go version`. If not checkout https://go.dev/doc/install
2. Clone the repo and run `make build`
3. Initialize the CLI with `./dist/getlit init`
4. Choose the Blockchain and network you'd like to use to grant access to
   your encrypted content.
5. Add an EVM compatible wallet private key for
   authentication.

***PRO TIP: turn on debug output with `LIT_DEBUG=true`***

## Encrypt and store content
1. Call `./dist/getlit encrypt` and type in the message you'd like to
   encrypt. End the input by typing `q` on a new line and enter.
2. So far, this application only supports EVM contract conditions for
   granting access to your encrypted content. Enter a contract address
   that is deployed to the chain and network the CLI was initialized
   with.
3. Type in the name of the method on the contract you'd like Lit to use
   for the access verification.
4. Enter the parameters needed for the method as a comma separated list.
5. Enter a JSON object for the ABI of the method.
6. Enter a key of the return value to compare against. Leave it blank
   for an unnamed return value.
7. Choose the comparison type for testing the return values of the
   verification method.
8. Enter the expected return value to compare against.
6. The CLI will encrypt your content, save the conditions with the Lit
   network, and return the AuthSig, EvmContractCondition object, the
   generated symmetric key, the encrypted content as the Ciphertext, and
   the encrypted key value.

## Retrieve and decrypt your content
1. Call `./dist/getlit decrypt`
2. Enter the EvmContractCondition JSON object.
3. Enter the Encrypted Key value.
4. Enter the Ciphertext value.
5. If the wallet you authenticated with satisfies the EVM contract
   condition you entered in the encryption process, your content will
   print to STDOUT.

## Test values

Network: mumbai

Private Key: 14526519c506b5f523c3e935ab8dba5d53ee6f93c6258af8023eea4aa7607ae5

Contract Address: 0x465fe903849d4d42ae674017BB5C7e20C9eB71a8

Method Name: verify

Params: :userAddress

Function ABI:
```json
{
  "inputs": [
    {
      "internalType": "address",
      "name": "node",
      "type": "address"
    }
  ],
  "name": "verify",
  "outputs": [
    {
      "internalType": "bool",
      "name": "",
      "type": "bool"
    }
  ],
  "stateMutability": "view",
  "type": "function"
}
```

Comparator: "="

Comparison Key: ""

Comparison Value: "true"

## Issues and Comments
Please open an issue in the lit-go-sdk repo.
https://github.com/crcls/lit-go-sdk/issues
