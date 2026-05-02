import { AxiosError, isAxiosError } from 'axios';
import type { ErrorResponse } from '@/shared/types/api';
import { ERROR_MESSAGES } from '@/shared/config/constants';

/**
 * Словарь переводов стандартных английских сообщений от backend на русский
 */
const MESSAGE_TRANSLATIONS: Record<string, string> = {
  // Auth & Permissions
  'Invalid or missing token': 'Неверный или отсутствующий токен',
  'Admin role required': 'Требуется роль администратора',
  'Guard role required': 'Требуется роль охранника',
  'Invalid username or password': 'Неверное имя пользователя или пароль',
  'Username and password are required': 'Требуется имя пользователя и пароль',
  'Invalid refresh token': 'Неверный токен обновления',
  
  // Residents
  'apartment_id, telegram_id, and chat_id are required': 'Обязательные поля: ID квартиры, Telegram ID и Chat ID',
  'Resident with this telegram_id already exists': 'Житель с таким Telegram ID уже существует',
  'Resident not found': 'Житель не найден',
  'Body must be an array of residents': 'Тело запроса должно содержать массив жителей',
  'apartment not found': 'Квартира не найдена',
  'telegram_id already exists': 'Telegram ID уже существует',
  'invalid telegram_id': 'Некорректный Telegram ID',
  'invalid chat_id': 'Некорректный Chat ID',
  'invalid apartment_id': 'Некорректный ID квартиры',
  'apartment_number is required for admin': 'Для администратора обязателен номер квартиры',
  'apartment_number is required when apartment_id is not provided': 'Требуется номер квартиры',
  'building_id is required when apartment_id is not provided': 'Требуется ID здания',
  'apartment not found in building': 'Квартира не найдена в указанном здании',
  'missing required fields': 'Отсутствуют обязательные поля',
  'invalid phone format': 'Неверный формат телефона',
  
  // Rules
  'building_id query parameter is required': 'Требуется параметр building_id',
  'Rules not found for this building': 'Правила не найдены для этого здания',
  
  // Passes
  'Pass not found': 'Пропуск не найден',
  'Pass already revoked': 'Пропуск уже отозван',
  'Pass expired': 'Срок действия пропуска истек',
  'Pass not yet valid': 'Пропуск еще не действителен',
  
  // Registration
  'Username already exists': 'Имя пользователя уже существует',
  'Username, password, and role are required': 'Требуется имя пользователя, пароль и роль',
  
  // General
  'Invalid request': 'Некорректный запрос',
  'Unauthorized': 'Требуется авторизация',
  'Forbidden': 'Доступ запрещен',
  'Not found': 'Не найдено',
  'Internal server error': 'Внутренняя ошибка сервера',
  'Network Error': 'Ошибка сети. Проверьте подключение',
  'Request failed with status code 401': 'Требуется авторизация',
  'Request failed with status code 403': 'Доступ запрещён',
  'Request failed with status code 404': 'Ресурс не найден',
  'Request failed with status code 500': 'Ошибка на сервере. Попробуйте позже',
};

/** Сообщения по коду причины проверки пропуска (совпадают с backend). */
export const PASS_REASON_MESSAGES: Record<string, string> = {
  PASS_NOT_FOUND: 'Пропуск не найден',
  PASS_EXPIRED: 'Срок действия пропуска истек',
  PASS_REVOKED: 'Пропуск отозван',
  PASS_NOT_YET_VALID: 'Пропуск ещё не действителен',
  PASS_ALREADY_USED: 'Пропуск уже был использован',
  QUIET_HOURS: 'Действие запрещено в тихие часы',
  INVALID_CAR_PLATE: 'Некорректный номер автомобиля',
  INVALID_PERSONAL_PASS: 'Недействительный личный пропуск',
  BUILDING_MISMATCH: 'Пропуск относится к другому зданию',
  APARTMENT_NOT_FOUND: 'Квартира не найдена',
  RESIDENT_NOT_FOUND: 'Житель не найден',
};

