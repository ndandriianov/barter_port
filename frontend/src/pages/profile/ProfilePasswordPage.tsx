import { useState } from "react";
import { Alert, Box, Button, Card, CardContent, Stack, TextField, Typography } from "@mui/material";
import { Link as RouterLink } from "react-router-dom";
import authApi from "@/features/auth/api/authApi.ts";
import usersApi from "@/features/users/api/usersApi.ts";
import { getErrorMessage } from "@/shared/utils/getErrorMessage.ts";
import { getStatusCode } from "@/shared/utils/getStatusCode.ts";
import { appRoutes } from "@/shared/config/appRoutes.ts";
import ProfileSectionShell from "@/widgets/profile/ProfileSectionShell.tsx";

const MIN_PASSWORD_LENGTH = 6;

function ProfilePasswordPage() {
  const { data: me } = usersApi.useGetCurrentUserQuery();
  const [changePassword, { isLoading: isChangingPassword, error: changePasswordError }] =
    authApi.useChangePasswordMutation();
  const [oldPassword, setOldPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmNewPassword, setConfirmNewPassword] = useState("");
  const [changePasswordSuccess, setChangePasswordSuccess] = useState<string | null>(null);
  const [changePasswordValidationError, setChangePasswordValidationError] = useState<string | null>(null);

  const isPasswordFormDirty = oldPassword !== "" || newPassword !== "" || confirmNewPassword !== "";

  const handleChangePassword = async () => {
    if (!me?.email) {
      setChangePasswordValidationError("Не удалось определить email пользователя.");
      return;
    }

    if (!oldPassword.trim()) {
      setChangePasswordValidationError("Введите текущий пароль.");
      return;
    }

    if (newPassword.length < MIN_PASSWORD_LENGTH) {
      setChangePasswordValidationError(`Новый пароль должен быть не короче ${MIN_PASSWORD_LENGTH} символов.`);
      return;
    }

    if (newPassword !== confirmNewPassword) {
      setChangePasswordValidationError("Подтверждение пароля не совпадает.");
      return;
    }

    setChangePasswordValidationError(null);
    setChangePasswordSuccess(null);

    try {
      await changePassword({
        oldEmail: me.email,
        oldPassword,
        newPassword,
      }).unwrap();

      setOldPassword("");
      setNewPassword("");
      setConfirmNewPassword("");
      setChangePasswordSuccess("Пароль обновлён.");
    } catch {
      // Error is rendered below.
    }
  };

  return (
    <ProfileSectionShell
      title="Смена пароля"
      description="Пароль вынесен из общей формы личных данных в отдельный account-flow."
      actions={
        <Button component={RouterLink} to={appRoutes.profile.account} variant="outlined">
          Назад к аккаунту
        </Button>
      }
    >
      <Box maxWidth={720}>
        <Card variant="outlined">
          <CardContent>
            <Stack spacing={2.5}>
              <Typography variant="h6" fontWeight={800}>
                Обновить пароль
              </Typography>

              <TextField
                label="Текущий пароль"
                type="password"
                value={oldPassword}
                onChange={(event) => {
                  setOldPassword(event.target.value);
                  setChangePasswordValidationError(null);
                  setChangePasswordSuccess(null);
                }}
                fullWidth
              />
              <TextField
                label="Новый пароль"
                type="password"
                value={newPassword}
                onChange={(event) => {
                  setNewPassword(event.target.value);
                  setChangePasswordValidationError(null);
                  setChangePasswordSuccess(null);
                }}
                helperText={`Минимум ${MIN_PASSWORD_LENGTH} символов`}
                fullWidth
              />
              <TextField
                label="Подтвердите новый пароль"
                type="password"
                value={confirmNewPassword}
                onChange={(event) => {
                  setConfirmNewPassword(event.target.value);
                  setChangePasswordValidationError(null);
                  setChangePasswordSuccess(null);
                }}
                fullWidth
              />

              {changePasswordValidationError && (
                <Alert severity="error">
                  {changePasswordValidationError}
                </Alert>
              )}

              {changePasswordError && !changePasswordValidationError && (
                <Alert severity="error">
                  {getStatusCode(changePasswordError) === 403
                    ? "Текущий пароль указан неверно."
                    : getStatusCode(changePasswordError) === 400
                      ? getErrorMessage(changePasswordError) ?? "Новый пароль не прошёл валидацию."
                      : getErrorMessage(changePasswordError) ?? "Не удалось изменить пароль."}
                </Alert>
              )}

              {changePasswordSuccess && (
                <Alert severity="success">
                  {changePasswordSuccess}
                </Alert>
              )}

              <Box display="flex" gap={2} flexWrap="wrap">
                <Button
                  variant="contained"
                  onClick={handleChangePassword}
                  disabled={!isPasswordFormDirty || isChangingPassword}
                >
                  Сохранить пароль
                </Button>
                <Button component={RouterLink} to={appRoutes.profile.account} variant="outlined">
                  Вернуться
                </Button>
              </Box>
            </Stack>
          </CardContent>
        </Card>
      </Box>
    </ProfileSectionShell>
  );
}

export default ProfilePasswordPage;
