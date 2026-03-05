export interface SendSMSRequest {
  recipients: string[];
  body: string;
  device_id?: string;
}

export interface SendSMSResponse {
  batch_id: string;
  message_ids: string[];
  recipients_count: number;
  status: string;
}

export interface Quota {
  plan: string;
  sms_sent_this_month: number;
  max_sms_per_month: number;
  devices_registered: number;
  max_devices: number;
  reset_date: string;
}

export interface VendelClientOptions {
  baseUrl: string;
  apiKey: string;
  /** Request timeout in milliseconds (default: 30000). */
  timeout?: number;
}
