from __future__ import annotations

import hashlib
import hmac
import json


def verify_webhook_signature(
    payload: str | bytes | dict,
    signature: str,
    secret: str,
) -> bool:
    """Verify a Vendel webhook ``X-Webhook-Signature`` header.

    The signature is an HMAC-SHA256 hex digest computed over the JSON
    payload string using the webhook secret as the key.

    Args:
        payload: The raw request body (string or bytes), or a parsed dict
                 (will be serialized with sorted keys, no spaces).
        signature: Value of the ``X-Webhook-Signature`` header.
        secret: The webhook secret configured in the Vendel dashboard.

    Returns:
        ``True`` if the signature is valid.
    """
    if isinstance(payload, dict):
        payload = json.dumps(payload, separators=(",", ":"), sort_keys=True)
    if isinstance(payload, str):
        payload = payload.encode()

    expected = hmac.new(secret.encode(), payload, hashlib.sha256).hexdigest()
    return hmac.compare_digest(expected, signature)
