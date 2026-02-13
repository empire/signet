package auth

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

func TestVerifySignature_SuccessAndVNormalization(t *testing.T) {
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	addr := crypto.PubkeyToAddress(key.PublicKey)
	chainID := big.NewInt(1)
	nonce := "abc123nonce"

	typedData := buildTypedData(loginMessage, nonce, addr, chainID)
	digest, _, err := apitypes.TypedDataAndHash(typedData)
	if err != nil {
		t.Fatalf("typed data hash: %v", err)
	}

	sig, err := crypto.Sign(digest, key)
	if err != nil {
		t.Fatalf("sign digest: %v", err)
	}

	t.Run("v as 0/1", func(t *testing.T) {
		ok, verifyErr := VerifySignature(addr.Hex(), loginMessage, nonce, addr.Hex(), hexutil.Encode(sig), chainID)
		if verifyErr != nil {
			t.Fatalf("verify error: %v", verifyErr)
		}
		if !ok {
			t.Fatalf("expected signature verification to pass")
		}
	})

	t.Run("v as 27/28", func(t *testing.T) {
		sig27 := make([]byte, len(sig))
		copy(sig27, sig)
		sig27[64] += 27

		ok, verifyErr := VerifySignature(addr.Hex(), loginMessage, nonce, addr.Hex(), hexutil.Encode(sig27), chainID)
		if verifyErr != nil {
			t.Fatalf("verify error: %v", verifyErr)
		}
		if !ok {
			t.Fatalf("expected signature verification to pass for 27/28 v")
		}
	})
}

func TestVerifySignature_MismatchAddress(t *testing.T) {
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	addr := crypto.PubkeyToAddress(key.PublicKey)
	chainID := big.NewInt(1)
	nonce := "mismatch"

	typedData := buildTypedData(loginMessage, nonce, addr, chainID)
	digest, _, err := apitypes.TypedDataAndHash(typedData)
	if err != nil {
		t.Fatalf("typed data hash: %v", err)
	}
	sig, err := crypto.Sign(digest, key)
	if err != nil {
		t.Fatalf("sign digest: %v", err)
	}

	other := common.HexToAddress("0x1000000000000000000000000000000000000001")
	ok, verifyErr := VerifySignature(other.Hex(), loginMessage, nonce, addr.Hex(), hexutil.Encode(sig), chainID)
	if verifyErr != nil {
		t.Fatalf("verify error: %v", verifyErr)
	}
	if ok {
		t.Fatalf("expected false for mismatched expected address")
	}
}
