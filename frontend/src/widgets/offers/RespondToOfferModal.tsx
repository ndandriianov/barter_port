import { useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  Alert,
  Box,
  Button,
  Checkbox,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControl,
  FormControlLabel,
  TextField,
  Typography,
} from "@mui/material";
import offersApi from "@/features/offers/api/offersApi";
import dealsApi from "@/features/deals/api/dealsApi";
import type { Offer } from "@/features/offers/model/types";
import { getCreateDraftDealErrorMessage } from "@/shared/utils/getCreateDraftErrorMessage.ts";

interface RespondToOfferModalProps {
  targetOffer: Offer;
  isOpen: boolean;
  onClose: () => void;
}

function RespondToOfferModal({ targetOffer, isOpen, onClose }: RespondToOfferModalProps) {
  const [selectedOfferIds, setSelectedOfferIds] = useState<string[]>([]);
  const [offerQuantities, setOfferQuantities] = useState<Record<string, string>>({});
  const navigate = useNavigate();

  const closeModal = () => {
    setSelectedOfferIds([]);
    setOfferQuantities({});
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

  const quantitiesByOfferId = useMemo(() => {
    return Object.fromEntries(
      selectedOfferIds.map((offerId) => {
        const quantity = Number.parseInt(offerQuantities[offerId] ?? "1", 10);
        return [offerId, Number.isInteger(quantity) && quantity >= 1 ? quantity : null];
      }),
    ) as Record<string, number | null>;
  }, [offerQuantities, selectedOfferIds]);

  const hasInvalidQuantity = selectedOfferIds.some((offerId) => quantitiesByOfferId[offerId] === null);

  const toggleSelectedOffer = (offerId: string) => {
    setSelectedOfferIds((current) => {
      const isSelected = current.includes(offerId);

      if (isSelected) {
        setOfferQuantities((quantities) => {
          const next = {...quantities};
          delete next[offerId];
          return next;
        });

        return current.filter((id) => id !== offerId);
      }

      setOfferQuantities((quantities) => ({
        ...quantities,
        [offerId]: quantities[offerId] ?? "1",
      }));

      return [...current, offerId];
    });
  };

  const updateOfferQuantity = (offerId: string, value: string) => {
    setOfferQuantities((current) => ({
      ...current,
      [offerId]: value,
    }));
  };

  const submit = async () => {
    if (selectedOffers.length === 0 || hasInvalidQuantity) return;

    const result = await createDraftDeal({
      offers: [
        { offerID: targetOffer.id, quantity: 1 },
        ...selectedOffers.map((offer) => ({
          offerID: offer.id,
          quantity: quantitiesByOfferId[offer.id] ?? 1,
        })),
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
          Выберите одно или несколько своих объявлений для обмена и укажите количество:
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
                {data.offers.map((offer) => {
                  const isSelected = selectedOfferIds.includes(offer.id);
                  const quantityError = isSelected && quantitiesByOfferId[offer.id] === null;

                  return (
                    <Box
                      key={offer.id}
                      display="flex"
                      alignItems="flex-start"
                      justifyContent="space-between"
                      gap={2}
                      py={1}
                    >
                      <FormControlLabel
                        control={
                          <Checkbox checked={isSelected} onChange={() => toggleSelectedOffer(offer.id)} />
                        }
                        label={
                          <Box>
                            <Typography variant="body1" fontWeight={500}>
                              {offer.name}
                            </Typography>
                            {offer.description && (
                              <Typography variant="body2" color="text.secondary">
                                {offer.description}
                              </Typography>
                            )}
                          </Box>
                        }
                        sx={{ alignItems: "flex-start", flex: 1, "& .MuiCheckbox-root": { mt: 0.5 } }}
                      />

                      {isSelected && (
                        <TextField
                          label="Количество"
                          type="number"
                          size="small"
                          value={offerQuantities[offer.id] ?? "1"}
                          onChange={(e) => updateOfferQuantity(offer.id, e.target.value)}
                          error={quantityError}
                          helperText={quantityError ? "Минимум 1" : " "}
                          slotProps={{ htmlInput: { min: 1, step: 1 } }}
                          sx={{ width: 140 }}
                        />
                      )}
                    </Box>
                  );
                })}
              </Box>
            </FormControl>
          )
        )}

        {createError && (
          <Alert severity="error" sx={{ mt: 2 }}>
            {getCreateDraftDealErrorMessage(createError)}
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
          disabled={selectedOfferIds.length === 0 || isCreating || isLoading || !!error || hasInvalidQuantity}
        >
          {isCreating ? <CircularProgress size={20} color="inherit" /> : "Создать черновик"}
        </Button>
      </DialogActions>
    </Dialog>
  );
}

export default RespondToOfferModal;