/**
 * Форматирует сообщение об ошибке для отображения пользователю
 * 
 * @param error - Ошибка от Axios
 * @returns Отформатированное сообщение на русском языке
 */
function containsCyrillic(text: string): boolean {
  return /[а-яёА-ЯЁ]/.test(text);
}

export function formatErrorMessage(error: AxiosError<ErrorResponse>): string {
  if (error.code === 'ERR_NETWORK' || !error.response) {
    return ERROR_MESSAGES.NETWORK_ERROR;
  }

  const errorCode = error.response.data?.error?.code || 'UNKNOWN_ERROR';
  const serverMessage = error.response.data?.error?.message?.trim();

  // Сообщение с сервера уже на русском — показываем его целиком (без дублирования с заголовком по коду).
  if (serverMessage && containsCyrillic(serverMessage)) {
    let msg = serverMessage;
    if (import.meta.env.DEV) {
      msg += ` (HTTP ${error.response.status})`;
    }
    return msg;
  }

  let errorMessage = ERROR_MESSAGES[errorCode] || ERROR_MESSAGES.UNKNOWN_ERROR;

  if (serverMessage) {
    const translatedMessage = MESSAGE_TRANSLATIONS[serverMessage] || serverMessage;
    const appendDetail =
      translatedMessage !== errorMessage &&
      (MESSAGE_TRANSLATIONS[serverMessage] !== undefined || containsCyrillic(translatedMessage));

    if (appendDetail) {
      errorMessage = `${errorMessage}: ${translatedMessage}`;
    }
  }

  if (import.meta.env.DEV) {
    errorMessage += ` (HTTP ${error.response.status})`;
  }

  return errorMessage;
}

/**
 * Совместимый хелпер для обработки неизвестных ошибок в UI-слое.
 *
 * @param error - Ошибка любого типа
 * @returns Сообщение для отображения пользователю
 */
export function getErrorMessage(error: unknown): string {
  if (isAxiosError<ErrorResponse>(error)) {
    return formatErrorMessage(error);
  }

  if (error instanceof Error && error.message.trim()) {
    return translateMessage(error.message.trim());
  }

  return ERROR_MESSAGES.UNKNOWN_ERROR;
}

/**
 * Добавляет новый перевод для сообщения от backend
 * Полезно для динамического расширения словаря переводов
 * 
 * @param englishMessage - Английское сообщение от backend
 * @param russianTranslation - Русский перевод
 */
export function addMessageTranslation(englishMessage: string, russianTranslation: string): void {
  MESSAGE_TRANSLATIONS[englishMessage] = russianTranslation;
}

/**
 * Переводит английское сообщение на русский или возвращает оригинал
 * 
 * @param message - Сообщение для перевода
 * @returns Переведенное сообщение или оригинал
 */
export function translateMessage(message: string): string {
  return MESSAGE_TRANSLATIONS[message] || message;
}

/**
 * Форматирует ошибку из bulk операции (создание/импорт) в человекочитаемый вид
 * 
 * @param errorItem - Объект ошибки с полями row и error
 * @returns Отформатированная строка ошибки на русском языке
 * 
 * @example
 * formatBulkError({ row: 1, error: "apartment not found" })
 * // => "Строка 1: Квартира не найдена"
 */
export function formatBulkError(errorItem: { row?: number; error: string } | string): string {
  // Если передана строка, пытаемся распарсить как JSON
  let item: { row?: number; error: string };
  
  if (typeof errorItem === 'string') {
    try {
      const parsed = JSON.parse(errorItem) as { row?: number; error: string };
      if (typeof parsed === 'object' && parsed !== null && 'error' in parsed) {
        item = parsed;
      } else {
        return translateMessage(errorItem);
      }
    } catch {
      // Если не JSON, возвращаем как есть с переводом
      return translateMessage(errorItem);
    }
  } else {
    item = errorItem;
  }

  const { row, error } = item;
  
  // Переводим текст ошибки
  const translatedError = translateMessage(error);
  
  // Форматируем в зависимости от наличия номера строки
  if (row !== undefined && row !== null) {
    return `Строка ${row}: ${translatedError}`;
  }
  
  return translatedError;
}

