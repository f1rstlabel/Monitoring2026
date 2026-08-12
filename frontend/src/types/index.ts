export type DeviceType = 'Access Point' | 'Switch' | 'Router' | 'SmartPower' | 'CCTV' | 'NVR';
export type DeviceStatus = 'UP' | 'DOWN';
export type AddressingMode = 'Static' | 'DHCP';
export type UserRole = 'admin' | 'pimpinan' | 'anggota';
export type SeverityLevel = 'critical' | 'warning' | 'info' | 'skipped';

export interface Device {
  id: string;
  name: string;
  type: DeviceType;
  ip: string;
  mac: string;
  status: DeviceStatus;
  addressingMode: AddressingMode;
  locationId?: string;
  location: string;
  rack?: string;
  checkedSecondsAgo: number;
  lastChecked: string;
  uptime30d: number;
  failureThreshold: number;
  useCustomThreshold?: boolean;
  customFailureThreshold?: number | null;
  model?: string;
  firmwareStatus?: string;
  snmpEnabled?: boolean;
  snmpCommunity?: string;
  snmpPort?: number;
  snmpIfIndex?: number;
  snmpSysName?: string;
  snmpSysDescr?: string;
  snmpSysUpTime?: string;
  snmpSysContact?: string;
  snmpSysLocation?: string;
  latencyMs?: number;
  createdByUserId?: string;
  createdByUserName?: string;
}

export interface PaginatedResponse<T> {
  items: T[];
  data: T[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

export interface DeviceMetric {
  id: string;
  deviceId: string;
  metricType: 'cpu' | 'memory';
  value: number;
  recordedAt: string;
}

export interface DashboardSummary {
  totalDevices: number;
  devicesUp: number;
  devicesDown: number;
  activeIncidents: number;
  upPercentage: number;
  downPercentage: number;
}

export interface EventTimelineItem {
  id: string;
  timestamp: string;
  title: string;
  description: string;
  severity: SeverityLevel;
  channel?: string;
  status?: string;
}

export interface NotificationLogRow {
  id: string;
  channel: string;
  channelIcon: string;
  recipient: string;
  status: 'Delivered' | 'Failed' | 'Sent';
  timestamp: string;
}

export interface Incident {
  id: string;
  deviceId: string;
  deviceName: string;
  deviceType: DeviceType;
  deviceIp: string;
  location?: string;
  status: 'ACTIVE' | 'RESOLVED';
  startTime: string;
  duration: string;
  affectedDevicesCount: number;
  packetLoss: number;
  latencyMs: number;
  dependenciesCount: number;
  timeline: EventTimelineItem[];
  notificationLog: NotificationLogRow[];
  notes?: string[];
  resolvedAt?: string;
  startedAt?: string;
}

export interface StatusHistoryPoint {
  date: string;
  upCount: number;
  downCount: number;
}

export interface WhatsAppTarget {
  id: string;
  label: string;
  phoneNumber: string;
  jid: string;
  createdAt?: string;
  updatedAt?: string;
}

export interface NotificationChannelSetting {
  id: string;
  name: string;
  type: 'whatsapp' | 'telegram';
  connected: boolean;
  handleOrNumber: string;
  lastSync: string;
}

export interface ThresholdDefault {
  type: DeviceType;
  consecutiveFailures: number;
}

export interface PollingEngineConfig {
  intervalSeconds: number;
  concurrencyBatchSize: number;
  debounceSeconds?: number;
  flapReuseWindowMinutes?: number;
}

export interface SystemSettings {
  rateLimitMaxMsgPerMin: number;
  thresholds: ThresholdDefault[];
  polling: PollingEngineConfig;
}

export interface User {
  id: string;
  username?: string;
  name: string;
  email: string;
  role: UserRole;
  status: 'Active' | 'Inactive';
  avatarUrl: string;
  lastActive: string;
  permissions?: string[];
  isActive?: boolean;
}

export interface LiveFeedItem {
  id: string;
  timestamp: string;
  title: string;
  description: string;
  severity: SeverityLevel;
  deviceId?: string;
}

// ─── Bulk Import Types ───────────────────────────────────────────────────────

export type ImportRowStatus = 'valid' | 'duplicate_mac' | 'missing_field' | 'invalid_mac' | 'failed';

export interface ImportRow {
  rowIndex: number;
  raw: Record<string, string>;
  mapped: Partial<Device>;
  status: ImportRowStatus;
  reason?: string;
}

export interface ImportResult {
  rowIndex: number;
  status: 'imported' | 'skipped' | 'failed';
  reason?: string;
  deviceId?: string;
}

export interface ImportSummary {
  imported: number;
  skipped: number;
  failed: number;
  results: ImportResult[];
}

// ─── Report Types ────────────────────────────────────────────────────────────

export interface ReportRow {
  deviceId: string;
  deviceName: string;
  deviceType: DeviceType;
  location: string;
  downCount: number;
  totalDowntimeMinutes: number;
  lastDown: string;
}

export interface FlapDevice {
  deviceId: string;
  deviceName: string;
  deviceType: DeviceType;
  location: string;
  ip: string;
  downCount7d: number;
  totalDowntimeMinutes: number;
}

export interface ColumnMapping {
  sourceColumn: string;
  targetField: keyof Device | 'model' | 'firmwareStatus' | '';
}
