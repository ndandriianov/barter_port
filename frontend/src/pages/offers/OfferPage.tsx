import { useState } from "react";
import { useParams } from "react-router-dom";
import { Alert, Box, Button, Divider, Typography } from "@mui/material";
import { useAppSelector } from "@/hooks/redux";
import offersApi from "@/features/offers/api/offersApi";
import usersApi from "@/features/users/api/usersApi";
import type { GetOffersResponse, Offer } from "@/features/offers/model/types";
import OfferCard from "@/widgets/offers/OfferCard";
import RespondToOfferModal from "@/widgets/offers/RespondToOfferModal";

interface CachedQueryState {
  endpointName?: string;
  data?: unknown;
}

function isGetOffersResponse(data: unknown): data is GetOffersResponse {
  return (
    typeof data === "object" &&
    data !== null &&
    "offers" in data &&
    Array.isArray(data.offers)
  );
}

function OfferPage() {
  const { offerId } = useParams<{ offerId: string }>();
  const [isRespondModalOpen, setIsRespondModalOpen] = useState(false);
  const { data: meData } = usersApi.useGetCurrentUserQuery();

  const offer = useAppSelector((state) => {
    const queries = state[offersApi.reducerPath].queries;
    const queryStates = Object.values(queries) as CachedQueryState[];

    for (const queryState of queryStates) {
      if (queryState?.endpointName !== "getOffers" || !isGetOffersResponse(queryState.data)) {
        continue;
      }
      const match = queryState.data.offers.find((entry: Offer) => entry.id === offerId);
      if (match) return match;
    }

    return null;
  });

  if (!offerId || !offer) {
    return <Alert severity="warning">Объявление не найдено</Alert>;
  }

  const canRespond = !!meData && offer.authorId !== meData.id;

  return (
    <Box maxWidth={700} mx="auto">
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
