import { useNavigate } from "react-router-dom";
import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  CircularProgress,
  Divider,
  Typography,
} from "@mui/material";
import PersonOutlineIcon from "@mui/icons-material/PersonOutline";
import authApi from "@/features/auth/api/authApi";
import { useAppDispatch } from "@/hooks/redux";
import { performLogout } from "@/features/auth/model/logoutThunk";

function ProfilePage() {
  const { data, isLoading, refetch } = authApi.useMeQuery();
  const dispatch = useAppDispatch();
  const navigate = useNavigate();

  const handleLogout = async () => {
    await dispatch(performLogout());
    navigate("/login");
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
    <Box maxWidth={480} mx="auto">
      <Typography variant="h4" fontWeight={700} mb={3}>
        Профиль
      </Typography>

      <Card variant="outlined">
        <CardContent>
          <Box display="flex" alignItems="center" gap={2} mb={2}>
            <PersonOutlineIcon fontSize="large" color="action" />
            <Box>
              <Typography variant="caption" color="text.secondary">
                User ID
              </Typography>
              <Typography variant="body2" fontFamily="monospace" fontWeight={500}>
                {data.userId}
              </Typography>
            </Box>
          </Box>

          <Divider sx={{ mb: 2 }} />

          <Box display="flex" gap={2}>
            <Button variant="contained" color="error" onClick={handleLogout}>
              Выйти
            </Button>
            <Button variant="outlined" onClick={() => refetch()}>
              Обновить
            </Button>
          </Box>
        </CardContent>
      </Card>
    </Box>
  );
}

export default ProfilePage;
