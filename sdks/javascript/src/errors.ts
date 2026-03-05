export class VendelError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "VendelError";
  }
}

export class VendelAPIError extends VendelError {
  public readonly statusCode: number;
  public readonly detail: Record<string, unknown>;

  constructor(
    statusCode: number,
    message: string,
    detail: Record<string, unknown> = {},
  ) {
    super(`[${statusCode}] ${message}`);
    this.name = "VendelAPIError";
    this.statusCode = statusCode;
    this.detail = detail;
  }
}

export class VendelQuotaError extends VendelAPIError {
  public readonly limit: number;
  public readonly used: number;
  public readonly available: number;

  constructor(message: string, detail: Record<string, unknown>) {
    super(429, message, detail);
    this.name = "VendelQuotaError";
    this.limit = (detail.limit as number) ?? 0;
    this.used = (detail.used as number) ?? 0;
    this.available = (detail.available as number) ?? 0;
  }
}
