import { VendelAPIError, VendelQuotaError } from "./errors.js";
import type { VendelClientOptions, Quota, SendSMSResponse } from "./types.js";

/**
 * Client for the Vendel SMS gateway API.
 *
 * Uses an integration API key (`vk_` prefix) for authentication.
 */
export class VendelClient {
  private readonly baseUrl: string;
  private readonly apiKey: string;
  private readonly timeout: number;

  constructor(options: VendelClientOptions) {
    this.baseUrl = options.baseUrl.replace(/\/+$/, "");
    this.apiKey = options.apiKey;
    this.timeout = options.timeout ?? 30_000;
  }

  /**
   * Send an SMS to one or more recipients.
   *
   * @param recipients - Phone numbers in E.164 format (e.g. `+1234567890`).
   * @param body - Message text (max 1600 characters).
   * @param deviceId - Optional device to route through.
   */
  async sendSms(
    recipients: string[],
    body: string,
    deviceId?: string,
  ): Promise<SendSMSResponse> {
    const payload: Record<string, unknown> = { recipients, body };
    if (deviceId) payload.device_id = deviceId;
    return this.post<SendSMSResponse>("/api/sms/send", payload);
  }

  /** Get the current quota for the authenticated user. */
  async getQuota(): Promise<Quota> {
    return this.get<Quota>("/api/plans/quota");
  }

  // ------------------------------------------------------------------
  // Internal helpers
  // ------------------------------------------------------------------

  private async get<T>(path: string): Promise<T> {
    const resp = await fetch(`${this.baseUrl}${path}`, {
      headers: { "X-API-Key": this.apiKey },
      signal: AbortSignal.timeout(this.timeout),
    });
    return this.handleResponse<T>(resp);
  }

  private async post<T>(path: string, body: unknown): Promise<T> {
    const resp = await fetch(`${this.baseUrl}${path}`, {
      method: "POST",
      headers: {
        "X-API-Key": this.apiKey,
        "Content-Type": "application/json",
      },
      body: JSON.stringify(body),
      signal: AbortSignal.timeout(this.timeout),
    });
    return this.handleResponse<T>(resp);
  }

  private async handleResponse<T>(resp: Response): Promise<T> {
    if (resp.status === 429) {
      const data = (await resp.json()) as Record<string, unknown>;
      throw new VendelQuotaError(
        (data.detail as string) ?? "Quota exceeded",
        data,
      );
    }
    if (!resp.ok) {
      let data: Record<string, unknown> = {};
      try {
        data = (await resp.json()) as Record<string, unknown>;
      } catch {
        /* empty */
      }
      throw new VendelAPIError(
        resp.status,
        (data.message as string) ?? resp.statusText,
        data,
      );
    }
    return (await resp.json()) as T;
  }
}
