# GetLit 🔥

This application's purpose is to showcase the use of
https://github.com/crcls/lit-go-sdk that I'm working on.

## Initialization Instructions

1. Clone the repo and run `make build`
2. Initialize the CLI with `./dist/getlit init`
3. Choose the Blockchain and network you'd like to use to grant access to
   your encrypted content.
4. Add an EVM compatible wallet private key for
   authentication.

## Encrypt and store content
1. Call `./dist/getlit encrypt` and pass in the data you'd like to
   encrypt. You can type your message or pipe in a file's content.
2. So far, this application only supports EVM contract conditions for
   granting access to your encrypted content. Enter a contract address
   that is deployed to the chain and network the CLI was initialized
   with.
3. Choose the method on the contract you'd like to call to run the
   verification.
4. Enter the parameters needed for the method as a comma separated list.
5. Choose the comparison type for testing the return values of the
   verification method.
6. The CLI will encrypt your content, set the conditions in the Lit
   network, and return an encrypted key and an IPFS CID where your
   content was stored.

## Retrieve and decrypt your content
1. Call `./dist/getlit decrypt`
2. Enter your encrypted key and your IPFS CID you received from the
   encryption process.
3. If the wallet you authenticated with satisfies the EVM contract
   condition you entered in the encryption process, your content will
   print to STDOUT.

## Test values

Private Key: 

Blockchain: polygon

Network: mumbai

Contract Address: 

Method Name: verify

Params: :userAddress

Comparator: "=="

Comparison Key: ""

Comparison Value: "true"

