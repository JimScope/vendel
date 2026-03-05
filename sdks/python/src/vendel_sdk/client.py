from __future__ import annotations

import requests

from .exceptions import VendelAPIError, VendelQuotaError
from .types import Quota, SendSMSResponse


class VendelClient:
    """Client for the Vendel SMS gateway API.

    Uses an integration API key (``vk_`` prefix) for authentication.
    """

    def __init__(self, base_url: str, api_key: str, timeout: int = 30):
        self.base_url = base_url.rstrip("/")
        self.api_key = api_key
        self.timeout = timeout
        self._session = requests.Session()
        self._session.headers.update({"X-API-Key": api_key})

    # ------------------------------------------------------------------
    # Public API
    # ------------------------------------------------------------------

    def send_sms(
        self,
        recipients: list[str],
        body: str,
        device_id: str | None = None,
    ) -> SendSMSResponse:
        """Send an SMS to one or more recipients.

        Args:
            recipients: Phone numbers in E.164 format (e.g. ``+1234567890``).
            body: Message text (max 1600 characters).
            device_id: Optional device to route through.

        Returns:
            A :class:`SendSMSResponse` with batch ID and message IDs.

        Raises:
            VendelQuotaError: If the monthly SMS quota is exceeded.
            VendelAPIError: On any other API error.
        """
        payload: dict = {"recipients": recipients, "body": body}
        if device_id:
            payload["device_id"] = device_id
        data = self._post("/api/sms/send", payload)
        return SendSMSResponse.from_dict(data)

    def get_quota(self) -> Quota:
        """Get the current quota for the authenticated user.

        Returns:
            A :class:`Quota` with plan limits and usage.
        """
        data = self._get("/api/plans/quota")
        return Quota.from_dict(data)

    # ------------------------------------------------------------------
    # Internal helpers
    # ------------------------------------------------------------------

    def _get(self, path: str) -> dict:
        resp = self._session.get(f"{self.base_url}{path}", timeout=self.timeout)
        return self._handle_response(resp)

    def _post(self, path: str, json: dict) -> dict:
        resp = self._session.post(
            f"{self.base_url}{path}", json=json, timeout=self.timeout
        )
        return self._handle_response(resp)

    @staticmethod
    def _handle_response(resp: requests.Response) -> dict:
        if resp.status_code == 429:
            data = resp.json()
            raise VendelQuotaError(
                data.get("detail", "Quota exceeded"),
                data,
            )
        if not resp.ok:
            try:
                data = resp.json()
            except ValueError:
                data = {}
            raise VendelAPIError(
                resp.status_code,
                data.get("message", resp.reason),
                data,
            )
        return resp.json()
