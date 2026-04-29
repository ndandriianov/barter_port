import { useMemo, useState } from "react";
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
import type { OfferReportStatus } from "@/features/offers/model/types.ts";
import { getStatusCode } from "@/shared/utils/getStatusCode.ts";

type AdminReportsFilter = "all" | OfferReportStatus;

function getReportStatusMeta(status: OfferReportStatus) {
  if (status === "Pending") {
    return { label: "На модерации", color: "warning" as const };
  }
  if (status === "Accepted") {
    return { label: "Принята", color: "error" as const };
  }
  return { label: "Отклонена", color: "success" as const };
}

const filterOptions: Array<{ value: AdminReportsFilter; label: string }> = [
  { value: "all", label: "Все" },
  { value: "Pending", label: "На модерации" },
  { value: "Accepted", label: "Принятые" },
  { value: "Rejected", label: "Отклоненные" },
];

function AdminOfferReportsPage() {
  const [filter, setFilter] = useState<AdminReportsFilter>("all");

  const queryStatus = useMemo<OfferReportStatus | void>(() => (filter === "all" ? undefined : filter), [filter]);

  const { data, isLoading, error, refetch, isFetching } = offersApi.useListAdminOfferReportsQuery(queryStatus);

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
        </Box>
        <Stack direction="row" spacing={1} useFlexGap flexWrap="wrap" alignItems="center">
          {filterOptions.map((option) => (
            <Button
              key={option.value}
              variant={filter === option.value ? "contained" : "outlined"}
              onClick={() => setFilter(option.value)}
            >
              {option.label}
            </Button>
          ))}
          <Button variant="outlined" onClick={() => refetch()} disabled={isFetching}>
            Обновить
          </Button>
        </Stack>
      </Box>

      {!data || data.length === 0 ? (
        <Alert severity="success">
          {filter === "all"
            ? "Жалоб на объявления сейчас нет."
            : "Жалоб с выбранным статусом сейчас нет."}
        </Alert>
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
                    <Chip
                      icon={<GavelOutlinedIcon />}
                      label={getReportStatusMeta(report.status).label}
                      color={getReportStatusMeta(report.status).color}
                    />
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

