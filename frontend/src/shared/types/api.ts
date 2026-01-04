// Domain models from backend

export interface User {
  id: number;
  username: string;
  email?: string;
  role: 'guard' | 'admin';
  status: string;
  created_at: string;
  updated_at: string;
}

export interface Building {
  id: number;
  name: string;
  address: string;
  created_at: string;
  updated_at: string;
}

export interface Apartment {
  id: number;
  building_id: number;
  number: string;
  floor?: number;
  created_at: string;
  updated_at: string;
}

export interface Pass {
  id: string; // UUID
  apartment_id: number;
  car_plate: string;
  guest_name?: string;
  valid_from: string; // ISO datetime
  valid_to: string; // ISO datetime
  status: 'active' | 'revoked' | 'expired';
  created_at: string;
  updated_at: string;
}

export interface ScanEvent {
  id: number;
  pass_id: string; // UUID
  guard_user_id: number;
  scanned_at: string; // ISO datetime
  result: 'valid' | 'invalid';
  reason?: string;
  meta?: string;
}

export interface Rule {
  id: number;
  building_id: number;
  quiet_hours_start?: string; // HH:mm format
  quiet_hours_end?: string; // HH:mm format
  daily_pass_limit_per_apartment: number;
  max_pass_duration_hours: number;
  created_at: string;
  updated_at: string;
}

// API Request/Response DTOs

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  token_type: 'Bearer';
}

export interface RefreshRequest {
  refresh_token: string;
}

export interface MeResponse {
  user_id: number;
  role: 'guard' | 'admin';
}

export interface CreatePassRequest {
  apartment_id: number;
  car_plate: string;
  guest_name?: string;
  valid_from?: string; // ISO datetime
  valid_to: string; // ISO datetime
}

export interface ValidatePassRequest {
  qr_uuid: string; // UUID from QR code
}

export interface ValidatePassResponse {
  valid: boolean;
  car_plate?: string;
  apartment?: string;
  valid_to?: string; // ISO datetime
  reason?: 'PASS_NOT_FOUND' | 'PASS_EXPIRED' | 'PASS_REVOKED' | 'PASS_NOT_YET_VALID' | 'QUIET_HOURS';
}

export interface GetActivePassesResponse {
  passes: Pass[];
}

export interface UpdateRuleRequest {
  quiet_hours_start?: string; // HH:mm
  quiet_hours_end?: string; // HH:mm
  daily_pass_limit_per_apartment?: number;
  max_pass_duration_hours?: number;
}

export interface RevokePassResponse {
  message: string;
  pass_id: string;
}

// Error response from backend

export interface ErrorResponse {
  error: {
    code: string;
    message: string;
  };
}

// Scan event filters for reports

export interface ScanEventFilters {
  pass_id?: string;
  guard_user_id?: number;
  result?: 'valid' | 'invalid';
  from?: string; // ISO datetime
  to?: string; // ISO datetime
  limit?: number;
  offset?: number;
}

