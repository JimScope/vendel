class EnderError(Exception):
    """Base exception for the Ender SDK."""


class EnderAPIError(EnderError):
    """Raised when the Ender API returns an error response."""

    def __init__(self, status_code: int, message: str, detail: dict | None = None):
        self.status_code = status_code
        self.message = message
        self.detail = detail or {}
        super().__init__(f"[{status_code}] {message}")


class EnderQuotaError(EnderAPIError):
    """Raised when a quota limit is exceeded (HTTP 429)."""

    def __init__(self, message: str, detail: dict):
        super().__init__(429, message, detail)
        self.limit: int = detail.get("limit", 0)
        self.used: int = detail.get("used", 0)
        self.available: int = detail.get("available", 0)
