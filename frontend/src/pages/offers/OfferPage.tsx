import { useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { Alert, Box, Button, CircularProgress, Divider, Typography } from "@mui/material";
import offersApi from "@/features/offers/api/offersApi";
import usersApi from "@/features/users/api/usersApi";
import OfferCard from "@/widgets/offers/OfferCard";
import RespondToOfferModal from "@/widgets/offers/RespondToOfferModal";

function OfferPage() {
  const { offerId } = useParams<{ offerId: string }>();
  const navigate = useNavigate();
  const [isRespondModalOpen, setIsRespondModalOpen] = useState(false);
  const { data: meData } = usersApi.useGetCurrentUserQuery();

  const { data: offer, isLoading, error } = offersApi.useGetOfferByIdQuery(offerId ?? "", {
    skip: !offerId,
  });

  if (!offerId) return <Alert severity="warning">Объявление не найдено</Alert>;

  if (isLoading) {
    return (
      <Box display="flex" justifyContent="center" py={6}>
        <CircularProgress />
      </Box>
    );
  }

  if (error || !offer) {
    return <Alert severity="warning">Объявление не найдено</Alert>;
  }

  const canRespond = !!meData && offer.authorId !== meData.id;

  return (
    <Box maxWidth={700} mx="auto">
      <Button
        size="small"
        variant="text"
        onClick={() => window.history.length > 1 ? navigate(-1) : navigate("/offers")}
        sx={{ mb: 2 }}
      >
        ← Назад
      </Button>

      <Typography variant="h4" fontWeight={700} mb={3}>
        {offer.name}
      </Typography>

      <OfferCard offer={offer} />

      <Divider sx={{ my: 3 }} />

      <Box display="flex" gap={2}>
        {canRespond && (
          <Button variant="contained" onClick={() => setIsRespondModalOpen(true)}>
            Откликнуться
          </Button>
        )}
        <Button variant="outlined" color="error">
          Пожаловаться
        </Button>
      </Box>

      <RespondToOfferModal
        targetOffer={offer}
        isOpen={isRespondModalOpen}
        onClose={() => setIsRespondModalOpen(false)}
      />
    </Box>
  );
}

export default OfferPage;
