package auth

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

const loginMessage = "Login to the application"

func LOGIN_MESSAGE() string { return loginMessage }

func buildTypedData(message, nonce string, walletAddr common.Address, chainID *big.Int) apitypes.TypedData {
	return apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": []apitypes.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"LoginRequest": []apitypes.Type{
				{Name: "message", Type: "string"},
				{Name: "nonce", Type: "string"},
				{Name: "walletAddress", Type: "uint256"},
			},
		},
		PrimaryType: "LoginRequest",
		Domain: apitypes.TypedDataDomain{
			Name:              "SignetAuth",
			Version:           "1",
			ChainId:           mathBigToHexOrDecimal256(chainID),
			VerifyingContract: "0x0000000000000000000000000000000000000000",
		},
		Message: apitypes.TypedDataMessage{
			"message":       message,
			"nonce":         nonce,
			"walletAddress": mathBigToHexOrDecimal256(addressToBigInt(walletAddr)),
		},
	}
}

// VerifySignature verifies an EIP-712 signature from `eth_signTypedData_v4`.
// expectedAddress is the wallet to authenticate, walletAddressField is the same value used in LoginRequest.walletAddress.
func VerifySignature(expectedAddress, message, nonce, walletAddressField, signatureHex string, chainID *big.Int) (bool, error) {
	if message != loginMessage {
		return false, fmt.Errorf("unexpected login message")
	}

	expected := common.HexToAddress(expectedAddress)
	walletAddr := common.HexToAddress(walletAddressField)

	typedData := buildTypedData(message, nonce, walletAddr, chainID)
	digest, _, err := apitypes.TypedDataAndHash(typedData)
	if err != nil {
		return false, fmt.Errorf("build typed data digest: %w", err)
	}

	sig, err := hexutil.Decode(signatureHex)
	if err != nil {
		return false, fmt.Errorf("decode signature: %w", err)
	}
	if len(sig) != 65 {
		return false, fmt.Errorf("invalid signature length: got %d, want 65", len(sig))
	}

	// MetaMask may return v as 27/28 while go-ethereum expects 0/1.
	if sig[64] >= 27 {
		sig[64] -= 27
	}
	if sig[64] != 0 && sig[64] != 1 {
		return false, fmt.Errorf("invalid recovery id (v): %d", sig[64])
	}

	pubKey, err := crypto.SigToPub(digest, sig)
	if err != nil {
		return false, fmt.Errorf("recover public key: %w", err)
	}
	recovered := crypto.PubkeyToAddress(*pubKey)

	return strings.EqualFold(recovered.Hex(), expected.Hex()), nil
}

func addressToBigInt(addr common.Address) *big.Int {
	return new(big.Int).SetBytes(addr.Bytes())
}

func mathBigToHexOrDecimal256(n *big.Int) *apitypes.HexOrDecimal256 {
	v := apitypes.NewHexOrDecimal256(0)
	if n == nil {
		return v
	}
	v = (*apitypes.HexOrDecimal256)(new(big.Int).Set(n))
	return v
}
