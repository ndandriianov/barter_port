import { Link as RouterLink } from "react-router-dom";
import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  CircularProgress,
  Stack,
  Typography,
} from "@mui/material";
import GavelOutlinedIcon from "@mui/icons-material/GavelOutlined";
import offersApi from "@/features/offers/api/offersApi.ts";
import { getStatusCode } from "@/shared/utils/getStatusCode.ts";

function AdminOfferReportsPage() {
  const { data, isLoading, error, refetch, isFetching } = offersApi.useListAdminOfferReportsQuery("Pending");

  if (isLoading) {
    return (
      <Box display="flex" justifyContent="center" py={8}>
        <CircularProgress />
      </Box>
    );
  }

  if (error) {
    return (
      <Alert severity="error">
        {getStatusCode(error) === 403
          ? "Раздел доступен только администратору."
          : "Не удалось загрузить очередь жалоб на объявления."}
      </Alert>
    );
  }

  return (
    <Stack spacing={3}>
      <Box display="flex" justifyContent="space-between" alignItems="flex-start" gap={2} flexWrap="wrap">
        <Box>
          <Typography variant="h4" fontWeight={700} mb={1}>
            Жалобы на объявления
          </Typography>
          <Typography variant="body1" color="text.secondary" maxWidth={760}>
            Очередь жалоб, ожидающих решения администратора. На детальной странице доступны
            материалы жалобы и действия модерации.
          </Typography>
        </Box>
        <Button variant="outlined" onClick={() => refetch()} disabled={isFetching}>
          Обновить
        </Button>
      </Box>

      {!data || data.length === 0 ? (
        <Alert severity="success">Активных жалоб на объявления сейчас нет.</Alert>
      ) : (
        <Stack spacing={2}>
          {data.map((report) => (
            <Card key={report.id} variant="outlined" sx={{ borderRadius: 3 }}>
              <CardContent>
                <Box display="flex" justifyContent="space-between" alignItems="flex-start" gap={2} flexWrap="wrap">
                  <Box>
                    <Typography variant="h6" fontWeight={700}>
                      Жалоба {report.id}
                    </Typography>
                    <Typography variant="body2" color="text.secondary" mt={0.5}>
                      Объявление: {report.offerId}
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      Автор объявления: {report.offerAuthorId}
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      Создана: {new Date(report.createdAt).toLocaleString("ru-RU")}
                    </Typography>
                  </Box>

                  <Stack direction="row" spacing={1} useFlexGap flexWrap="wrap">
                    <Chip icon={<GavelOutlinedIcon />} label="На модерации" color="warning" />
                    <Button component={RouterLink} to={`/admin/offer-reports/${report.id}`} variant="contained">
                      Открыть
                    </Button>
                  </Stack>
                </Box>
              </CardContent>
            </Card>
          ))}
        </Stack>
      )}
    </Stack>
  );
}

export default AdminOfferReportsPage;

