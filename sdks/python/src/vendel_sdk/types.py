from __future__ import annotations

from dataclasses import dataclass


@dataclass
class SendSMSResponse:
    batch_id: str
    message_ids: list[str]
    recipients_count: int
    status: str

    @classmethod
    def from_dict(cls, data: dict) -> SendSMSResponse:
        return cls(
            batch_id=data["batch_id"],
            message_ids=data["message_ids"],
            recipients_count=data["recipients_count"],
            status=data["status"],
        )


@dataclass
class Quota:
    plan: str
    sms_sent_this_month: int
    max_sms_per_month: int
    devices_registered: int
    max_devices: int
    reset_date: str

    @classmethod
    def from_dict(cls, data: dict) -> Quota:
        return cls(
            plan=data["plan"],
            sms_sent_this_month=data["sms_sent_this_month"],
            max_sms_per_month=data["max_sms_per_month"],
            devices_registered=data["devices_registered"],
            max_devices=data["max_devices"],
            reset_date=data.get("reset_date", ""),
        )
