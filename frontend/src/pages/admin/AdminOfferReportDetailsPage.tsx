import { useMemo, useState } from "react";
import { Link as RouterLink, useNavigate, useParams } from "react-router-dom";
import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  CircularProgress,
  Divider,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import offersApi from "@/features/offers/api/offersApi.ts";
import OfferCard from "@/widgets/offers/OfferCard.tsx";
import { getErrorMessage } from "@/shared/utils/getErrorMessage.ts";
import { getStatusCode } from "@/shared/utils/getStatusCode.ts";

function AdminOfferReportDetailsPage() {
  const { reportId } = useParams<{ reportId: string }>();
  const navigate = useNavigate();
  const [comment, setComment] = useState("");
  const { data, isLoading, error } = offersApi.useGetAdminOfferReportByIdQuery(reportId ?? "", {
    skip: !reportId,
  });
  const [resolveReport, { isLoading: isResolving, error: resolveError }] =
    offersApi.useResolveAdminOfferReportMutation();

  const hasResolution = useMemo(() => data?.report.status !== "Pending", [data?.report.status]);

  if (!reportId) {
    return <Alert severity="warning">Жалоба не найдена.</Alert>;
  }

  if (isLoading) {
    return (
      <Box display="flex" justifyContent="center" py={8}>
        <CircularProgress />
      </Box>
    );
  }

  if (error || !data) {
    return (
      <Alert severity="error">
        {getStatusCode(error) === 404
          ? "Жалоба не найдена."
          : getStatusCode(error) === 403
            ? "Раздел доступен только администратору."
            : "Не удалось загрузить материалы жалобы."}
      </Alert>
    );
  }

  const handleResolve = async (accepted: boolean) => {
    try {
      await resolveReport({
        reportId,
        body: {
          accepted,
          comment: comment.trim() || undefined,
        },
      }).unwrap();
      navigate("/admin/offer-reports", { replace: true });
    } catch {
      // Error is rendered below.
    }
  };

  const statusLabel =
    data.report.status === "Pending"
      ? "На модерации"
      : data.report.status === "Accepted"
        ? "Принята"
        : "Отклонена";

  return (
    <Stack spacing={3}>
      <Box display="flex" justifyContent="space-between" alignItems="flex-start" gap={2} flexWrap="wrap">
        <Box>
          <Typography variant="h4" fontWeight={700} mb={1}>
            Разбор жалобы
          </Typography>
          <Typography variant="body1" color="text.secondary">
            Здесь администратор видит объявление, все сообщения жалобы и может принять решение.
          </Typography>
        </Box>
        <Stack direction="row" spacing={1} useFlexGap flexWrap="wrap">
          <Chip label={statusLabel} color={data.report.status === "Pending" ? "warning" : data.report.status === "Accepted" ? "error" : "success"} />
          <Button component={RouterLink} to="/admin/offer-reports" variant="outlined">
            Назад к списку
          </Button>
        </Stack>
      </Box>

      <Card variant="outlined" sx={{ borderRadius: 3 }}>
        <CardContent>
          <Typography variant="h6" fontWeight={700} mb={2}>
            Объявление
          </Typography>
          <OfferCard offer={data.offer} />
        </CardContent>
      </Card>

      <Card variant="outlined" sx={{ borderRadius: 3 }}>
        <CardContent>
          <Typography variant="h6" fontWeight={700} mb={2}>
            Сообщения жалобы
          </Typography>
          <Stack spacing={1.5}>
            {data.messages.map((message, index) => (
              <Box key={`${message.authorId}-${index}`} sx={{ p: 1.75, borderRadius: 2, bgcolor: "action.hover" }}>
                <Typography variant="caption" color="text.secondary" display="block" mb={0.5}>
                  Автор: {message.authorId}
                </Typography>
                <Typography variant="body2">{message.message}</Typography>
              </Box>
            ))}
          </Stack>
        </CardContent>
      </Card>

      <Card variant="outlined" sx={{ borderRadius: 3 }}>
        <CardContent>
          <Typography variant="h6" fontWeight={700} mb={2}>
            Решение администратора
          </Typography>
          <Stack spacing={2}>
            {data.report.resolutionComment && (
              <Alert severity="info">Комментарий по жалобе: {data.report.resolutionComment}</Alert>
            )}

            <TextField
              label="Комментарий администратора"
              value={comment}
              onChange={(event) => setComment(event.target.value)}
              multiline
              minRows={4}
              disabled={hasResolution || isResolving}
            />

            {resolveError && (
              <Alert severity="error">
                {getErrorMessage(resolveError) ??
                  (getStatusCode(resolveError) === 409
                    ? "Эта жалоба уже была рассмотрена."
                    : "Не удалось сохранить решение администратора.")}
              </Alert>
            )}

            {hasResolution ? (
              <Alert severity="info">
                Жалоба уже рассмотрена. Текущий статус: <strong>{statusLabel}</strong>.
              </Alert>
            ) : (
              <>
                <Divider />
                <Box display="flex" gap={2} flexWrap="wrap">
                  <Button
                    variant="contained"
                    color="error"
                    onClick={() => handleResolve(true)}
                    disabled={isResolving}
                  >
                    Принять жалобу
                  </Button>
                  <Button
                    variant="outlined"
                    color="inherit"
                    onClick={() => handleResolve(false)}
                    disabled={isResolving}
                  >
                    Отклонить жалобу
                  </Button>
                </Box>
              </>
            )}
          </Stack>
        </CardContent>
      </Card>
    </Stack>
  );
}

export default AdminOfferReportDetailsPage;
