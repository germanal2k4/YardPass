import { AxiosError } from 'axios';
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
};

/**
 * Форматирует сообщение об ошибке для отображения пользователю
 * 
 * @param error - Ошибка от Axios
 * @returns Отформатированное сообщение на русском языке
 */
export function formatErrorMessage(error: AxiosError<ErrorResponse>): string {
  // Получаем код и сообщение ошибки от сервера
  const errorCode = error.response?.data?.error?.code || 'UNKNOWN_ERROR';
  const serverMessage = error.response?.data?.error?.message;
  
  // Базовое сообщение из словаря кодов ошибок
  let errorMessage = ERROR_MESSAGES[errorCode] || ERROR_MESSAGES.UNKNOWN_ERROR;
  
  // Если есть сообщение от сервера, добавляем его
  if (serverMessage) {
    // Пытаемся найти перевод
    const translatedMessage = MESSAGE_TRANSLATIONS[serverMessage] || serverMessage;
    
    // Если перевод отличается от базового сообщения, добавляем детали
    if (translatedMessage !== errorMessage) {
      errorMessage = `${errorMessage}: ${translatedMessage}`;
    }
  }
  
  // Обработка сетевых ошибок
  if (error.code === 'ERR_NETWORK' || !error.response) {
    return ERROR_MESSAGES.NETWORK_ERROR;
  }
  
  // Добавляем HTTP статус для отладки (только в dev режиме)
  if (import.meta.env.DEV && error.response) {
    errorMessage += ` (HTTP ${error.response.status})`;
  }
  
  return errorMessage;
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
  if (typeof errorItem === 'string') {
    try {
      errorItem = JSON.parse(errorItem);
    } catch {
      // Если не JSON, возвращаем как есть с переводом
      return translateMessage(errorItem);
    }
  }

  // Проверяем, что это объект
  if (typeof errorItem !== 'object' || errorItem === null) {
    return String(errorItem);
  }

  const { row, error } = errorItem as { row?: number; error: string };
  
  // Переводим текст ошибки
  const translatedError = translateMessage(error);
  
  // Форматируем в зависимости от наличия номера строки
  if (row !== undefined && row !== null) {
    return `Строка ${row}: ${translatedError}`;
  }
  
  return translatedError;
}

