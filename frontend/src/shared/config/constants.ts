export const APP_ROUTES = {
  HOME: '/',
  LOGIN: '/login',
  REGISTER: '/register',
  SECURITY: '/security',
  ADMIN: '/admin',
  ADMIN_RULES: '/admin/rules',
  ADMIN_REPORTS: '/admin/reports',
  ADMIN_GUARDS: '/admin/guards',
  ADMIN_RESIDENTS: '/admin/residents',
  FORBIDDEN: '/forbidden',
} as const;

export const API_ENDPOINTS = {
  // Auth
  LOGIN: '/auth/login',
  REFRESH: '/auth/refresh',
  LOGOUT: '/auth/logout',
  PURCHASE_SUBSCRIPTION: '/auth/purchase-subscription',
  ME: '/api/v1/me',
  
  // Passes
  PASSES: '/api/v1/passes',
  PASS_BY_ID: (id: string) => `/api/v1/passes/${id}`,
  REVOKE_PASS: (id: string) => `/api/v1/passes/${id}/revoke`,
  VALIDATE_PASS: '/api/v1/passes/validate',
  ACTIVE_PASSES: '/api/v1/passes/active',
  
  // Rules
  RULES: '/api/v1/rules',
  USERS: '/api/v1/users',
  BUILDING_BY_ID: (id: number) => `/api/v1/buildings/${id}`,
  BUILDING_APARTMENT_COUNT: (id: number) => `/api/v1/buildings/${id}/apartment-count`,
  
  // Residents
  RESIDENTS: '/api/v1/residents',
  RESIDENTS_BULK: '/api/v1/residents/bulk',
  RESIDENTS_IMPORT: '/api/v1/residents/import',
  
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
  PASS_ALREADY_USED: 'Пропуск уже был использован',
  QUIET_HOURS: 'Действие запрещено в тихие часы',
  RATE_LIMIT_EXCEEDED: 'Превышен лимит запросов',
  INVALID_CREDENTIALS: 'Неверные учетные данные',
  INVALID_TOKEN: 'Неверный или истекший токен',
  INVALID_REFRESH_TOKEN: 'Сессия устарела. Войдите снова',
  MISSING_TOKEN: 'Требуется вход в систему',
  INSUFFICIENT_PERMISSIONS: 'Недостаточно прав',
  MISSING_ROLE: 'Роль пользователя не найдена',
  INVALID_ROLE: 'Некорректная роль в сессии',
  MISSING_SERVICE_TOKEN: 'Требуется служебный токен',
  INVALID_SERVICE_TOKEN: 'Неверный служебный токен',
  NETWORK_ERROR: 'Ошибка сети. Проверьте подключение',
  UNKNOWN_ERROR: 'Произошла неизвестная ошибка',
  INTERNAL_ERROR: 'Внутренняя ошибка сервера. Попробуйте позже',
  SERVICE_UNAVAILABLE: 'Сервис временно недоступен',
  // Passes
  CREATE_PASS_FAILED: 'Не удалось создать пропуск',
  REVOKE_FAILED: 'Не удалось отозвать пропуск',
  VALIDATION_ERROR: 'Не удалось проверить пропуск',
  INVALID_UUID: 'Некорректный идентификатор пропуска',
  MISSING_PARAMETER: 'Не хватает параметров запроса',
  MISSING_CAR_PLATE: 'Укажите номер автомобиля',
  INVALID_APARTMENT_ID: 'Некорректный номер квартиры',
  FETCH_FAILED: 'Не удалось загрузить данные',
  SEARCH_FAILED: 'Не удалось выполнить поиск',
  INVALID_PERSONAL_PASS: 'Недействительный личный пропуск',
  BUILDING_MISMATCH: 'Пропуск от другого жилого комплекса',
  APARTMENT_NOT_FOUND: 'Квартира не найдена',
  INVALID_CAR_PLATE: 'Некорректный номер автомобиля',
  // Residents
  RESIDENT_EXISTS: 'Житель с таким Telegram ID уже существует',
  RESIDENT_NOT_FOUND: 'Житель не найден',
  CREATE_FAILED: 'Не удалось добавить жителя',
  DELETE_FAILED: 'Не удалось удалить жителя',
  MISSING_APARTMENT_NUMBER: 'Укажите номер квартиры',
  INVALID_ID: 'Некорректный идентификатор',
  MISSING_FILE: 'Прикрепите файл',
  FILE_OPEN_ERROR: 'Не удалось прочитать файл',
  INVALID_REQUEST: 'Некорректный запрос',
  UNAUTHORIZED: 'Требуется авторизация',
  FORBIDDEN: 'Доступ запрещен',
  // Rules
  RULE_NOT_FOUND: 'Правила не найдены',
  UPDATE_FAILED: 'Не удалось сохранить правила',
  MISSING_BUILDING_ID: 'Не указан ID здания',
  INVALID_BUILDING_ID: 'Некорректный ID здания',
  // Building / reports / subscription
  BUILDING_NOT_FOUND: 'Здание не найдено',
  FETCH_BUILDING_FAILED: 'Не удалось загрузить сведения о здании',
  UPDATE_BUILDING_FAILED: 'Не удалось обновить здание',
  INVALID_APARTMENT_COUNT: 'Некорректное количество квартир',
  INVALID_FORMAT: 'Некорректный формат экспорта',
  EXCEL_ERROR: 'Не удалось сформировать отчёт',
  PURCHASE_FAILED: 'Не удалось оформить подписку',
  REGISTRATION_FAILED: 'Не удалось зарегистрировать пользователя',
  NOT_IMPLEMENTED: 'Функция пока недоступна',
};
