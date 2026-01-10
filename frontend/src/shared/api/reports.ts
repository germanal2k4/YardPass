import { apiClient } from './client';
import { API_ENDPOINTS } from '@/shared/config/constants';
import type {
  GetScanEventsRequest,
  GetScanEventsResponse,
  GetStatisticsRequest,
  Statistics,
  ExportReportRequest,
} from '@/shared/types/api';

export const reportsApi = {
  /**
   * Получить журнал событий сканирования
   */
  getScanEvents: async (params?: GetScanEventsRequest): Promise<GetScanEventsResponse> => {
    const queryParams = new URLSearchParams();
    
    if (params?.limit !== undefined) {
      queryParams.append('limit', params.limit.toString());
    }
    if (params?.offset !== undefined) {
      queryParams.append('offset', params.offset.toString());
    }
    if (params?.from) {
      queryParams.append('from', params.from);
    }
    if (params?.to) {
      queryParams.append('to', params.to);
    }
    if (params?.result) {
      queryParams.append('result', params.result);
    }

    const url = `${API_ENDPOINTS.SCAN_EVENTS}${queryParams.toString() ? '?' + queryParams.toString() : ''}`;
    const response = await apiClient.get<GetScanEventsResponse>(url);
    return response.data;
  },

  /**
   * Получить статистику по сканированиям
   */
  getStatistics: async (params?: GetStatisticsRequest): Promise<Statistics> => {
    const queryParams = new URLSearchParams();
    
    if (params?.from) {
      queryParams.append('from', params.from);
    }
    if (params?.to) {
      queryParams.append('to', params.to);
    }

    const url = `${API_ENDPOINTS.STATISTICS}${queryParams.toString() ? '?' + queryParams.toString() : ''}`;
    const response = await apiClient.get<Statistics>(url);
    return response.data;
  },

  /**
   * Экспортировать отчет в Excel
   * @returns Blob с файлом Excel
   */
  exportReport: async (params: ExportReportRequest): Promise<Blob> => {
    const queryParams = new URLSearchParams();
    
    queryParams.append('format', params.format);
    
    if (params?.from) {
      queryParams.append('from', params.from);
    }
    if (params?.to) {
      queryParams.append('to', params.to);
    }

    const url = `${API_ENDPOINTS.EXPORT_REPORT}?${queryParams.toString()}`;
    const response = await apiClient.get(url, {
      responseType: 'blob', // Important for file download
    });
    
    return response.data;
  },
};

