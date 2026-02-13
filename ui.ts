import { BrowserProvider } from "ethers";
import { buildLoginTypedData } from "./client";

type ChallengeResponse = {
  nonce: string;
  userExists: boolean;
};

const connectBtn = document.getElementById("connect-btn") as HTMLButtonElement;
const signInBtn = document.getElementById("signin-btn") as HTMLButtonElement;
const registerBtn = document.getElementById("register-btn") as HTMLButtonElement;
const walletLabel = document.getElementById("wallet") as HTMLSpanElement;
const registerPanel = document.getElementById("register-panel") as HTMLDivElement;
const usernameInput = document.getElementById("username-input") as HTMLInputElement;
const statusBox = document.getElementById("status") as HTMLDivElement;

let currentAddress = "";
let currentNonce = "";

function setStatus(message: string): void {
  statusBox.textContent = message;
}

async function getProvider(): Promise<BrowserProvider> {
  if (!window.ethereum) {
    throw new Error("MetaMask not found.");
  }
  return new BrowserProvider(window.ethereum);
}

async function connectWallet(): Promise<void> {
  const provider = await getProvider();
  await provider.send("eth_requestAccounts", []);
  const signer = await provider.getSigner();
  currentAddress = await signer.getAddress();
  walletLabel.textContent = currentAddress;
  setStatus("Wallet connected.");
}

async function getChallenge(address: string): Promise<ChallengeResponse> {
  const res = await fetch(`/api/challenge?address=${encodeURIComponent(address)}`);
  if (!res.ok) {
    throw new Error(await res.text());
  }
  return res.json();
}

async function signIn(): Promise<void> {
  if (!currentAddress) {
    throw new Error("Connect wallet first.");
  }

  const provider = await getProvider();
  const network = await provider.getNetwork();
  const challenge = await getChallenge(currentAddress);
  currentNonce = challenge.nonce;

  const typedData = buildLoginTypedData(currentAddress, Number(network.chainId), challenge.nonce);
  const signature = await provider.send("eth_signTypedData_v4", [currentAddress, JSON.stringify(typedData)]);

  const res = await fetch("/api/signin", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      address: currentAddress,
      nonce: challenge.nonce,
      signature,
      walletAddressField: currentAddress,
    }),
  });

  const body = await res.json();
  if (body.code === "user_not_exists") {
    registerPanel.style.display = "block";
    setStatus("User not exists. Please register username.");
    return;
  }

  registerPanel.style.display = "none";
  setStatus(`Signed in. Welcome ${body.username}.`);
}

async function encryptUsername(username: string): Promise<string> {
  const res = await fetch("/api/public-key");
  if (!res.ok) throw new Error("Failed to load server public key");
  const { publicKeyPem } = await res.json();

  const cleanPem = publicKeyPem.replace("-----BEGIN PUBLIC KEY-----", "").replace("-----END PUBLIC KEY-----", "").replace(/\s+/g, "");
  const der = Uint8Array.from(atob(cleanPem), (c) => c.charCodeAt(0));

  const key = await crypto.subtle.importKey(
    "spki",
    der.buffer,
    { name: "RSA-OAEP", hash: "SHA-256" },
    false,
    ["encrypt"],
  );

  const ciphertext = await crypto.subtle.encrypt(
    { name: "RSA-OAEP" },
    key,
    new TextEncoder().encode(username),
  );

  const bytes = new Uint8Array(ciphertext);
  return btoa(String.fromCharCode(...bytes));
}

async function registerUser(): Promise<void> {
  if (!currentAddress || !currentNonce) {
    throw new Error("Sign in first to get nonce.");
  }
  const username = usernameInput.value.trim();
  if (!username) throw new Error("Username is required");

  const provider = await getProvider();
  const network = await provider.getNetwork();
  const typedData = buildLoginTypedData(currentAddress, Number(network.chainId), currentNonce);
  const signature = await provider.send("eth_signTypedData_v4", [currentAddress, JSON.stringify(typedData)]);

  const encryptedUsername = await encryptUsername(username);

  const res = await fetch("/api/register", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      address: currentAddress,
      nonce: currentNonce,
      signature,
      walletAddressField: currentAddress,
      encryptedUsername,
    }),
  });

  const body = await res.json();
  if (!res.ok) {
    throw new Error(body.error || "Registration failed");
  }

  registerPanel.style.display = "none";
  setStatus(`Registered and signed in as ${body.username}.`);
}

connectBtn.addEventListener("click", () => connectWallet().catch((e: Error) => setStatus(e.message)));
signInBtn.addEventListener("click", () => signIn().catch((e: Error) => setStatus(e.message)));
registerBtn.addEventListener("click", () => registerUser().catch((e: Error) => setStatus(e.message)));
