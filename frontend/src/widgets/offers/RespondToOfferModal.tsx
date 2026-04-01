import { useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  Alert,
  Box,
  Button,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControl,
  FormControlLabel,
  Checkbox,
  Typography,
} from "@mui/material";
import offersApi from "@/features/offers/api/offersApi";
import dealsApi from "@/features/deals/api/dealsApi";
import type { Offer } from "@/features/offers/model/types";

interface RespondToOfferModalProps {
  targetOffer: Offer;
  isOpen: boolean;
  onClose: () => void;
}

function RespondToOfferModal({ targetOffer, isOpen, onClose }: RespondToOfferModalProps) {
  const [selectedOfferIds, setSelectedOfferIds] = useState<string[]>([]);
  const navigate = useNavigate();

  const closeModal = () => {
    setSelectedOfferIds([]);
    onClose();
  };

  const { data, isLoading, error } = offersApi.useGetOffersQuery(
    { sort: "ByTime", my: true, cursor_limit: 100 },
    { skip: !isOpen },
  );

  const [createDraftDeal, { isLoading: isCreating, error: createError }] =
    dealsApi.useCreateDraftDealMutation();

  const selectedOffers = useMemo(
    () => data?.offers.filter((entry) => selectedOfferIds.includes(entry.id)) ?? [],
    [data?.offers, selectedOfferIds],
  );

  const toggleSelectedOffer = (offerId: string) => {
    setSelectedOfferIds((current) =>
      current.includes(offerId)
        ? current.filter((id) => id !== offerId)
        : [...current, offerId],
    );
  };

  const submit = async () => {
    if (selectedOffers.length === 0) return;
    const result = await createDraftDeal({
      offers: [
        { offerID: targetOffer.id, quantity: 1 },
        ...selectedOffers.map((offer) => ({ offerID: offer.id, quantity: 1 })),
      ],
    }).unwrap();
    closeModal();
    navigate(`/deals/drafts/${result.id}`);
  };

  return (
    <Dialog open={isOpen} onClose={closeModal} maxWidth="sm" fullWidth>
      <DialogTitle>Откликнуться на объявление</DialogTitle>

      <DialogContent dividers>
        <Typography variant="body2" color="text.secondary" mb={2}>
          Выберите одно или несколько своих объявлений для обмена:
        </Typography>

        {isLoading && (
          <Box display="flex" justifyContent="center" py={3}>
            <CircularProgress />
          </Box>
        )}

        {error && <Alert severity="error">Не удалось загрузить ваши объявления</Alert>}

        {!isLoading && !error && data && (
          data.offers.length === 0 ? (
            <Typography color="text.secondary">У вас пока нет объявлений для отклика</Typography>
          ) : (
            <FormControl component="fieldset" fullWidth>
              <Box display="flex" flexDirection="column">
                {data.offers.map((offer) => (
                  <FormControlLabel
                    key={offer.id}
                    control={
                      <Checkbox
                        checked={selectedOfferIds.includes(offer.id)}
                        onChange={() => toggleSelectedOffer(offer.id)}
                      />
                    }
                    label={
                      <Box>
                        <Typography variant="body1" fontWeight={500}>{offer.name}</Typography>
                        {offer.description && (
                          <Typography variant="body2" color="text.secondary">{offer.description}</Typography>
                        )}
                      </Box>
                    }
                    sx={{ mb: 1, alignItems: "flex-start", "& .MuiCheckbox-root": { mt: 0.5 } }}
                  />
                ))}
              </Box>
            </FormControl>
          )
        )}

        {createError && (
          <Alert severity="error" sx={{ mt: 2 }}>
            Не удалось создать черновик сделки
          </Alert>
        )}
      </DialogContent>

      <DialogActions>
        <Button onClick={closeModal} color="inherit">
          Отмена
        </Button>
        <Button
          variant="contained"
          onClick={submit}
          disabled={selectedOfferIds.length === 0 || isCreating || isLoading || !!error}
        >
          {isCreating ? <CircularProgress size={20} color="inherit" /> : "Создать черновик"}
        </Button>
      </DialogActions>
    </Dialog>
  );
}

export default RespondToOfferModal;
