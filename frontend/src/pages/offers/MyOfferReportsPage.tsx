import { useMemo } from "react";
import { Link as RouterLink } from "react-router-dom";
import {
  Accordion,
  AccordionDetails,
  AccordionSummary,
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  CircularProgress,
  Divider,
  Stack,
  Typography,
} from "@mui/material";
import ExpandMoreIcon from "@mui/icons-material/ExpandMore";
import ReportProblemOutlinedIcon from "@mui/icons-material/ReportProblemOutlined";
import offersApi from "@/features/offers/api/offersApi.ts";
import type { Offer, OfferReportThread } from "@/features/offers/model/types.ts";
import {
  getOfferModerationLabel,
  getOfferModerationState,
} from "@/features/offers/model/getOfferModerationState.ts";
import { getStatusCode } from "@/shared/utils/getStatusCode.ts";

function OfferReportStatusChip({ status }: { status: OfferReportThread["report"]["status"] }) {
  const color = status === "Pending" ? "warning" : status === "Accepted" ? "error" : "success";
  const label = status === "Pending" ? "На модерации" : status === "Accepted" ? "Принята" : "Отклонена";
  return <Chip label={label} color={color} size="small" variant="outlined" />;
}

function MyOfferReportsItem({ offer }: { offer: Offer }) {
  const { data, isLoading, error } = offersApi.useGetOfferReportsQuery(offer.id);

  if (isLoading) {
    return (
      <Card variant="outlined" sx={{ borderRadius: 3 }}>
        <CardContent>
          <Box display="flex" justifyContent="center" py={2}>
            <CircularProgress size={24} />
          </Box>
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Alert severity="error" variant="outlined">
        Не удалось загрузить жалобы по объявлению «{offer.name}».
      </Alert>
    );
  }

  if (!data || data.reports.length === 0) {
    return null;
  }

  const moderationState = getOfferModerationState(offer, data);
  const moderationLabel = getOfferModerationLabel(moderationState);

  return (
    <Card variant="outlined" sx={{ borderRadius: 3 }}>
      <CardContent>
        <Box display="flex" justifyContent="space-between" gap={2} flexWrap="wrap" mb={2}>
          <Box>
            <Typography variant="h6" fontWeight={700}>
              {offer.name}
            </Typography>
            <Typography variant="body2" color="text.secondary" mt={0.5}>
              {offer.description}
            </Typography>
          </Box>
          <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
            {moderationLabel && (
              <Chip
                label={moderationLabel}
                color={moderationState === "hidden" ? "error" : moderationState === "pending" ? "warning" : "info"}
              />
            )}
            <Button component={RouterLink} to={`/offers/${offer.id}`} variant="outlined" size="small">
              Открыть объявление
            </Button>
          </Stack>
        </Box>

        <Stack spacing={1.5}>
          {data.reports.map((thread) => (
            <Accordion key={thread.report.id} disableGutters sx={{ borderRadius: 2, "&::before": { display: "none" } }}>
              <AccordionSummary expandIcon={<ExpandMoreIcon />}>
                <Box display="flex" justifyContent="space-between" alignItems="center" width="100%" gap={2} flexWrap="wrap">
                  <Box>
                    <Typography fontWeight={600}>
                      Жалоба от {new Date(thread.report.createdAt).toLocaleString("ru-RU")}
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      Сообщений: {thread.messages.length}
                    </Typography>
                  </Box>
                  <OfferReportStatusChip status={thread.report.status} />
                </Box>
              </AccordionSummary>
              <AccordionDetails>
                <Stack spacing={1.5}>
                  {thread.messages.map((message, index) => (
                    <Box key={`${message.authorId}-${index}`} sx={{ p: 1.5, borderRadius: 2, bgcolor: "action.hover" }}>
                      <Typography variant="caption" color="text.secondary" display="block" mb={0.5}>
                        Автор жалобы: {message.authorId}
                      </Typography>
                      <Typography variant="body2">{message.message}</Typography>
                    </Box>
                  ))}

                  {thread.report.resolutionComment && (
                    <>
                      <Divider />
                      <Box>
                        <Typography variant="caption" color="text.secondary" display="block" mb={0.5}>
                          Комментарий модератора
                        </Typography>
                        <Typography variant="body2">{thread.report.resolutionComment}</Typography>
                      </Box>
                    </>
                  )}
                </Stack>
              </AccordionDetails>
            </Accordion>
          ))}
        </Stack>
      </CardContent>
    </Card>
  );
}

function MyOfferReportsPage() {
  const { data, isLoading, error, refetch, isFetching } = offersApi.useGetOffersQuery({
    sort: "ByTime",
    my: true,
    cursor_limit: 50,
  });

  const offers = useMemo(() => data?.offers ?? [], [data]);

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
        {getStatusCode(error) === 401
          ? "Нужно авторизоваться, чтобы увидеть жалобы на свои объявления."
          : "Не удалось загрузить ваши объявления."}
      </Alert>
    );
  }

  return (
    <Stack spacing={3}>
      <Box display="flex" justifyContent="space-between" alignItems="flex-start" gap={2} flexWrap="wrap">
        <Box>
          <Typography variant="h4" fontWeight={700} mb={1}>
            Жалобы на мои объявления
          </Typography>
          <Typography variant="body1" color="text.secondary" maxWidth={760}>
            Здесь отображаются только объявления, по которым уже были жалобы. У каждого объявления
            список жалоб свернут по умолчанию и раскрывается по нажатию.
          </Typography>
        </Box>
        <Button variant="outlined" onClick={() => refetch()} disabled={isFetching}>
          Обновить
        </Button>
      </Box>

      {offers.length === 0 ? (
        <Alert severity="info">У вас пока нет объявлений.</Alert>
      ) : (
        <Stack spacing={2}>
          {offers.map((offer) => (
            <MyOfferReportsItem key={offer.id} offer={offer} />
          ))}
        </Stack>
      )}

      <Alert severity="info" icon={<ReportProblemOutlinedIcon />}>
        Если список пуст, значит по вашим объявлениям пока нет жалоб или сервер еще не вернул их в
        текущую выборку.
      </Alert>
    </Stack>
  );
}

export default MyOfferReportsPage;

