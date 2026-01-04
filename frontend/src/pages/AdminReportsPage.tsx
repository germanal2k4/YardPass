import {
  Container,
  Paper,
  Typography,
  Alert,
  Box,
} from '@mui/material';
import { Layout } from '@/shared/ui/Layout';
import WarningIcon from '@mui/icons-material/Warning';

// NOTE: Backend currently does NOT have endpoints for reports/statistics
// Required endpoints that need to be added to backend:
// - GET /api/v1/scan-events?from=...&to=...&limit=...&offset=...
// - GET /api/v1/reports/statistics?from=...&to=...
// - GET /api/v1/reports/export?format=xlsx&from=...&to=...
// - GET /api/v1/parking/occupancy
// - GET /api/v1/parking/vehicles

export function AdminReportsPage() {
  return (
    <Layout title="Отчеты и статистика">
      <Container maxWidth="lg" sx={{ py: 4 }}>
        <Paper elevation={2} sx={{ p: 4 }}>
          <Box sx={{ display: 'flex', alignItems: 'flex-start', mb: 3 }}>
            <WarningIcon sx={{ fontSize: 40, color: 'warning.main', mr: 2 }} />
            <Box>
              <Typography variant="h5" gutterBottom>
                Функция в разработке
              </Typography>
              <Typography variant="body1" color="text.secondary">
                Для реализации раздела отчетов необходимо добавить следующие endpoints в backend API
              </Typography>
            </Box>
          </Box>

          <Alert severity="info" sx={{ mb: 2 }}>
            <Typography variant="subtitle2" gutterBottom>
              Требуемые API endpoints:
            </Typography>
            <Box component="ul" sx={{ mt: 1, pl: 2 }}>
              <li>
                <code>GET /api/v1/scan-events</code> — журнал событий сканирования с фильтрами
                <br />
                <Typography variant="caption" color="text.secondary">
                  Параметры: from, to, result, guard_user_id, limit, offset
                </Typography>
              </li>
              <li style={{ marginTop: 8 }}>
                <code>GET /api/v1/reports/statistics</code> — общая статистика
                <br />
                <Typography variant="caption" color="text.secondary">
                  Параметры: from, to (период для выборки)
                </Typography>
              </li>
              <li style={{ marginTop: 8 }}>
                <code>GET /api/v1/reports/export</code> — экспорт в Excel
                <br />
                <Typography variant="caption" color="text.secondary">
                  Параметры: format=xlsx, from, to
                </Typography>
              </li>
              <li style={{ marginTop: 8 }}>
                <code>GET /api/v1/parking/occupancy</code> — загруженность парковки
                <br />
                <Typography variant="caption" color="text.secondary">
                  Возвращает: текущее количество занятых мест, общее количество мест
                </Typography>
              </li>
              <li style={{ marginTop: 8 }}>
                <code>GET /api/v1/parking/vehicles</code> — список автомобилей на парковке
                <br />
                <Typography variant="caption" color="text.secondary">
                  Параметры: limit, offset
                </Typography>
              </li>
            </Box>
          </Alert>

          <Alert severity="warning">
            <Typography variant="body2">
              После добавления указанных endpoints в backend, данный раздел будет дополнен следующим функционалом:
            </Typography>
            <Box component="ul" sx={{ mt: 1, pl: 2 }}>
              <li>Просмотр журнала всех событий сканирования</li>
              <li>Фильтрация по дате, результату проверки, охраннику</li>
              <li>Статистика: количество входов/выходов, активность по квартирам</li>
              <li>Информация о нарушениях (недействительные попытки прохода)</li>
              <li>Экспорт отчетов в формате Excel</li>
              <li>Информация о парковке: занятость, список автомобилей</li>
            </Box>
          </Alert>
        </Paper>
      </Container>
    </Layout>
  );
}

