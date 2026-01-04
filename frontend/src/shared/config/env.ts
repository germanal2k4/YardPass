export const config = {
  apiBaseUrl: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080',
  defaultBuildingId: Number(import.meta.env.VITE_DEFAULT_BUILDING_ID) || 1,
} as const;

