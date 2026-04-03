import { useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import {Alert, Box, Button, Card, CardContent, CircularProgress, Divider, Stack, TextField, Typography,} from "@mui/material";
import PersonOutlineIcon from "@mui/icons-material/PersonOutline";
import usersApi from "@/features/users/api/usersApi";
import { useAppDispatch } from "@/hooks/redux";
import { performLogout } from "@/features/auth/model/logoutThunk";

function ProfilePage() {
  const { data, isLoading, refetch } = usersApi.useGetCurrentUserQuery();
  const [updateCurrentUser, { isLoading: isSaving, error: updateError }] =
    usersApi.useUpdateCurrentUserMutation();
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const [draftName, setDraftName] = useState<string | null>(null);
  const [draftBio, setDraftBio] = useState<string | null>(null);

  const currentName = draftName ?? (data?.name ?? "");
  const currentBio = draftBio ?? (data?.bio ?? "");

  const hasChanges = useMemo(() => {
    if (!data) {
      return false;
    }

    return currentName !== (data.name ?? "") || currentBio !== (data.bio ?? "");
  }, [currentBio, currentName, data]);

  const handleLogout = async () => {
    await dispatch(performLogout());
    navigate("/login");
  };

  const handleSave = async () => {
    try {
      await updateCurrentUser({
        name: currentName.trim(),
        bio: currentBio.trim(),
      }).unwrap();
      // Drop local draft and rely on fresh server state after mutation invalidation.
      setDraftName(null);
      setDraftBio(null);
    } catch {
      // Error state is already exposed by RTK Query and rendered in UI.
    }
  };

  const handleClear = () => {
    setDraftName("");
    setDraftBio("");
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
            <PersonOutlineIcon fontSize="large" color="action" />
            <Box>
              <Typography variant="caption" color="text.secondary">
                ID
              </Typography>
              <Typography variant="body2" fontFamily="monospace" fontWeight={500}>
                {data.id}
              </Typography>
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

          {updateError && (
            <Alert severity="error" sx={{ mb: 2 }}>
              Не удалось обновить профиль
            </Alert>
          )}

          <Box display="flex" gap={2} flexWrap="wrap">
            <Button
              variant="contained"
              onClick={handleSave}
              disabled={!hasChanges || isSaving}
            >
              Сохранить
            </Button>
            <Button variant="outlined" onClick={handleClear} disabled={isSaving}>
              Очистить поля
            </Button>
            <Button variant="contained" color="error" onClick={handleLogout}>
              Выйти
            </Button>
            <Button variant="outlined" onClick={() => refetch()} disabled={isSaving}>
              Обновить
            </Button>
          </Box>
        </CardContent>
      </Card>
    </Box>
  );
}

export default ProfilePage;
