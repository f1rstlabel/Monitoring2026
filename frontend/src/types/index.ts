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

export type PublicMonitorStatus = 'UP' | 'DOWN' | 'PAUSED';

export type PublicMonitorType = 'http' | 'http_keyword' | 'http_json' | 'tcp' | 'ping' | 'dns';

export interface PublicMonitor {
  id: string;
  name: string;
  monitorType: PublicMonitorType;
  targetUrl: string;
  targetHost?: string;
  targetPort?: number;
  dnsRecordType?: 'A' | 'AAAA' | 'CNAME' | string;
  keyword?: string;
  jsonPath?: string;
  expectedValue?: string;
  groupId?: string;
  groupName: string;
  intervalSeconds: number;
  timeoutSeconds: number;
  retryCount: number;
  enabled: boolean;
  notifyOnFailure: boolean;
  status: PublicMonitorStatus;
  uptime24h: number;
  uptime30d: number;
  lastStatusCode: number;
  lastLatencyMs: number;
  lastError?: string;
  consecutiveFailures: number;
  lastChecked?: string;
  deletedAt?: string;
  deletedByUserId?: string;
  deletionReason?: string;
  purgedAt?: string;
  purgedByUserId?: string;
  purgeReason?: string;
}

export interface PublicMonitorGroup {
  id: string;
  name: string;
  description?: string;
  displayOrder: number;
  enabled: boolean;
  monitorCount: number;
  createdAt?: string;
  updatedAt?: string;
}

export interface PublicMonitorCheck {
  id: string;
  monitorId: string;
  checkedAt: string;
  status: 'UP' | 'DOWN';
  statusCode: number;
  latencyMs: number;
  errorMessage?: string;
}

export interface PublicMonitorEvent {
  id: string;
  monitorId: string;
  eventType: 'down' | 'recovered' | string;
  status: PublicMonitorStatus;
  message: string;
  occurredAt: string;
}

export interface PublicMonitorDetail {
  monitor: PublicMonitor;
  checks: PublicMonitorCheck[];
  events: PublicMonitorEvent[];
}

export type PublicMonitorIncidentStatus = 'ACTIVE' | 'RESOLVED';

export interface PublicMonitorIncident {
  id: string;
  monitorId: string;
  monitorName: string;
  monitorType?: PublicMonitorType | string;
  targetUrl: string;
  status: PublicMonitorIncidentStatus;
  startedAt: string;
  resolvedAt?: string;
  durationSeconds: number;
  initialStatusCode: number;
  finalStatusCode: number;
  firstError?: string;
  lastError?: string;
  resolutionReason?: string;
}

export interface PublicMonitorIncidentEvent {
  id: string;
  incidentId: string;
  eventType: string;
  channel?: string;
  detail: string;
  occurredAt: string;
}

export interface PublicMonitorNotificationLog {
  id: string;
  incidentId: string;
  channel: string;
  recipient: string;
  status: string;
  error?: string;
  sentAt: string;
}

export interface PublicMonitorIncidentDetail {
  incident: PublicMonitorIncident;
  events: PublicMonitorIncidentEvent[];
  notifications: PublicMonitorNotificationLog[];
}

export interface PublicMonitorReportPoint {
  bucket: string;
  upCount: number;
  downCount: number;
  uptimePercent: number;
  avgLatencyMs: number;
}

export interface PublicMonitorSummary {
  monitorId: string;
  monitorName: string;
  monitorType: string;
  targetUrl: string;
  groupName: string;
  status: PublicMonitorStatus;
  enabled: boolean;
  lastChecked?: string;
  lastLatencyMs: number;
  lastError?: string;
  totalChecks: number;
  upChecks: number;
  downChecks: number;
  uptimePercent: number;
  avgLatencyMs: number;
  incidentCount: number;
}

export interface PublicMonitorReport {
  monitorId?: string;
  monitorName?: string;
  from: string;
  to: string;
  totalChecks: number;
  upChecks: number;
  downChecks: number;
  uptimePercent: number;
  downtimeMinutes: number;
  incidentCount: number;
  avgLatencyMs: number;
  points: PublicMonitorReportPoint[];
  monitorSummaries?: PublicMonitorSummary[];
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
  status: 'Delivered' | 'Failed' | 'Sent' | 'Skipped';
  errorMsg?: string;
  timestamp: string;
}

export interface Incident {
  id: string;
  source?: 'DEVICE' | 'PUBLIC_MONITOR' | string;
  sourceId?: string;
  category?: string;
  targetUrl?: string;
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

export interface CoreSwitchConfig {
  ip: string;
  community: string;
  port: number;
  version: string;
}

export interface SystemSettings {
  rateLimitMaxMsgPerMin: number;
  thresholds: ThresholdDefault[];
  polling: PollingEngineConfig;
  coreSwitch?: CoreSwitchConfig;
  retentionDays?: number;
}

export interface LocationItem {
  id: string;
  name: string;
  description?: string;
  createdAt?: string;
  deviceCount?: number;
}

export interface BulkDeviceUpdates {
  locationId?: string;
  location?: string;
  snmpEnabled?: boolean;
  type?: DeviceType;
}

export interface BulkDeviceRequest {
  deviceIds: string[];
  action: 'update' | 'delete';
  updates?: BulkDeviceUpdates;
}

export interface BulkDeviceResponse {
  success: boolean;
  updatedCount: number;
  failedCount: number;
  details?: { deviceId: string; status: string; reason?: string }[];
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
  mfaEnabled?: boolean;
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

// ─── Bulk Operations Types ──────────────────────────────────────────────────

export interface BulkDeviceUpdates {
  locationId?: string;
  location?: string;
  rack?: string;
  type?: DeviceType;
  addressingMode?: 'Static' | 'DHCP';
  snmpEnabled?: boolean;
  snmpCommunity?: string;
  snmpPort?: number;
  snmpIfIndex?: number;
  useCustomThreshold?: boolean;
  customFailureThreshold?: number;
  failureThreshold?: number;
}

export interface BulkDeviceRequest {
  deviceIds: string[];
  action: 'update' | 'delete';
  updates?: BulkDeviceUpdates;
}

export interface BulkDeviceResponse {
  success: boolean;
  updatedCount: number;
  failedCount: number;
  details?: Array<{ deviceId: string; status: string; reason?: string }>;
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

export interface BrandingSettings {
  appTitle: string;
  appSubtitle: string;
  logoUrl: string;
  logoFit?: 'cover' | 'contain';
  logoScale?: number;
  faviconUrl: string;
  footerText: string;
}

export interface PingResult {
  target: string;
  success: boolean;
  durationMs: number;
  output: string[];
  raw: string;
}

export interface TracerouteResult {
  target: string;
  durationMs: number;
  output: string[];
  raw: string;
}

export interface PortProbeResult {
  target: string;
  port: number;
  open: boolean;
  latencyMs: number;
  message: string;
  error?: string;
}

export interface IPChangeEvent {
  id: string;
  deviceId: string;
  deviceName?: string;
  oldIp: string;
  newIp: string;
  source?: 'KEA_DHCP' | 'L3_ARP' | string;
  timestamp: string;
}

