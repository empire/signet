import { describe, expect, it } from "vitest";
import { buildLoginTypedData, LOGIN_MESSAGE, VERIFYING_CONTRACT } from "../client";

describe("buildLoginTypedData", () => {
  it("builds EIP-712 LoginRequest payload", () => {
    const address = "0x1111111111111111111111111111111111111111";
    const nonce = "nonce-123";

    const typedData = buildLoginTypedData(address, 1, nonce);

    expect(typedData.primaryType).toBe("LoginRequest");
    expect(typedData.message.message).toBe(LOGIN_MESSAGE);
    expect(typedData.message.nonce).toBe(nonce);
    expect(typedData.domain.verifyingContract).toBe(VERIFYING_CONTRACT);
    expect(typedData.message.walletAddress).toBe(address);
    expect(typedData.types.LoginRequest).toEqual([
      { name: "message", type: "string" },
      { name: "nonce", type: "string" },
      { name: "walletAddress", type: "uint256" },
    ]);
  });
});
