import { useState } from "react";
import { useLocation, useNavigate, Link as RouterLink } from "react-router-dom";
import {
  Alert,
  Box,
  Button,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Link,
  TextField,
  Typography,
} from "@mui/material";
import authApi from "@/features/auth/api/authApi";
import { getErrorMessage } from "@/shared/utils/getErrorMessage.ts";
import { getStatusCode } from "@/shared/utils/getStatusCode.ts";

type LoginLocationState = {
  from?: {
    pathname?: string;
    search?: string;
    hash?: string;
  };
  reason?: string;
};

function LoginForm() {
  const [login, { isLoading, error }] = authApi.useLoginMutation();
  const [requestPasswordReset, { isLoading: isRequestingReset, error: requestPasswordResetError }] =
    authApi.useRequestPasswordResetMutation();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [resetEmail, setResetEmail] = useState("");
  const [isResetDialogOpen, setIsResetDialogOpen] = useState(false);
  const [resetRequestSuccess, setResetRequestSuccess] = useState(false);
  const location = useLocation();
  const navigate = useNavigate();
  const locationState = location.state as LoginLocationState | null;
  const redirectTarget = locationState?.from?.pathname
    ? `${locationState.from.pathname}${locationState.from.search ?? ""}${locationState.from.hash ?? ""}`
    : "/";

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      await login({ email, password }).unwrap();
      navigate(redirectTarget, { replace: true });
    } catch {
      // Error state is rendered below.
    }
  };

  const openResetDialog = () => {
    setResetEmail(email);
    setResetRequestSuccess(false);
    setIsResetDialogOpen(true);
  };

  const closeResetDialog = () => {
    setIsResetDialogOpen(false);
    setResetRequestSuccess(false);
  };

  const submitPasswordReset = async () => {
    try {
      await requestPasswordReset({ email: resetEmail }).unwrap();
      setResetRequestSuccess(true);
    } catch {
      // Error state is rendered below.
    }
  };

  return (
    <>
      <Box component="form" onSubmit={submit} noValidate>
        <Typography variant="h5" fontWeight={700} mb={3} textAlign="center">
          Вход
        </Typography>

        {locationState?.reason && (
          <Alert severity="warning" sx={{ mb: 2 }}>
            {locationState.reason}
          </Alert>
        )}

        <TextField
          label="Email"
          type="email"
          fullWidth
          required
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          sx={{ mb: 2 }}
        />
        <TextField
          label="Пароль"
          type="password"
          fullWidth
          required
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          sx={{ mb: 1 }}
        />

        <Box mb={3} textAlign="right">
          <Button variant="text" size="small" onClick={openResetDialog}>
            Забыли пароль?
          </Button>
        </Box>

        {error && (
          <Alert severity="error" sx={{ mb: 2 }}>
            Неверный email или пароль
          </Alert>
        )}

        <Button type="submit" variant="contained" fullWidth size="large" disabled={isLoading}>
          {isLoading ? <CircularProgress size={24} color="inherit" /> : "Войти"}
        </Button>

        <Box mt={2} textAlign="center">
          <Typography variant="body2">
            Нет аккаунта?{" "}
            <Link component={RouterLink} to="/register">
              Зарегистрироваться
            </Link>
          </Typography>
        </Box>
      </Box>

      <Dialog open={isResetDialogOpen} onClose={closeResetDialog} fullWidth maxWidth="sm">
        <DialogTitle>Восстановление пароля</DialogTitle>
        <DialogContent>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 2, mt: 1 }}>
            Укажите email. Мы отправим ссылку для восстановления пароля.
          </Typography>
          <TextField
            label="Email"
            type="email"
            fullWidth
            value={resetEmail}
            onChange={(event) => {
              setResetEmail(event.target.value);
              setResetRequestSuccess(false);
            }}
          />

          {requestPasswordResetError && !resetRequestSuccess && (
            <Alert severity="error" sx={{ mt: 2 }}>
              {getStatusCode(requestPasswordResetError) === 400
                ? getErrorMessage(requestPasswordResetError) ?? "Проверьте email."
                : "Не удалось отправить письмо для восстановления."}
            </Alert>
          )}

          {resetRequestSuccess && (
            <Alert severity="success" sx={{ mt: 2 }}>
              Если аккаунт с таким email существует, письмо со ссылкой для восстановления уже отправлено.
            </Alert>
          )}
        </DialogContent>
        <DialogActions sx={{ px: 3, pb: 3 }}>
          <Button onClick={closeResetDialog}>Закрыть</Button>
          <Button variant="contained" onClick={submitPasswordReset} disabled={!resetEmail.trim() || isRequestingReset}>
            {isRequestingReset ? <CircularProgress size={20} color="inherit" /> : "Отправить ссылку"}
          </Button>
        </DialogActions>
      </Dialog>
    </>
  );
}

export default LoginForm;
