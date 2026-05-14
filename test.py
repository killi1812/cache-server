import base64
from nacl.signing import VerifyKey
from nacl.exceptions import BadSignatureError

# 1. Your data
pub_key_b64 = "dtwETT5GRYLZkCfVtIYLjiC8AE1lsXx9aIQjlDRUvFc="
sig_b64 = "MFEM/TIDvUbHRpJPsDUs2cdvNcfjcn5kIIUWyo/7juuL3qOD+N/B6ijRoR+0varD0zCVlUQGm/+DhNKg+4yLDg=="

# Replace this with the exact fingerprint string constructed from the .narinfo
fingerprint_string = "1;/nix/store/ffgmyxfrc3v77azm9g8lix2kp3rcf443-testhello;1p4a6kwhz5h1ppcqc5k10mgjbbqj55pzwr98d68n048yrqs3bj5s;191640;/nix/store/ffgmyxfrc3v77azm9g8lix2kp3rcf443-testhello,/nix/store/j193mfi0f921y0kfs8vjc1znnr45ispv-glibc-2.40-66"
message = fingerprint_string.encode("utf-8")

# 2. Decode the base64 keys
try:
    verify_key = VerifyKey(base64.b64decode(pub_key_b64))
    signature = base64.b64decode(sig_b64)
except Exception as e:
    print(f"Error decoding Base64: {e}")
    exit(1)

# 3. Verify the math
try:
    verify_key.verify(message, signature)
    print(
        "✅ SUCCESS: The signature is valid! The private key used matches your public key."
    )
    print(
        "If Nix is still rejecting it, the issue is likely a mismatched key name (e.g., 'test.localhost-1') or a malformed fingerprint."
    )
except BadSignatureError:
    print("❌ FAILED: The signature is invalid.")
    print(
        "The cache server signed this package with a completely different private key than the one you provided to Nix."
    )
