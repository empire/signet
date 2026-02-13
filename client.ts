import { BrowserProvider, toBeHex } from "ethers";

declare global {
  interface Window {
    ethereum?: unknown;
  }
}

const LOGIN_MESSAGE = "Login to the application";

function randomNonce(bytes = 16): string {
  const buffer = new Uint8Array(bytes);
  crypto.getRandomValues(buffer);
  return Array.from(buffer, (b) => b.toString(16).padStart(2, "0")).join("");
}

export async function signLogin(): Promise<{
  address: string;
  nonce: string;
  signature: string;
}> {
  if (!window.ethereum) {
    throw new Error("MetaMask is not available in this browser.");
  }

  const provider = new BrowserProvider(window.ethereum);
  const signer = await provider.getSigner();
  const address = await signer.getAddress();
  const network = await provider.getNetwork();
  const nonce = randomNonce();

  const domain = {
    name: "SignetAuth",
    version: "1",
    chainId: Number(network.chainId),
    verifyingContract: "0x0000000000000000000000000000000000000000",
  };

  const types = {
    LoginRequest: [
      { name: "message", type: "string" },
      { name: "nonce", type: "string" },
      { name: "walletAddress", type: "uint256" },
    ],
  };

  const value = {
    message: LOGIN_MESSAGE,
    nonce,
    walletAddress: BigInt(address),
  };

  await provider.send("eth_requestAccounts", []);

  const signature = await provider.send("eth_signTypedData_v4", [
    address,
    JSON.stringify({
      types: {
        EIP712Domain: [
          { name: "name", type: "string" },
          { name: "version", type: "string" },
          { name: "chainId", type: "uint256" },
          { name: "verifyingContract", type: "address" },
        ],
        ...types,
      },
      primaryType: "LoginRequest",
      domain,
      message: {
        ...value,
        walletAddress: toBeHex(value.walletAddress),
      },
    }),
  ]);

  console.log("EIP-712 signature:", signature); // 65-byte (130 hex chars + 0x)

  return { address, nonce, signature };
}
