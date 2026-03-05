import { createHmac, timingSafeEqual } from "node:crypto";

/**
 * Verify a Vendel webhook `X-Webhook-Signature` header.
 *
 * The signature is an HMAC-SHA256 hex digest computed over the JSON
 * payload string using the webhook secret as the key.
 *
 * @param payload - The raw request body string, or a parsed object
 *                  (will be serialized with sorted keys, no spaces).
 * @param signature - Value of the `X-Webhook-Signature` header.
 * @param secret - The webhook secret configured in the Vendel dashboard.
 * @returns `true` if the signature is valid.
 */
export function verifyWebhookSignature(
  payload: string | Record<string, unknown>,
  signature: string,
  secret: string,
): boolean {
  const body =
    typeof payload === "string"
      ? payload
      : JSON.stringify(payload, Object.keys(payload).sort());

  const expected = createHmac("sha256", secret).update(body).digest("hex");

  if (expected.length !== signature.length) return false;
  return timingSafeEqual(Buffer.from(expected), Buffer.from(signature));
}
