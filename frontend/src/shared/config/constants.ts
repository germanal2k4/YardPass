export const APP_ROUTES = {
  HOME: '/',
  LOGIN: '/login',
  REGISTER: '/register',
  SECURITY: '/security',
  ADMIN: '/admin',
  ADMIN_RULES: '/admin/rules',
  ADMIN_REPORTS: '/admin/reports',
  FORBIDDEN: '/forbidden',
} as const;

export const API_ENDPOINTS = {
  // Auth
  LOGIN: '/auth/login',
  REFRESH: '/auth/refresh',
  ME: '/api/v1/me',
  
  // Passes
  PASSES: '/api/v1/passes',
  PASS_BY_ID: (id: string) => `/api/v1/passes/${id}`,
  REVOKE_PASS: (id: string) => `/api/v1/passes/${id}/revoke`,
  VALIDATE_PASS: '/api/v1/passes/validate',
  ACTIVE_PASSES: '/api/v1/passes/active',
  
  // Rules
  RULES: '/api/v1/rules',
  
  // Scan Events & Reports
  SCAN_EVENTS: '/api/v1/scan-events',
  STATISTICS: '/api/v1/reports/statistics',
  EXPORT_REPORT: '/api/v1/reports/export',
  
  // Health
  HEALTH: '/health',
} as const;

export const ERROR_MESSAGES: Record<string, string> = {
  PASS_NOT_FOUND: 'Пропуск не найден',
  PASS_EXPIRED: 'Срок действия пропуска истек',
  PASS_REVOKED: 'Пропуск отозван',
  PASS_NOT_YET_VALID: 'Пропуск еще не действителен',
  QUIET_HOURS: 'Действие запрещено в тихие часы',
  RATE_LIMIT_EXCEEDED: 'Превышен лимит запросов',
  INVALID_CREDENTIALS: 'Неверные учетные данные',
  INVALID_TOKEN: 'Неверный или истекший токен',
  INSUFFICIENT_PERMISSIONS: 'Недостаточно прав',
  NETWORK_ERROR: 'Ошибка сети. Проверьте подключение',
  UNKNOWN_ERROR: 'Произошла неизвестная ошибка',
};

export const STORAGE_KEYS = {
  ACCESS_TOKEN: 'access_token',
  REFRESH_TOKEN: 'refresh_token',
} as const;

