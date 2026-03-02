import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { ThemeProvider } from '@mui/material/styles';
import { theme } from '@/shared/ui/theme';
import { PassDetailsCard } from '../PassDetailsCard';
import type { ValidatePassResponse } from '@/shared/types/api';

function renderCard(result: ValidatePassResponse) {
  return render(
    <ThemeProvider theme={theme}>
      <PassDetailsCard result={result} />
    </ThemeProvider>,
  );
}

describe('PassDetailsCard', () => {
  it('renders valid pass with details', () => {
    renderCard({
      valid: true,
      car_plate: 'А123ВС777',
      apartment: '101',
      valid_to: '2026-07-15T12:00:00Z',
    });

    expect(screen.getByText('Пропуск действителен')).toBeInTheDocument();
    expect(screen.getByText('А123ВС777')).toBeInTheDocument();
    expect(screen.getByText('101')).toBeInTheDocument();
    expect(screen.getByText(/2026/)).toBeInTheDocument();
    expect(screen.getByText('Действителен до')).toBeInTheDocument();
  });

  it('renders invalid pass with reason from ERROR_MESSAGES', () => {
    renderCard({
      valid: false,
      reason: 'PASS_EXPIRED',
    });

    expect(screen.getByText('Пропуск недействителен')).toBeInTheDocument();
    expect(screen.getByText('Срок действия пропуска истек')).toBeInTheDocument();
  });

  it('renders invalid pass with translated reason', () => {
    renderCard({
      valid: false,
      reason: 'PASS_REVOKED' as any,
    });

    expect(screen.getByText('Пропуск недействителен')).toBeInTheDocument();
    expect(screen.getByText('Пропуск отозван')).toBeInTheDocument();
  });

  it('does not show details section for invalid pass', () => {
    renderCard({
      valid: false,
      reason: 'PASS_NOT_FOUND',
    });

    expect(screen.queryByText('Номер автомобиля')).not.toBeInTheDocument();
    expect(screen.queryByText('Квартира')).not.toBeInTheDocument();
  });

  it('renders valid pass without valid_to gracefully', () => {
    renderCard({
      valid: true,
      car_plate: 'В456ОР50',
      apartment: '202',
    });

    expect(screen.getByText('Пропуск действителен')).toBeInTheDocument();
    expect(screen.getByText('В456ОР50')).toBeInTheDocument();
    expect(screen.queryByText('Действителен до')).not.toBeInTheDocument();
  });
});
