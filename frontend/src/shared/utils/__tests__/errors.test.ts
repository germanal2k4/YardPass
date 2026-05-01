import { describe, it, expect } from 'vitest';
import { AxiosError, AxiosHeaders, InternalAxiosRequestConfig } from 'axios';
import {
  formatErrorMessage,
  translateMessage,
  formatBulkError,
  addMessageTranslation,
} from '../errors';
import type { ErrorResponse } from '@/shared/types/api';

function makeAxiosError(
  overrides: {
    status?: number;
    code?: string;
    errorCode?: string;
    errorMessage?: string;
    axiosCode?: string;
    hasResponse?: boolean;
  } = {},
): AxiosError<ErrorResponse> {
  const {
    status = 400,
    errorCode = 'UNKNOWN_ERROR',
    errorMessage = '',
    axiosCode,
    hasResponse = true,
  } = overrides;

  const config: InternalAxiosRequestConfig = {
    headers: new AxiosHeaders(),
  };

  const error = new AxiosError<ErrorResponse>(
    'Request failed',
    axiosCode,
    config,
    {},
    hasResponse
      ? {
          data: { error: { code: errorCode, message: errorMessage } },
          status,
          statusText: 'Bad Request',
          headers: {},
          config,
        }
      : undefined,
  );

  return error;
}

// --------------- translateMessage ---------------

describe('translateMessage', () => {
  it('translates a known English message to Russian', () => {
    expect(translateMessage('Invalid username or password')).toBe(
      'Неверное имя пользователя или пароль',
    );
  });

  it('returns the original if translation is missing', () => {
    expect(translateMessage('Some unknown phrase')).toBe('Some unknown phrase');
  });
});

// --------------- addMessageTranslation ---------------

describe('addMessageTranslation', () => {
  it('registers a new translation that translateMessage can use', () => {
    addMessageTranslation('Custom error', 'Пользовательская ошибка');
    expect(translateMessage('Custom error')).toBe('Пользовательская ошибка');
  });
});

// --------------- formatBulkError ---------------

describe('formatBulkError', () => {
  it('formats an object with row and error', () => {
    const result = formatBulkError({ row: 1, error: 'apartment not found' });
    expect(result).toBe('Строка 1: Квартира не найдена');
  });

  it('formats an object without row', () => {
    const result = formatBulkError({ error: 'missing required fields' });
    expect(result).toBe('Отсутствуют обязательные поля');
  });

  it('parses a JSON string into an object', () => {
    const json = JSON.stringify({ row: 3, error: 'invalid telegram_id' });
    expect(formatBulkError(json)).toBe('Строка 3: Некорректный Telegram ID');
  });

  it('handles a plain (non-JSON) string with translation', () => {
    expect(formatBulkError('Resident not found')).toBe('Житель не найден');
  });

  it('returns original string if no translation exists', () => {
    expect(formatBulkError('random error text')).toBe('random error text');
  });

});

describe('formatBulkError edge cases', () => {
  it('handles JSON string that parses to a non-object', () => {
    // JSON.parse('"hello"') === 'hello', not an object with error field
    const result = formatBulkError('"hello"');
    expect(result).toBe('"hello"');
  });

  it('handles JSON string that parses to an object without error field', () => {
    const result = formatBulkError('{"foo": "bar"}');
    expect(result).toBe('{"foo": "bar"}');
  });

  it('handles row = 0 (falsy but defined)', () => {
    const result = formatBulkError({ row: 0, error: 'test error' });
    expect(result).toBe('Строка 0: test error');
  });
});

// --------------- formatErrorMessage ---------------

describe('formatErrorMessage', () => {
  it('returns NETWORK_ERROR for ERR_NETWORK axios code', () => {
    const error = makeAxiosError({ axiosCode: 'ERR_NETWORK', hasResponse: false });
    expect(formatErrorMessage(error)).toBe('Ошибка сети. Проверьте подключение');
  });

  it('returns NETWORK_ERROR when response is absent', () => {
    const error = makeAxiosError({ hasResponse: false });
    expect(formatErrorMessage(error)).toBe('Ошибка сети. Проверьте подключение');
  });

  it('returns base message from ERROR_MESSAGES for known error code', () => {
    const error = makeAxiosError({
      status: 401,
      errorCode: 'INVALID_CREDENTIALS',
      errorMessage: '',
    });
    expect(formatErrorMessage(error)).toBe('Неверные учетные данные');
  });

  it('appends translated server message when it differs from base', () => {
    const error = makeAxiosError({
      status: 401,
      errorCode: 'INVALID_CREDENTIALS',
      errorMessage: 'Invalid username or password',
    });
    expect(formatErrorMessage(error)).toBe(
      'Неверные учетные данные: Неверное имя пользователя или пароль',
    );
  });

  it('does NOT append server message when translation equals base message', () => {
    const error = makeAxiosError({
      status: 404,
      errorCode: 'PASS_NOT_FOUND',
      errorMessage: 'Pass not found',
    });
    // Both resolve to 'Пропуск не найден'
    expect(formatErrorMessage(error)).toBe('Пропуск не найден');
  });

  it('uses UNKNOWN_ERROR for unrecognised error code', () => {
    const error = makeAxiosError({
      status: 500,
      errorCode: 'SOMETHING_WEIRD',
      errorMessage: '',
    });
    expect(formatErrorMessage(error)).toContain('Произошла неизвестная ошибка');
  });

  it('falls back to UNKNOWN_ERROR when error code is absent', () => {
    const error = makeAxiosError({
      status: 500,
      errorCode: '',
      errorMessage: '',
    });
    expect(formatErrorMessage(error)).toContain('Произошла неизвестная ошибка');
  });

  it('appends untranslated server message as-is', () => {
    const error = makeAxiosError({
      status: 400,
      errorCode: 'INVALID_REQUEST',
      errorMessage: 'some exotic detail',
    });
    expect(formatErrorMessage(error)).toBe('Некорректный запрос: some exotic detail');
  });
});

describe('formatErrorMessage with DEV suffix', () => {
  it('does not add HTTP suffix in production (default vitest define)', () => {
    const error = makeAxiosError({
      status: 403,
      errorCode: 'FORBIDDEN',
      errorMessage: '',
    });
    const msg = formatErrorMessage(error);
    expect(msg).not.toContain('(HTTP');
  });
});
