import { useMemo, useRef, useState } from "react";
import { Link as RouterLink, useNavigate } from "react-router-dom";
import YandexMapPicker, { type LatLon } from "@/shared/ui/YandexMapPicker";
import {
  Alert,
  Avatar,
  Box,
  Button,
  Card,
  CardContent,
  CircularProgress,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import PersonOutlineIcon from "@mui/icons-material/PersonOutline";
import usersApi from "@/features/users/api/usersApi";
import { useAppDispatch } from "@/hooks/redux";
import { performLogout } from "@/features/auth/model/logoutThunk";
import { imageToAvatarDataUrl } from "@/shared/utils/imageToAvatarDataUrl.ts";
import { getErrorMessage } from "@/shared/utils/getErrorMessage.ts";
import { formatPhoneNumber, formatPhoneNumberInput, isValidPhoneNumber } from "@/shared/utils/phoneNumber.ts";
import { appRoutes } from "@/shared/config/appRoutes.ts";
import ProfileSectionShell from "@/widgets/profile/ProfileSectionShell.tsx";

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
  const [draftPhoneNumber, setDraftPhoneNumber] = useState<string | null>(null);
  const [draftAvatarFile, setDraftAvatarFile] = useState<File | null>(null);
  const [draftLocation, setDraftLocation] = useState<LatLon | null | undefined>(undefined);
  const [avatarError, setAvatarError] = useState<string | null>(null);
  const [phoneError, setPhoneError] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement | null>(null);

  const currentName = draftName ?? (data?.name ?? "");
  const currentBio = draftBio ?? (data?.bio ?? "");
  const currentAvatarUrl = draftAvatarUrl ?? (data?.avatarUrl ?? "");
  const currentPhoneNumber = draftPhoneNumber ?? formatPhoneNumber(data?.phoneNumber) ?? "";
  const currentLocation: LatLon | null = draftLocation !== undefined
    ? draftLocation
    : (data?.currentLatitude != null && data?.currentLongitude != null
        ? { lat: data.currentLatitude, lon: data.currentLongitude }
        : null);

  const hasChanges = useMemo(() => {
    if (!data) {
      return false;
    }

    const serverLat = data.currentLatitude;
    const serverLon = data.currentLongitude;
    const locationChanged = draftLocation !== undefined && (
      draftLocation === null
        ? serverLat != null || serverLon != null
        : draftLocation.lat !== serverLat || draftLocation.lon !== serverLon
    );

    return (
      currentName !== (data.name ?? "") ||
      currentBio !== (data.bio ?? "") ||
      currentAvatarUrl !== (data.avatarUrl ?? "") ||
      currentPhoneNumber !== (data.phoneNumber ?? "") ||
      locationChanged
    );
  }, [currentAvatarUrl, currentBio, currentName, currentPhoneNumber, data, draftLocation]);

  const normalizedAvatarUrl = currentAvatarUrl.trim();
  const avatarPreviewUrl = normalizedAvatarUrl || undefined;
  const hasAvatarPreview = Boolean(avatarPreviewUrl);
  const isSubmitting = isSaving || isUploadingAvatar;

  const handleLogout = async () => {
    await dispatch(performLogout());
    navigate("/login");
  };

  const handleSave = async () => {
    const normalizedPhoneNumber = currentPhoneNumber.trim();
    if (normalizedPhoneNumber && !isValidPhoneNumber(normalizedPhoneNumber)) {
      setPhoneError("Введите номер в формате +7 (999) 123-45-67.");
      return;
    }

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
        phoneNumber: normalizedPhoneNumber,
        currentLatitude: draftLocation !== undefined ? (draftLocation?.lat ?? null) : undefined,
        currentLongitude: draftLocation !== undefined ? (draftLocation?.lon ?? null) : undefined,
      }).unwrap();
      setDraftName(null);
      setDraftBio(null);
      setDraftAvatarUrl(null);
      setDraftPhoneNumber(null);
      setDraftAvatarFile(null);
      setDraftLocation(undefined);
      setPhoneError(null);
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
    setDraftLocation(undefined);
    setAvatarError(null);
    setPhoneError(null);
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
    <ProfileSectionShell
      title="Личные данные"
      description="Здесь живут только данные аккаунта: профиль, контакты, локация и связанные с аккаунтом действия."
      actions={
        <Button variant="outlined" onClick={() => refetch()} disabled={isSubmitting}>
          Обновить
        </Button>
      }
    >
      <Stack spacing={3}>
        <Card variant="outlined">
          <CardContent>
            <Stack spacing={3}>
              <Box display="flex" alignItems="center" gap={2} flexWrap="wrap">
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
                  sx={{ width: 88, height: 88, bgcolor: "action.selected" }}
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
                  <Typography variant="body2" color="text.secondary" mt={0.5}>
                    Зарегистрирован: {new Date(data.createdAt).toLocaleString("ru-RU")}
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

              <Stack spacing={1}>
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
                  <Typography variant="body2">{formatPhoneNumber(data.phoneNumber) || "Не указан"}</Typography>
                </Box>
              </Stack>
            </Stack>
          </CardContent>
        </Card>

        <Card variant="outlined">
          <CardContent>
            <Stack spacing={2}>
              <Typography variant="h6" fontWeight={800}>
                Редактирование профиля
              </Typography>

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
                onChange={(event) => {
                  setDraftPhoneNumber(formatPhoneNumberInput(event.target.value));
                  setPhoneError(null);
                }}
                placeholder="+7 (999) 123-45-67"
                error={Boolean(phoneError)}
                helperText={phoneError || "Формат: +7 (999) 123-45-67"}
                fullWidth
              />
              <Box>
                <Typography variant="subtitle2" mb={1}>
                  Моё местоположение
                </Typography>
                <YandexMapPicker
                  value={currentLocation}
                  onChange={(v) => setDraftLocation(v)}
                  height="280px"
                />
                {currentLocation && (
                  <Typography variant="caption" color="text.secondary" mt={0.5} display="block">
                    {currentLocation.lat.toFixed(6)}, {currentLocation.lon.toFixed(6)}
                  </Typography>
                )}
              </Box>

              {avatarError && (
                <Alert severity="error">
                  {avatarError}
                </Alert>
              )}

              {(normalizedAvatarUrl || draftAvatarUrl !== null) && !avatarError && (
                <Alert severity="info">
                  Аватар выбран локально и будет сохранён вместе с остальными изменениями.
                </Alert>
              )}

              {uploadAvatarError && (
                <Alert severity="error">
                  Не удалось загрузить аватар.
                </Alert>
              )}

              {updateError && (
                <Alert severity="error">
                  {getErrorMessage(updateError) || "Не удалось обновить профиль"}
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
              </Box>
            </Stack>
          </CardContent>
        </Card>

        <Card variant="outlined">
          <CardContent>
            <Stack spacing={2}>
              <Typography variant="h6" fontWeight={800}>
                Действия аккаунта
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Служебные действия убраны из основной формы и живут отдельно, чтобы не конкурировать с данными профиля.
              </Typography>
              <Box display="flex" gap={2} flexWrap="wrap">
                <Button component={RouterLink} to={appRoutes.profile.accountPassword} variant="outlined">
                  Сменить пароль
                </Button>
                <Button variant="outlined" color="error" onClick={handleLogout}>
                  Выйти из аккаунта
                </Button>
              </Box>
            </Stack>
          </CardContent>
        </Card>
      </Stack>
    </ProfileSectionShell>
  );
}

export default ProfilePage;
