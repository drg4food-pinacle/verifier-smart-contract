package sign

import (
	"bytes"
	"deployer/internal/types"
	"errors"
	"fmt"
	"text/template"

	"github.com/ethereum/go-ethereum/crypto"
)

type Role types.Role

// ! Must be in sync with the Solidity contract
// Message templates for different roles.
var messageTemplates = map[Role]string{
	0: `I am an Election Leader with hashed address {{.HashedAddress}} and nonce {{.Nonce}}`,
	1: `I am a Rollup Leader with hashed address {{.HashedAddress}} and I commit block {{.BlockNumber}} with hash {{.BlockHash}} and nonce {{.Nonce}}`, //! Need changes also in Solitidy
	2: `I am a Voter with hashed address {{.HashedAddress}} and I vote {{.VoteOption}} for election {{.ElectionID}} and nonce {{.Nonce}}`,
}

// GenerateMessage generates a formatted message string for the given role and parameters.
// It is a wrapper around generateMessage, exposing it for external use.
func GenerateMessage(role Role, params map[string]interface{}) (string, error) {
	return generateMessage(role, params)
}

// GenerateHashedMessage generates a message for the given role and parameters,
// then hashes it using the hashPersonalMessage function. It returns the hashed
// message as a byte slice, or an error if message generation fails.
func GenerateHashedMessage(role Role, params map[string]interface{}) ([]byte, error) {
	msg, err := generateMessage(role, params)
	if err != nil {
		return nil, err
	}

	return hashPersonalMessage(msg), nil
}

// generateMessage generates a formatted message string based on the provided role and parameters.
// It retrieves the message template associated with the given role, parses it, and executes the template
// using the provided parameters. The resulting message is returned as a string.
// Returns an error if the role is invalid, the template cannot be parsed, or the template execution fails.
//
// Parameters:
//   - role:   The Role for which the message template should be used.
//   - params: A map of parameters to be injected into the template.
//
// Returns:
//   - string: The generated message.
//   - error:  An error if message generation fails.
func generateMessage(role Role, params map[string]interface{}) (string, error) {
	tmplStr, ok := messageTemplates[role]
	if !ok {
		return "", errors.New("invalid role specified for message generation")
	}

	tmpl, err := template.New(string(role)).Parse(tmplStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, params); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// HashPersonalMessage returns the keccak256 hash of the Ethereum signed message prefix + msg
// hashPersonalMessage prefixes the given message with the standard Ethereum message prefix,
// then computes and returns the Keccak256 hash of the resulting byte slice.
// This is used to produce a hash compatible with Ethereum's personal_sign and eth_sign RPC methods.
//
// Parameters:
//   - msg: The message to be hashed.
//
// Returns:
//   - The Keccak256 hash of the prefixed message as a byte slice.
func hashPersonalMessage(msg string) []byte {
	msgPrefix := fmt.Sprintf("\x19Ethereum Signed Message:\n%d", len(msg))
	prefixedMsg := msgPrefix + msg
	return crypto.Keccak256([]byte(prefixedMsg))
}
