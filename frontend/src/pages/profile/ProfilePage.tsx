import { useMemo, useRef, useState } from "react";
import { Link as RouterLink, useNavigate } from "react-router-dom";
import {
  Alert,
  Avatar,
  Box,
  Button,
  Card,
  CardContent,
  CircularProgress,
  Divider,
  Drawer,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import PersonOutlineIcon from "@mui/icons-material/PersonOutline";
import usersApi from "@/features/users/api/usersApi";
import { useAppDispatch } from "@/hooks/redux";
import { performLogout } from "@/features/auth/model/logoutThunk";
import { imageToAvatarDataUrl } from "@/shared/utils/imageToAvatarDataUrl.ts";

const MAX_AVATAR_FILE_SIZE = 5 * 1024 * 1024;

function ProfilePage() {
  const { data, isLoading, refetch } = usersApi.useGetCurrentUserQuery();
  const [updateCurrentUser, { isLoading: isSaving, error: updateError }] =
    usersApi.useUpdateCurrentUserMutation();
  const [uploadCurrentUserAvatar, { isLoading: isUploadingAvatar, error: uploadAvatarError }] =
    usersApi.useUploadCurrentUserAvatarMutation();
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const [draftName, setDraftName] = useState<string | null>(null);
  const [draftBio, setDraftBio] = useState<string | null>(null);
  const [draftAvatarUrl, setDraftAvatarUrl] = useState<string | null>(null);
  const [draftAvatarFile, setDraftAvatarFile] = useState<File | null>(null);
  const [avatarError, setAvatarError] = useState<string | null>(null);
  const [isReputationDrawerOpen, setIsReputationDrawerOpen] = useState(false);
  const fileInputRef = useRef<HTMLInputElement | null>(null);

  const currentName = draftName ?? (data?.name ?? "");
  const currentBio = draftBio ?? (data?.bio ?? "");
  const currentAvatarUrl = draftAvatarUrl ?? (data?.avatarUrl ?? "");

  const hasChanges = useMemo(() => {
    if (!data) {
      return false;
    }

    return (
      currentName !== (data.name ?? "") ||
      currentBio !== (data.bio ?? "") ||
      currentAvatarUrl !== (data.avatarUrl ?? "")
    );
  }, [currentAvatarUrl, currentBio, currentName, data]);

  const normalizedAvatarUrl = currentAvatarUrl.trim();
  const avatarPreviewUrl = normalizedAvatarUrl || undefined;
  const hasAvatarPreview = Boolean(avatarPreviewUrl);
  const isSubmitting = isSaving || isUploadingAvatar;

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
      }).unwrap();
      // Drop local draft and rely on fresh server state after mutation invalidation.
      setDraftName(null);
      setDraftBio(null);
      setDraftAvatarUrl(null);
      setDraftAvatarFile(null);
    } catch {
      // Error state is already exposed by RTK Query and rendered in UI.
    }
  };

  const handleClear = () => {
    setDraftName("");
    setDraftBio("");
    setDraftAvatarUrl("");
    setDraftAvatarFile(null);
    setAvatarError(null);
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
              helperText="Чтобы удалить имя или bio, очистите поле и сохраните"
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
            <Button variant="outlined" onClick={() => setIsReputationDrawerOpen(true)}>
              История рейтинга
            </Button>
            <Button component={RouterLink} to="/offer-reports/mine" variant="outlined" color="warning">
              Жалобы на меня
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

          <Alert severity="info">
            История `user_reputation_events` пока не отдается отдельным backend endpoint, поэтому
            в этом drawer сейчас доступно только текущее значение рейтинга.
          </Alert>
        </Box>
      </Drawer>
    </Box>
  );
}

export default ProfilePage;
