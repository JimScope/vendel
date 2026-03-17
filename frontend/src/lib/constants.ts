import { getBrowserTimezone } from "./datetime"

const BASE_TIMEZONES = [
  "UTC",
  "America/New_York",
  "America/Chicago",
  "America/Denver",
  "America/Los_Angeles",
  "America/Havana",
  "Europe/London",
  "Europe/Paris",
  "Europe/Berlin",
  "Europe/Madrid",
  "Asia/Tokyo",
  "Asia/Shanghai",
  "Asia/Kolkata",
  "Australia/Sydney",
]

/** Timezone list that always includes the browser's timezone. */
export const COMMON_TIMEZONES: string[] = BASE_TIMEZONES.includes(
  getBrowserTimezone(),
)
  ? BASE_TIMEZONES
  : [getBrowserTimezone(), ...BASE_TIMEZONES]
