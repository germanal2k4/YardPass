import { useState, FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Container,
  Paper,
  TextField,
  Button,
  Typography,
  Box,
  Alert,
  IconButton,
  Link,
  Divider,
} from '@mui/material';
import ArrowBackIcon from '@mui/icons-material/ArrowBack';
import CreditCardIcon from '@mui/icons-material/CreditCard';
import { APP_ROUTES } from '@/shared/config/constants';
import { authApi } from '@/shared/api/auth';
import { getErrorMessage } from '@/shared/utils/errors';

export function RegistrationPage() {
  const navigate = useNavigate();

  const [buildingName, setBuildingName] = useState('');
  const [email, setEmail] = useState('');
  const [cardNumber, setCardNumber] = useState('');
  const [cardHolder, setCardHolder] = useState('');
  const [expiry, setExpiry] = useState('');
  const [cvv, setCvv] = useState('');
  const [error, setError] = useState('');
  const [success, setSuccess] = useState(false);
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError('');
    setSuccess(false);

    if (!buildingName.trim()) {
      setError('Укажите название здания');
      return;
    }

    if (!email.trim()) {
      setError('Укажите email');
      return;
    }

    if (!cardNumber.trim() || !cardHolder.trim() || !expiry.trim() || !cvv.trim()) {
      setError('Заполните все поля карты');
      return;
    }

    setIsLoading(true);

    try {
      await authApi.purchaseSubscription({
        email: email.trim(),
        building_name: buildingName.trim(),
        card_number: cardNumber.trim(),
        card_holder: cardHolder.trim(),
        expiry: expiry.trim(),
        cvv: cvv.trim(),
      });

      setSuccess(true);
      setError('');

      setTimeout(() => {
        navigate(APP_ROUTES.LOGIN);
      }, 2200);
    } catch (err) {
      setError(getErrorMessage(err));
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Container maxWidth="sm">
      <Box
        sx={{
          minHeight: '100vh',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          py: 4,
          background: 'linear-gradient(135deg, rgba(229, 57, 53, 0.05) 0%, rgba(255, 109, 0, 0.05) 100%)',
        }}
      >
        <Paper 
          elevation={6} 
          sx={{ 
            p: 5, 
            width: '100%', 
            position: 'relative',
            borderRadius: 4,
            background: 'linear-gradient(to bottom, #FFFFFF 0%, #FAFAFA 100%)',
          }}
        >
          <IconButton
            onClick={() => navigate(APP_ROUTES.HOME)}
            sx={{ 
              position: 'absolute', 
              top: 20, 
              left: 20,
              color: '#E53935',
              '&:hover': {
                backgroundColor: 'rgba(229, 57, 53, 0.08)',
              },
            }}
            aria-label="назад"
          >
            <ArrowBackIcon />
          </IconButton>

          <Box sx={{ textAlign: 'center', mb: 4, display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
            <Box
              component="img"
              src="/logo.png"
              alt="YardPass Logo"
              sx={{
                height: { xs: 70, sm: 90 },
                width: 'auto',
                mb: 2,
                display: 'block',
                transition: 'transform 0.3s ease',
                '&:hover': {
                  transform: 'scale(1.05)',
                },
              }}
            />
            <CreditCardIcon
              sx={{ 
                fontSize: 56, 
                color: '#FF6D00',
                mb: 1,
                display: 'block',
                filter: 'drop-shadow(0 2px 4px rgba(255, 109, 0, 0.3))',
              }} 
            />
            <Typography 
              variant="h3" 
              component="h1" 
              gutterBottom
              fontWeight="800"
              sx={{
                background: 'linear-gradient(135deg, #E53935 0%, #FF6D00 100%)',
                backgroundClip: 'text',
                WebkitBackgroundClip: 'text',
                WebkitTextFillColor: 'transparent',
              }}
            >
              Оплата подписки
            </Typography>
            <Typography variant="body1" color="text.secondary" fontWeight="600">
              Тариф: 200 000 ₽ в год
            </Typography>
          </Box>

          <Alert severity="info" sx={{ mb: 3 }}>
            <Typography variant="body2">
              После оплаты на указанный email отправим логины и пароли для аккаунтов администратора и охранника вашего здания.
            </Typography>
          </Alert>

          {success && (
            <Alert severity="success" sx={{ mb: 2 }}>
              Оплата прошла успешно! Данные для входа отправлены на email. Перенаправление на вход...
            </Alert>
          )}

          {error && (
            <Alert severity="error" sx={{ mb: 2 }}>
              {error}
            </Alert>
          )}

          <form onSubmit={handleSubmit}>
            <TextField
              label="Здание"
              type="text"
              fullWidth
              required
              value={buildingName}
              onChange={(e) => setBuildingName(e.target.value)}
              sx={{ mb: 2 }}
              autoFocus
              helperText="Для какого здания подключаем сервис"
              disabled={isLoading || success}
            />

            <TextField
              label="Email для получения доступов"
              type="email"
              fullWidth
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              sx={{ mb: 2 }}
              disabled={isLoading || success}
            />

            <Divider sx={{ mb: 2 }}>
              <Typography variant="caption" color="text.secondary">
                Платежная карта
              </Typography>
            </Divider>

            <TextField
              label="Номер карты"
              type="text"
              fullWidth
              required
              value={cardNumber}
              onChange={(e) => setCardNumber(e.target.value)}
              sx={{ mb: 2 }}
              placeholder="4111 1111 1111 1111"
              disabled={isLoading || success}
              inputProps={{ 'data-testid': 'card-number-input' }}
            />

            <TextField
              label="Владелец карты"
              type="text"
              fullWidth
              required
              value={cardHolder}
              onChange={(e) => setCardHolder(e.target.value)}
              sx={{ mb: 2 }}
              disabled={isLoading || success}
              placeholder="IVAN PETROV"
            />

            <Box sx={{ display: 'flex', gap: 2, mb: 3 }}>
              <TextField
                label="Срок (MM/YY)"
                type="text"
                fullWidth
                required
                value={expiry}
                onChange={(e) => setExpiry(e.target.value)}
                disabled={isLoading || success}
                placeholder="12/30"
              />
              <TextField
                label="CVV"
                type="password"
                fullWidth
                required
                value={cvv}
                onChange={(e) => setCvv(e.target.value)}
                disabled={isLoading || success}
                inputProps={{ maxLength: 4 }}
              />
            </Box>

            <Button
              type="submit"
              variant="contained"
              fullWidth
              size="large"
              disabled={isLoading || success}
              startIcon={<CreditCardIcon />}
              color="primary"
              sx={{
                py: 1.5,
                fontSize: '1.1rem',
                fontWeight: 700,
              }}
            >
              {isLoading ? 'Проводим оплату...' : 'Оплатить 200 000 ₽ / год'}
            </Button>
          </form>

          <Box sx={{ mt: 3, textAlign: 'center' }}>
            <Typography variant="body2" color="text.secondary">
              Уже есть аккаунт?{' '}
              <Link
                component="button"
                variant="body2"
                onClick={() => navigate(APP_ROUTES.LOGIN)}
                sx={{ 
                  cursor: 'pointer',
                  color: '#E53935',
                  fontWeight: 700,
                  '&:hover': {
                    color: '#FF6D00',
                  },
                }}
              >
                Войти
              </Link>
            </Typography>
          </Box>
        </Paper>
      </Box>
    </Container>
  );
}

