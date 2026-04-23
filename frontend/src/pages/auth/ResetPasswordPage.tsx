import { useState } from "react";
import { Link as RouterLink, useSearchParams } from "react-router-dom";
import { Alert, Box, Button, CircularProgress, Link, TextField, Typography } from "@mui/material";
import authApi from "@/features/auth/api/authApi";
import { getErrorMessage } from "@/shared/utils/getErrorMessage.ts";
import { getStatusCode } from "@/shared/utils/getStatusCode.ts";

const MIN_PASSWORD_LENGTH = 6;

function ResetPasswordPage() {
  const [searchParams] = useSearchParams();
  const token = searchParams.get("token") || "";
  const [resetPassword, { isLoading, error }] = authApi.useResetPasswordMutation();
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [validationError, setValidationError] = useState<string | null>(null);
  const [isSuccess, setIsSuccess] = useState(false);

  const submit = async (event: React.FormEvent) => {
    event.preventDefault();

    if (!token) {
      setValidationError("Ссылка для восстановления недействительна.");
      return;
    }

    if (newPassword.length < MIN_PASSWORD_LENGTH) {
      setValidationError(`Пароль должен быть не короче ${MIN_PASSWORD_LENGTH} символов.`);
      return;
    }

    if (newPassword !== confirmPassword) {
      setValidationError("Подтверждение пароля не совпадает.");
      return;
    }

    setValidationError(null);

    try {
      await resetPassword({ token, newPassword }).unwrap();
      setIsSuccess(true);
    } catch {
      // Error state is rendered below.
    }
  };

  return (
    <Box component="form" onSubmit={submit} noValidate>
      <Typography variant="h5" fontWeight={700} mb={3} textAlign="center">
        Восстановление пароля
      </Typography>

      <TextField
        label="Новый пароль"
        type="password"
        fullWidth
        required
        value={newPassword}
        onChange={(event) => {
          setNewPassword(event.target.value);
          setValidationError(null);
        }}
        helperText={`Минимум ${MIN_PASSWORD_LENGTH} символов`}
        sx={{ mb: 2 }}
      />
      <TextField
        label="Подтвердите новый пароль"
        type="password"
        fullWidth
        required
        value={confirmPassword}
        onChange={(event) => {
          setConfirmPassword(event.target.value);
          setValidationError(null);
        }}
        sx={{ mb: 3 }}
      />

      {validationError && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {validationError}
        </Alert>
      )}

      {error && !validationError && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {getStatusCode(error) === 400
            ? getErrorMessage(error) ?? "Ссылка недействительна или устарела."
            : getStatusCode(error) === 404
              ? "Пользователь не найден."
              : "Не удалось обновить пароль."}
        </Alert>
      )}

      {isSuccess && (
        <Alert severity="success" sx={{ mb: 2 }}>
          Пароль обновлен. Теперь можно <Link component={RouterLink} to="/login">войти</Link> с новым паролем.
        </Alert>
      )}

      <Button type="submit" variant="contained" fullWidth size="large" disabled={!token || isLoading || isSuccess}>
        {isLoading ? <CircularProgress size={24} color="inherit" /> : "Сохранить новый пароль"}
      </Button>
    </Box>
  );
}

export default ResetPasswordPage;
