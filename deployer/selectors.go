package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"golang.org/x/crypto/sha3"
)

func main() {
	const solidityErrors = `
ContractPaused()
ContractNotPaused()
ContractInitialized()
ContractNotInitialized()
NotContractOwner()
ZeroAddressDetected()
ReentrancyDetected()
InputExceedsUint64(string,uint256)
NotCalledViaProxy()
ProxyInitialized()
SignatureInvalidLength(uint256)
SignatureInvalidS(uint256)
SignatureInvalidV(uint8)
InvalidTree(uint256,uint256,uint256)
InvalidSubtree(uint256,uint256,uint256)
InvalidLevel(uint256,uint256,uint256)
ZeroInputDetected()
MerkleTreeFull(uint256,uint256)
MerkleProofsInvalidLength(uint256,uint256)
IndexOutOfBounds(uint256,uint256,uint256)
InvalidRoleDetected()
UnauthorizedAccess(string)
InvalidPublicSignals(string)
InvalidSignature(string)
InvalidProofs(string)
UnknownRootDetected(string)
InvalidElectionId(uint32)
InvalidElectionStatus(uint8,uint8)
InvalidElectionPeriod()
BlacklistedUserDetected(uint256)
InvalidNonce(uint256,uint256,uint256)
EmptyArray(string)
AlreadyRegistered(string)
`

	// Updated regex: match ErrorName(...) without the "error " prefix or semicolon
	re := regexp.MustCompile(`(\w+)\(([^)]*)\)`)

	matches := re.FindAllStringSubmatch(solidityErrors, -1)
	if matches == nil {
		log.Fatal("No error definitions found.")
	}

	for _, match := range matches {
		errorName := match[1]
		params := strings.TrimSpace(match[2])

		var signature string
		if params == "" {
			signature = fmt.Sprintf("%s()", errorName)
		} else {
			paramList := []string{}
			for _, p := range strings.Split(params, ",") {
				p = strings.TrimSpace(p)
				if p == "" {
					continue
				}
				parts := strings.Fields(p)
				paramType := parts[0]
				paramList = append(paramList, paramType)
			}
			signature = fmt.Sprintf("%s(%s)", errorName, strings.Join(paramList, ","))
		}

		hasher := sha3.NewLegacyKeccak256()
		hasher.Write([]byte(signature))
		hash := hasher.Sum(nil)
		selector := hash[:4]

		fmt.Printf("%-30s -> 0x%x\n", signature, selector)
	}

}
