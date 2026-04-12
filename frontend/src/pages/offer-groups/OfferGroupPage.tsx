import { useMemo, useState } from "react";
import { Link as RouterLink, useNavigate, useParams } from "react-router-dom";
import {
  Alert,
  Box,
  Button,
  Chip,
  CircularProgress,
  Divider,
  Grid,
  Paper,
  Typography,
} from "@mui/material";
import usersApi from "@/features/users/api/usersApi.ts";
import offerGroupsApi from "@/features/offer-groups/api/offerGroupsApi.ts";
import {
  getOfferGroupOwnerId,
  getOfferGroupOwnerName,
  getOfferGroupUniformAction,
} from "@/features/offer-groups/model/utils.ts";
import OfferCard from "@/widgets/offers/OfferCard.tsx";
import RespondToOfferGroupModal from "@/widgets/offer-groups/RespondToOfferGroupModal.tsx";

const actionLabels = {
  give: "Отдаю",
  take: "Ищу",
} as const;

function OfferGroupPage() {
  const { offerGroupId } = useParams<{ offerGroupId: string }>();
  const navigate = useNavigate();
  const [isRespondModalOpen, setIsRespondModalOpen] = useState(false);
  const { data: me } = usersApi.useGetCurrentUserQuery();

  const { data: offerGroup, isLoading, error } = offerGroupsApi.useGetOfferGroupByIdQuery(offerGroupId ?? "", {
    skip: !offerGroupId,
  });

  const ownerId = useMemo(
    () => (offerGroup ? getOfferGroupOwnerId(offerGroup) : undefined),
    [offerGroup],
  );
  const uniformAction = useMemo(
    () => (offerGroup ? getOfferGroupUniformAction(offerGroup) : null),
    [offerGroup],
  );
  const isOwnOfferGroup = !!me && !!ownerId && me.id === ownerId;

  if (!offerGroupId) {
    return <Alert severity="warning">Композитное объявление не найдено</Alert>;
  }

  if (isLoading) {
    return (
      <Box display="flex" justifyContent="center" py={6}>
        <CircularProgress />
      </Box>
    );
  }

  if (error || !offerGroup) {
    return <Alert severity="warning">Композитное объявление не найдено</Alert>;
  }

  return (
    <Box maxWidth={1100} mx="auto">
      <Button
        size="small"
        variant="text"
        onClick={() => (window.history.length > 1 ? navigate(-1) : navigate("/offer-groups"))}
        sx={{ mb: 2 }}
      >
        ← Назад
      </Button>

      <Box display="flex" justifyContent="space-between" alignItems="flex-start" gap={2} mb={3} flexWrap="wrap">
        <Box>
          <Typography variant="h4" fontWeight={700} mb={1}>
            {offerGroup.name}
          </Typography>
          <Typography variant="body1" color="text.secondary">
            Автор: {getOfferGroupOwnerName(offerGroup)}
          </Typography>
        </Box>

        <Box display="flex" gap={1} flexWrap="wrap">
          <Chip label={`${offerGroup.units.length} AND-блок(ов)`} color="primary" variant="outlined" />
          <Chip
            label={`${offerGroup.units.reduce((total, unit) => total + unit.offers.length, 0)} вариантов`}
            color="info"
            variant="outlined"
          />
        </Box>
      </Box>

      <Paper
        variant="outlined"
        sx={{
          p: 3,
          mb: 3,
          background:
            "linear-gradient(135deg, rgba(7,116,129,0.08) 0%, rgba(255,255,255,1) 55%)",
        }}
      >
        <Typography variant="body1" color="text.secondary">
          {offerGroup.description?.trim() || "Описание не указано"}
        </Typography>
        <Alert severity={uniformAction ? "info" : "success"} sx={{ mt: 2 }}>
          {uniformAction
            ? `У всех unit одинаковый action: "${actionLabels[uniformAction]}". При отклике потребуется приложить свой offer с тем же action.`
            : "В группе есть разные action. При отклике можно выбрать unit-варианты и при желании приложить свой offer."}
        </Alert>
      </Paper>

      <Box display="flex" gap={2} flexWrap="wrap" mb={3}>
        {!isOwnOfferGroup && (
          <Button variant="contained" onClick={() => setIsRespondModalOpen(true)}>
            Откликнуться на группу
          </Button>
        )}
        <Button component={RouterLink} to="/offer-groups" variant="outlined">
          Все группы
        </Button>
      </Box>

      <Divider sx={{ mb: 3 }} />

      <Box display="flex" flexDirection="column" gap={4}>
        {offerGroup.units.map((unit, unitIndex) => (
          <Box key={unit.id}>
            <Box mb={2}>
              <Typography variant="h5" fontWeight={700}>
                Unit {unitIndex + 1}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Выберите один из этих вариантов при отклике.
              </Typography>
            </Box>

            <Grid container spacing={2}>
              {unit.offers.map((offer) => (
                <Grid key={offer.id} size={{ xs: 12, md: 6 }}>
                  <OfferCard offer={offer} offerHref={`/offers/${offer.id}`} showRating />
                </Grid>
              ))}
            </Grid>
          </Box>
        ))}
      </Box>

      <RespondToOfferGroupModal
        offerGroup={offerGroup}
        isOpen={isRespondModalOpen}
        onClose={() => setIsRespondModalOpen(false)}
      />
    </Box>
  );
}

export default OfferGroupPage;
