import { useMemo, useRef, useState } from "react";
import { skipToken } from "@reduxjs/toolkit/query";
import { Link as RouterLink, useNavigate } from "react-router-dom";
import {
  Alert,
  Avatar,
  Box,
  Button,
  Card,
  CardContent,
  CircularProgress,
  Dialog,
  DialogContent,
  DialogTitle,
  Divider,
  Drawer,
  List,
  ListItem,
  ListItemAvatar,
  ListItemText,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import PersonOutlineIcon from "@mui/icons-material/PersonOutline";
import usersApi from "@/features/users/api/usersApi";
import authApi from "@/features/auth/api/authApi";
import { useAppDispatch } from "@/hooks/redux";
import { performLogout } from "@/features/auth/model/logoutThunk";
import { imageToAvatarDataUrl } from "@/shared/utils/imageToAvatarDataUrl.ts";
import { getErrorMessage } from "@/shared/utils/getErrorMessage.ts";
import { getStatusCode } from "@/shared/utils/getStatusCode.ts";
import type { User } from "@/features/users/model/types.ts";

const MAX_AVATAR_FILE_SIZE = 5 * 1024 * 1024;
const MIN_PASSWORD_LENGTH = 6;

function ProfilePage() {
  const { data, isLoading, refetch } = usersApi.useGetCurrentUserQuery();
  const [updateCurrentUser, { isLoading: isSaving, error: updateError }] =
    usersApi.useUpdateCurrentUserMutation();
  const [uploadCurrentUserAvatar, { isLoading: isUploadingAvatar, error: uploadAvatarError }] =
    usersApi.useUploadCurrentUserAvatarMutation();
  const [changePassword, { isLoading: isChangingPassword, error: changePasswordError }] =
    authApi.useChangePasswordMutation();
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const [draftName, setDraftName] = useState<string | null>(null);
  const [draftBio, setDraftBio] = useState<string | null>(null);
  const [draftAvatarUrl, setDraftAvatarUrl] = useState<string | null>(null);
  const [draftPhoneNumber, setDraftPhoneNumber] = useState<string | null>(null);
  const [draftAvatarFile, setDraftAvatarFile] = useState<File | null>(null);
  const [avatarError, setAvatarError] = useState<string | null>(null);
  const [isReputationDrawerOpen, setIsReputationDrawerOpen] = useState(false);
  const [subscriptionsDialogOpen, setSubscriptionsDialogOpen] = useState(false);
  const [subscribersDialogOpen, setSubscribersDialogOpen] = useState(false);
  const [isChangePasswordDialogOpen, setIsChangePasswordDialogOpen] = useState(false);
  const [oldPassword, setOldPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmNewPassword, setConfirmNewPassword] = useState("");
  const [changePasswordSuccess, setChangePasswordSuccess] = useState<string | null>(null);
  const [changePasswordValidationError, setChangePasswordValidationError] = useState<string | null>(null);
  const {
    data: reputationEvents,
    isFetching: isReputationEventsLoading,
    error: reputationEventsError,
    refetch: refetchReputationEvents,
  } = usersApi.useGetCurrentUserReputationEventsQuery(
    isReputationDrawerOpen ? undefined : skipToken,
  );
  const {
    data: subscriptions,
    isFetching: isSubscriptionsLoading,
    error: subscriptionsError,
  } = usersApi.useGetSubscriptionsQuery();
  const {
    data: subscribers,
    isFetching: isSubscribersLoading,
    error: subscribersError,
  } = usersApi.useGetSubscribersQuery();
  const fileInputRef = useRef<HTMLInputElement | null>(null);

  const currentName = draftName ?? (data?.name ?? "");
  const currentBio = draftBio ?? (data?.bio ?? "");
  const currentAvatarUrl = draftAvatarUrl ?? (data?.avatarUrl ?? "");
  const currentPhoneNumber = draftPhoneNumber ?? (data?.phoneNumber ?? "");

  const hasChanges = useMemo(() => {
    if (!data) {
      return false;
    }

    return (
      currentName !== (data.name ?? "") ||
      currentBio !== (data.bio ?? "") ||
      currentAvatarUrl !== (data.avatarUrl ?? "") ||
      currentPhoneNumber !== (data.phoneNumber ?? "")
    );
  }, [currentAvatarUrl, currentBio, currentName, currentPhoneNumber, data]);

  const normalizedAvatarUrl = currentAvatarUrl.trim();
  const avatarPreviewUrl = normalizedAvatarUrl || undefined;
  const hasAvatarPreview = Boolean(avatarPreviewUrl);
  const isSubmitting = isSaving || isUploadingAvatar;
  const isPasswordFormDirty = oldPassword !== "" || newPassword !== "" || confirmNewPassword !== "";

  const formatReputationDelta = (delta: number) => (delta > 0 ? `+${delta}` : `${delta}`);

  const formatSourceType = (sourceType: string) => {
    switch (sourceType) {
      case "deals.offer_report.penalty":
        return "штраф по жалобе на предложение";
      case "deals.deal_failure.responsible":
        return "штраф за провал сделки";
      case "deals.deal_completion.reward":
        return "завершение сделки";
      case "deals.review_creation.reward":
        return "оставленный отзыв";
      default:
        return sourceType;
    }
  };

  const handleLogout = async () => {
    await dispatch(performLogout());
    navigate("/login");
  };

  const handleSave = async () => {
    try {
      let nextAvatarUrl = normalizedAvatarUrl;

      if (draftAvatarFile) {
        const formData = new FormData();
        formData.append("file", draftAvatarFile);
        const uploadResponse = await uploadCurrentUserAvatar(formData).unwrap();
        nextAvatarUrl = uploadResponse.avatarUrl.trim();
      }

      await updateCurrentUser({
        name: currentName.trim(),
        bio: currentBio.trim(),
        avatarUrl: nextAvatarUrl,
        phoneNumber: currentPhoneNumber.trim(),
      }).unwrap();
      // Drop local draft and rely on fresh server state after mutation invalidation.
      setDraftName(null);
      setDraftBio(null);
      setDraftAvatarUrl(null);
      setDraftPhoneNumber(null);
      setDraftAvatarFile(null);
    } catch {
      // Error state is already exposed by RTK Query and rendered in UI.
    }
  };

  const handleClear = () => {
    setDraftName("");
    setDraftBio("");
    setDraftAvatarUrl("");
    setDraftPhoneNumber("");
    setDraftAvatarFile(null);
    setAvatarError(null);
  };

  const resetChangePasswordState = () => {
    setOldPassword("");
    setNewPassword("");
    setConfirmNewPassword("");
    setChangePasswordSuccess(null);
    setChangePasswordValidationError(null);
  };

  const handleOpenChangePasswordDialog = () => {
    resetChangePasswordState();
    setIsChangePasswordDialogOpen(true);
  };

  const handleCloseChangePasswordDialog = () => {
    resetChangePasswordState();
    setIsChangePasswordDialogOpen(false);
  };

  const handleChangePassword = async () => {
    if (!data?.email) {
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
        oldEmail: data.email,
        oldPassword,
        newPassword,
      }).unwrap();

      setOldPassword("");
      setNewPassword("");
      setConfirmNewPassword("");
      setChangePasswordSuccess("Пароль обновлен.");
    } catch {
      // RTK Query exposes error state that is rendered below.
    }
  };

  const handleAvatarButtonClick = () => {
    fileInputRef.current?.click();
  };

  const handleAvatarFileChange = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    event.target.value = "";

    if (!file) {
      return;
    }

    if (!file.type.startsWith("image/")) {
      setAvatarError("Нужно выбрать изображение.");
      return;
    }

    if (file.size > MAX_AVATAR_FILE_SIZE) {
      setAvatarError("Размер файла не должен превышать 5 МБ.");
      return;
    }

    try {
      const nextAvatarUrl = await imageToAvatarDataUrl(file);
      setDraftAvatarUrl(nextAvatarUrl);
      setDraftAvatarFile(file);
      setAvatarError(null);
    } catch (error) {
      setAvatarError(error instanceof Error ? error.message : "Не удалось обработать изображение.");
    }
  };

  const handleClearAvatar = () => {
    setDraftAvatarUrl("");
    setDraftAvatarFile(null);
    setAvatarError(null);
  };

  const handleOpenSubscriptionsDialog = () => {
    setSubscriptionsDialogOpen(true);
  };

  const handleOpenSubscribersDialog = () => {
    setSubscribersDialogOpen(true);
  };

  const handleCloseSubscriptionsDialog = () => {
    setSubscriptionsDialogOpen(false);
  };

  const handleCloseSubscribersDialog = () => {
    setSubscribersDialogOpen(false);
  };

  const handleCloseUserDialogs = () => {
    handleCloseSubscriptionsDialog();
    handleCloseSubscribersDialog();
  };

  const renderUserListItem = (user: User) => (
    <ListItem
      key={user.id}
      component={RouterLink}
      to={`/users/${user.id}`}
      onClick={handleCloseUserDialogs}
      sx={{ textDecoration: "none", color: "inherit" }}
    >
      <ListItemAvatar>
        <Avatar src={user.avatarUrl?.trim() || undefined} sx={{ width: 40, height: 40 }}>
          {!user.avatarUrl?.trim() && <PersonOutlineIcon fontSize="small" />}
        </Avatar>
      </ListItemAvatar>
      <ListItemText
        primary={user.name?.trim() || "Имя не указано"}
        secondary={`ID: ${user.id}`}
      />
    </ListItem>
  );

  if (isLoading) {
    return (
      <Box display="flex" justifyContent="center" py={6}>
        <CircularProgress />
      </Box>
    );
  }

  if (!data) {
    return <Alert severity="warning">Вы не авторизованы</Alert>;
  }

  const subscriptionsCount = subscriptions?.length ?? 0;
  const subscribersCount = subscribers?.length ?? 0;

  return (
    <Box maxWidth={560} mx="auto">
      <Typography variant="h4" fontWeight={700} mb={3}>
        Профиль
      </Typography>

      <Card variant="outlined">
        <CardContent>
          <Box display="flex" alignItems="center" gap={2} mb={3}>
            <input
              ref={fileInputRef}
              type="file"
              accept="image/*"
              hidden
              onChange={handleAvatarFileChange}
            />
            <Avatar
              src={avatarPreviewUrl}
              alt={currentName.trim() || "Пользователь"}
              sx={{ width: 72, height: 72, bgcolor: "action.selected" }}
            >
              {!hasAvatarPreview && <PersonOutlineIcon fontSize="large" color="action" />}
            </Avatar>
            <Box>
              <Typography variant="caption" color="text.secondary">
                ID
              </Typography>
              <Typography variant="body2" fontFamily="monospace" fontWeight={500}>
                {data.id}
              </Typography>
              <Box display="flex" gap={1} flexWrap="wrap" mt={1.5}>
                <Button variant="outlined" size="small" onClick={handleAvatarButtonClick}>
                  Загрузить аватар
                </Button>
                <Button variant="text" size="small" color="inherit" onClick={handleClearAvatar}>
                  Удалить аватар
                </Button>
              </Box>
            </Box>
          </Box>

          <Stack spacing={1.5} mb={3}>
            <Box>
              <Typography variant="caption" color="text.secondary">
                Email
              </Typography>
              <Typography variant="body2">{data.email}</Typography>
            </Box>
            <Box>
              <Typography variant="caption" color="text.secondary">
                Телефон
              </Typography>
              <Typography variant="body2">{data.phoneNumber?.trim() || "Не указан"}</Typography>
            </Box>
            <Box>
              <Typography variant="caption" color="text.secondary">
                Зарегистрирован
              </Typography>
              <Typography variant="body2">
                {new Date(data.createdAt).toLocaleString("ru-RU")}
              </Typography>
            </Box>
            <Box>
              <Typography variant="caption" color="text.secondary">
                Рейтинг
              </Typography>
              <Typography variant="h6" fontWeight={700}>
                {data.reputationPoints}
              </Typography>
            </Box>
            <Box>
              <Typography variant="caption" color="text.secondary">
                Подписки
              </Typography>
              <Button
                variant="text"
                onClick={handleOpenSubscriptionsDialog}
                sx={{ display: "block", px: 0, minWidth: 0, fontWeight: 700 }}
              >
                {isSubscriptionsLoading ? "Загрузка..." : subscriptionsCount}
              </Button>
            </Box>
            <Box>
              <Typography variant="caption" color="text.secondary">
                Подписчики
              </Typography>
              <Button
                variant="text"
                onClick={handleOpenSubscribersDialog}
                sx={{ display: "block", px: 0, minWidth: 0, fontWeight: 700 }}
              >
                {isSubscribersLoading ? "Загрузка..." : subscribersCount}
              </Button>
            </Box>
          </Stack>

          <Divider sx={{ mb: 3 }} />

          <Stack spacing={2} mb={3}>
            <TextField
              label="Имя"
              value={currentName}
              onChange={(event) => setDraftName(event.target.value)}
              placeholder="Введите имя"
              fullWidth
            />
            <TextField
              label="Bio"
              value={currentBio}
              onChange={(event) => setDraftBio(event.target.value)}
              placeholder="Расскажите о себе"
              fullWidth
              multiline
              minRows={3}
              helperText="Чтобы удалить имя, bio или телефон, очистите поле и сохраните"
            />
            <TextField
              label="Телефон"
              type="tel"
              value={currentPhoneNumber}
              onChange={(event) => setDraftPhoneNumber(event.target.value)}
              placeholder="+79991234567"
              fullWidth
            />
          </Stack>

          {avatarError && (
            <Alert severity="error" sx={{ mb: 2 }}>
              {avatarError}
            </Alert>
          )}

          {(normalizedAvatarUrl || draftAvatarUrl !== null) && !avatarError && (
            <Alert severity="info" sx={{ mb: 2 }}>
              Аватар выбран локально и будет сохранен после нажатия «Сохранить».
            </Alert>
          )}

          {uploadAvatarError && (
            <Alert severity="error" sx={{ mb: 2 }}>
              Не удалось загрузить аватар
            </Alert>
          )}

          {updateError && (
            <Alert severity="error" sx={{ mb: 2 }}>
              Не удалось обновить профиль
            </Alert>
          )}

          <Box display="flex" gap={2} flexWrap="wrap">
            <Button
              variant="contained"
              onClick={handleSave}
              disabled={!hasChanges || isSubmitting}
            >
              Сохранить
            </Button>
            <Button variant="outlined" onClick={handleClear} disabled={isSubmitting}>
              Очистить поля
            </Button>
            <Button variant="contained" color="error" onClick={handleLogout}>
              Выйти
            </Button>
            <Button variant="outlined" onClick={() => refetch()} disabled={isSubmitting}>
              Обновить
            </Button>
            <Button
              variant="outlined"
              onClick={handleOpenChangePasswordDialog}
            >
              Сменить пароль
            </Button>
            <Button variant="outlined" onClick={() => setIsReputationDrawerOpen(true)}>
              История рейтинга
            </Button>
            <Button component={RouterLink} to="/statistics" variant="outlined">
              Моя статистика
            </Button>
            <Button component={RouterLink} to="/offer-reports/mine" variant="outlined" color="warning">
              Жалобы на меня
            </Button>
            <Button variant="outlined" onClick={handleOpenSubscriptionsDialog}>
              Мои подписки
            </Button>
            <Button variant="outlined" onClick={handleOpenSubscribersDialog}>
              Мои подписчики
            </Button>
          </Box>

          <Divider sx={{ my: 3 }} />

          <Box display="flex" gap={2} flexWrap="wrap">
            <Button component={RouterLink} to="/reviews?tab=available" variant="outlined">
              Оставить отзыв
            </Button>
            <Button component={RouterLink} to="/reviews?tab=mine" variant="outlined">
              Мои отзывы
            </Button>
            <Button component={RouterLink} to="/reviews?tab=about-me" variant="outlined">
              Отзывы обо мне
            </Button>
          </Box>
        </CardContent>
      </Card>

      <Drawer
        anchor="right"
        open={isReputationDrawerOpen}
        onClose={() => setIsReputationDrawerOpen(false)}
      >
        <Box sx={{ width: { xs: 320, sm: 420 }, p: 3 }}>
          <Typography variant="h5" fontWeight={700} mb={1}>
            История рейтинга
          </Typography>
          <Typography variant="body2" color="text.secondary" mb={3}>
            Текущее значение рейтинга: <strong>{data.reputationPoints}</strong>
          </Typography>

          {isReputationEventsLoading && (
            <Box display="flex" justifyContent="center" py={3}>
              <CircularProgress size={28} />
            </Box>
          )}

          {!isReputationEventsLoading && reputationEventsError && (
            <Alert
              severity="error"
              action={
                <Button color="inherit" size="small" onClick={() => refetchReputationEvents()}>
                  Повторить
                </Button>
              }
            >
              Не удалось загрузить историю рейтинга.
            </Alert>
          )}

          {!isReputationEventsLoading && !reputationEventsError && (!reputationEvents || reputationEvents.length === 0) && (
            <Alert severity="info">История изменения рейтинга пока пуста.</Alert>
          )}

          {!isReputationEventsLoading && !reputationEventsError && reputationEvents && reputationEvents.length > 0 && (
            <Stack spacing={1.5}>
              {reputationEvents.map((event) => (
                <Card key={event.id} variant="outlined">
                  <CardContent sx={{ p: 2, "&:last-child": { pb: 2 } }}>
                    <Stack spacing={0.5}>
                      <Typography variant="body2" color="text.secondary">
                        {new Date(event.createdAt).toLocaleString("ru-RU")}
                      </Typography>
                      <Typography variant="subtitle2" fontWeight={700}>
                        {formatSourceType(event.sourceType)}
                      </Typography>
                      <Typography
                        variant="h6"
                        color={event.delta >= 0 ? "success.main" : "error.main"}
                        fontWeight={700}
                      >
                        {formatReputationDelta(event.delta)}
                      </Typography>
                      {event.comment && <Typography variant="body2">{event.comment}</Typography>}
                    </Stack>
                  </CardContent>
                </Card>
              ))}
            </Stack>
          )}
        </Box>
      </Drawer>

      <Dialog
        open={isChangePasswordDialogOpen}
        onClose={handleCloseChangePasswordDialog}
        maxWidth="sm"
        fullWidth
      >
        <DialogTitle>Смена пароля</DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ pt: 1 }}>
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
                    ? getErrorMessage(changePasswordError) ?? "Новый пароль не прошел валидацию."
                    : getErrorMessage(changePasswordError) ?? "Не удалось изменить пароль."}
              </Alert>
            )}

            {changePasswordSuccess && (
              <Alert severity="success">
                {changePasswordSuccess}
              </Alert>
            )}

            <Box display="flex" gap={2} justifyContent="flex-end" flexWrap="wrap">
              <Button variant="text" onClick={handleCloseChangePasswordDialog}>
                Закрыть
              </Button>
              <Button
                variant="contained"
                onClick={handleChangePassword}
                disabled={!isPasswordFormDirty || isChangingPassword}
              >
                Сохранить пароль
              </Button>
            </Box>
          </Stack>
        </DialogContent>
      </Dialog>

      <Dialog
        open={subscriptionsDialogOpen}
        onClose={handleCloseSubscriptionsDialog}
        maxWidth="sm"
        fullWidth
      >
        <DialogTitle>Мои подписки</DialogTitle>
        <DialogContent>
          {isSubscriptionsLoading ? (
            <Box display="flex" justifyContent="center" py={3}>
              <CircularProgress size={28} />
            </Box>
          ) : subscriptionsError ? (
            <Alert severity="error">Не удалось загрузить список подписок.</Alert>
          ) : !subscriptions || subscriptions.length === 0 ? (
            <Alert severity="info">Вы пока ни на кого не подписаны.</Alert>
          ) : (
            <List>
              {subscriptions.map((user) => renderUserListItem(user))}
            </List>
          )}
        </DialogContent>
      </Dialog>

      <Dialog
        open={subscribersDialogOpen}
        onClose={handleCloseSubscribersDialog}
        maxWidth="sm"
        fullWidth
      >
        <DialogTitle>Мои подписчики</DialogTitle>
        <DialogContent>
          {isSubscribersLoading ? (
            <Box display="flex" justifyContent="center" py={3}>
              <CircularProgress size={28} />
            </Box>
          ) : subscribersError ? (
            <Alert severity="error">Не удалось загрузить список подписчиков.</Alert>
          ) : !subscribers || subscribers.length === 0 ? (
            <Alert severity="info">У вас пока нет подписчиков.</Alert>
          ) : (
            <List>
              {subscribers.map((user) => renderUserListItem(user))}
            </List>
          )}
        </DialogContent>
      </Dialog>
    </Box>
  );
}

export default ProfilePage;
