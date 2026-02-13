import { BrowserProvider, toBeHex } from "ethers";

export const LOGIN_MESSAGE = "Login to the application";
export const VERIFYING_CONTRACT = "0x0000000000000000000000000000000000000000";

export type LoginRequestValue = {
  message: string;
  nonce: string;
  walletAddress: bigint;
};

export type LoginTypedData = {
  types: Record<string, Array<{ name: string; type: string }>>;
  primaryType: "LoginRequest";
  domain: {
    name: string;
    version: string;
    chainId: number;
    verifyingContract: string;
  };
  message: {
    message: string;
    nonce: string;
    walletAddress: string;
  };
};

declare global {
  interface Window {
    ethereum?: unknown;
  }
}

function randomNonce(bytes = 16): string {
  const buffer = new Uint8Array(bytes);
  crypto.getRandomValues(buffer);
  return Array.from(buffer, (b) => b.toString(16).padStart(2, "0")).join("");
}

export function buildLoginTypedData(address: string, chainId: number, nonce: string): LoginTypedData {
  const value: LoginRequestValue = {
    message: LOGIN_MESSAGE,
    nonce,
    walletAddress: BigInt(address),
  };

  return {
    types: {
      EIP712Domain: [
        { name: "name", type: "string" },
        { name: "version", type: "string" },
        { name: "chainId", type: "uint256" },
        { name: "verifyingContract", type: "address" },
      ],
      LoginRequest: [
        { name: "message", type: "string" },
        { name: "nonce", type: "string" },
        { name: "walletAddress", type: "uint256" },
      ],
    },
    primaryType: "LoginRequest",
    domain: {
      name: "SignetAuth",
      version: "1",
      chainId,
      verifyingContract: VERIFYING_CONTRACT,
    },
    message: {
      message: value.message,
      nonce: value.nonce,
      walletAddress: toBeHex(value.walletAddress),
    },
  };
}

export async function signLogin(): Promise<{
  address: string;
  nonce: string;
  signature: string;
}> {
  if (typeof window === "undefined" || !window.ethereum) {
    throw new Error("MetaMask is not available in this browser.");
  }

  const provider = new BrowserProvider(window.ethereum);
  await provider.send("eth_requestAccounts", []);

  const signer = await provider.getSigner();
  const address = await signer.getAddress();
  const network = await provider.getNetwork();
  const nonce = randomNonce();

  const typedData = buildLoginTypedData(address, Number(network.chainId), nonce);

  const signature = await provider.send("eth_signTypedData_v4", [
    address,
    JSON.stringify(typedData),
  ]);

  console.log("EIP-712 signature:", signature); // 65-byte (130 hex chars + 0x)

  return { address, nonce, signature };
}
