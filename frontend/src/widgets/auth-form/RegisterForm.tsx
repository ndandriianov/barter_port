import { useState } from "react";
import { Link as RouterLink } from "react-router-dom";
import { Alert, Box, Button, CircularProgress, Link, TextField, Typography } from "@mui/material";
import authApi from "@/features/auth/api/authApi";

function RegisterForm() {
  const [register, { isLoading, error, isSuccess }] = authApi.useRegisterMutation();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    await register({ email, password });
  };

  return (
    <Box component="form" onSubmit={submit} noValidate>
      <Typography variant="h5" fontWeight={700} mb={3} textAlign="center">
        Регистрация
      </Typography>

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
        sx={{ mb: 3 }}
      />

      {error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          Ошибка при регистрации
        </Alert>
      )}

      {isSuccess && (
        <Alert severity="success" sx={{ mb: 2 }}>
          Аккаунт создан. Проверьте почту для подтверждения.
        </Alert>
      )}

      <Button type="submit" variant="contained" fullWidth size="large" disabled={isLoading}>
        {isLoading ? <CircularProgress size={24} color="inherit" /> : "Зарегистрироваться"}
      </Button>

      <Box mt={2} textAlign="center">
        <Typography variant="body2">
          Уже есть аккаунт?{" "}
          <Link component={RouterLink} to="/login">
            Войти
          </Link>
        </Typography>
      </Box>
    </Box>
  );
}

export default RegisterForm;
