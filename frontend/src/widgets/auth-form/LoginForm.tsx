import { useState } from "react";
import { useNavigate, Link as RouterLink } from "react-router-dom";
import { Alert, Box, Button, CircularProgress, Link, TextField, Typography } from "@mui/material";
import authApi from "@/features/auth/api/authApi";

function LoginForm() {
  const [login, { isLoading, error }] = authApi.useLoginMutation();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const navigate = useNavigate();

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    await login({ email, password });
    navigate("/");
  };

  return (
    <Box component="form" onSubmit={submit} noValidate>
      <Typography variant="h5" fontWeight={700} mb={3} textAlign="center">
        Вход
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
  );
}

export default LoginForm;
