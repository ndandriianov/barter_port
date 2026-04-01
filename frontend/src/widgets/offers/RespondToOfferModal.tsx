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
  Radio,
  RadioGroup,
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
  const [selectedOfferId, setSelectedOfferId] = useState<string | null>(null);
  const navigate = useNavigate();

  const closeModal = () => {
    setSelectedOfferId(null);
    onClose();
  };

  const { data, isLoading, error } = offersApi.useGetOffersQuery(
    { sort: "ByTime", my: true, cursor_limit: 100 },
    { skip: !isOpen },
  );

  const [createDraftDeal, { isLoading: isCreating, error: createError }] =
    dealsApi.useCreateDraftDealMutation();

  const selectedOffer = useMemo(
    () => data?.offers.find((entry) => entry.id === selectedOfferId),
    [data?.offers, selectedOfferId],
  );

  const submit = async () => {
    if (!selectedOffer) return;
    const result = await createDraftDeal({
      offers: [
        { offerID: targetOffer.id, quantity: 1 },
        { offerID: selectedOffer.id, quantity: 1 },
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
          Выберите своё объявление для обмена:
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
              <RadioGroup
                value={selectedOfferId ?? ""}
                onChange={(e) => setSelectedOfferId(e.target.value)}
              >
                {data.offers.map((offer) => (
                  <FormControlLabel
                    key={offer.id}
                    value={offer.id}
                    control={<Radio />}
                    label={
                      <Box>
                        <Typography variant="body1" fontWeight={500}>{offer.name}</Typography>
                        {offer.description && (
                          <Typography variant="body2" color="text.secondary">{offer.description}</Typography>
                        )}
                      </Box>
                    }
                    sx={{ mb: 1, alignItems: "flex-start", "& .MuiRadio-root": { mt: 0.5 } }}
                  />
                ))}
              </RadioGroup>
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
          disabled={!selectedOfferId || isCreating || isLoading || !!error}
        >
          {isCreating ? <CircularProgress size={20} color="inherit" /> : "Создать черновик"}
        </Button>
      </DialogActions>
    </Dialog>
  );
}

export default RespondToOfferModal;
