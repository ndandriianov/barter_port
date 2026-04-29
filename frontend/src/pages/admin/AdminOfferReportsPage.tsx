import { useEffect, useMemo, useState } from "react";
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
import type { OfferReport, OfferReportStatus, OfferReportsForOffer } from "@/features/offers/model/types.ts";
import usersApi from "@/features/users/api/usersApi.ts";
import { useAppDispatch, useAppSelector } from "@/hooks/redux.ts";
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

function getReporterIds(report: OfferReport, offerReports?: OfferReportsForOffer) {
  const thread = offerReports?.reports.find((entry) => entry.report.id === report.id);

  if (!thread) {
    return [];
  }

  return [...new Set(thread.messages.map((message) => message.authorId))];
}

function AdminOfferReportsPage() {
  const dispatch = useAppDispatch();
  const [filter, setFilter] = useState<AdminReportsFilter>("all");

  const queryStatus = useMemo<OfferReportStatus | void>(() => (filter === "all" ? undefined : filter), [filter]);

  const { data, isLoading, error, refetch, isFetching } = offersApi.useListAdminOfferReportsQuery(queryStatus);
  const { data: adminUsers = [] } = usersApi.useListAdminUsersQuery();

  const offerIds = useMemo(() => {
    if (!data) {
      return [];
    }

    return [...new Set(data.map((report) => report.offerId))];
  }, [data]);

  useEffect(() => {
    if (offerIds.length === 0) {
      return;
    }

    const subscriptions = offerIds.map((offerId) => dispatch(offersApi.endpoints.getOfferReports.initiate(offerId)));
    return () => subscriptions.forEach((subscription) => subscription.unsubscribe());
  }, [dispatch, offerIds]);

  const offerReportsByOfferId = useAppSelector((state) =>
    offerIds.reduce<Record<string, OfferReportsForOffer | undefined>>((acc, offerId) => {
      acc[offerId] = offersApi.endpoints.getOfferReports.select(offerId)(state).data;
      return acc;
    }, {}),
  );

  const usersById = useMemo(
    () =>
      adminUsers.reduce<Record<string, (typeof adminUsers)[number]>>((acc, user) => {
        acc[user.id] = user;
        return acc;
      }, {}),
    [adminUsers],
  );

  const getUserDisplayName = (userId: string) => usersById[userId]?.name?.trim() || "Имя не указано";

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
            (() => {
              const offerReports = offerReportsByOfferId[report.offerId];
              const offerName = offerReports?.offer.name ?? "Загрузка названия объявления...";
              const reporterIds = getReporterIds(report, offerReports);
              const reporterLabel =
                reporterIds.length > 0
                  ? reporterIds.map((reporterId) => getUserDisplayName(reporterId)).join(", ")
                  : "Загрузка автора жалобы...";

              return (
                <Card key={report.id} variant="outlined" sx={{ borderRadius: 3 }}>
                  <CardContent>
                    <Box display="flex" justifyContent="space-between" alignItems="flex-start" gap={2} flexWrap="wrap">
                      <Box>
                        <Typography variant="h6" fontWeight={700}>
                          {offerName}
                        </Typography>
                        <Typography variant="body2" color="text.secondary" mt={0.5}>
                          Автор объявления: {getUserDisplayName(report.offerAuthorId)}
                        </Typography>
                        <Typography variant="body2" color="text.secondary">
                          {reporterIds.length > 1 ? "Пожаловались" : "Пожаловался"}: {reporterLabel}
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
              );
            })()
          ))}
        </Stack>
      )}
    </Stack>
  );
}

export default AdminOfferReportsPage;
