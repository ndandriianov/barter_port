import { useEffect, useMemo, useState } from "react";
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
import type { Offer, SuitableOffersListItem } from "@/features/offers/model/types";
import { getCreateDraftDealErrorMessage } from "@/shared/utils/getCreateDraftErrorMessage.ts";

const RANGED_TIMEOUT_MS = Number(import.meta.env.VITE_LLM_RANGED_TIMEOUT_MS ?? 5000);

interface RespondToOfferModalProps {
  targetOffer: Offer;
  isOpen: boolean;
  onClose: () => void;
}

type AnyOfferItem = SuitableOffersListItem & { comment?: string };

function RespondToOfferModal({ targetOffer, isOpen, onClose }: RespondToOfferModalProps) {
  const [selectedOfferIds, setSelectedOfferIds] = useState<string[]>([]);
  const [offerQuantities, setOfferQuantities] = useState<Record<string, string>>({});
  const [useFallback, setUseFallback] = useState(false);
  const navigate = useNavigate();

  const closeModal = () => {
    setSelectedOfferIds([]);
    setOfferQuantities({});
    onClose();
  };

  // Сброс fallback при закрытии модала, чтобы при следующем открытии снова пробовать ranged
  useEffect(() => {
    if (!isOpen) setUseFallback(false);
  }, [isOpen]);

  const rangedQuery = offersApi.useListSuitableOffersRangedQuery(
    targetOffer.id,
    { skip: !isOpen || useFallback },
  );

  const fallbackQuery = offersApi.useListSuitableOffersQuery(
    targetOffer.id,
    { skip: !isOpen || !useFallback },
  );

  // Таймаут: если ranged не ответил за RANGED_TIMEOUT_MS — переключаемся на fallback
  useEffect(() => {
    if (!isOpen || useFallback || rangedQuery.isSuccess) return;
    const timer = setTimeout(() => setUseFallback(true), RANGED_TIMEOUT_MS);
    return () => clearTimeout(timer);
  }, [isOpen, useFallback, rangedQuery.isSuccess]);

  // 503 от бэкенда (LLM недоступен) — немедленный fallback
  useEffect(() => {
    const err = rangedQuery.error;
    if (err && typeof err === "object" && "status" in err && err.status === 503) {
      setUseFallback(true);
    }
  }, [rangedQuery.error]);

  const { data, isLoading, error } = useFallback ? fallbackQuery : rangedQuery;
  const offers = (data ?? []) as AnyOfferItem[];

  const [createDraftDeal, { isLoading: isCreating, error: createError }] =
    dealsApi.useCreateDraftDealMutation();

  const selectedOffers = useMemo(
    () => offers.filter((entry) => selectedOfferIds.includes(entry.offerId)),
    [offers, selectedOfferIds],
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
          const next = { ...quantities };
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
          offerID: offer.offerId,
          quantity: quantitiesByOfferId[offer.offerId] ?? 1,
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
          offers.length === 0 ? (
            <Typography color="text.secondary">У вас пока нет объявлений для отклика</Typography>
          ) : (
            <FormControl component="fieldset" fullWidth>
              <Box display="flex" flexDirection="column">
                {offers.map((offer) => {
                  const isSelected = selectedOfferIds.includes(offer.offerId);
                  const quantityError = isSelected && quantitiesByOfferId[offer.offerId] === null;

                  return (
                    <Box
                      key={offer.offerId}
                      display="flex"
                      alignItems="flex-start"
                      justifyContent="space-between"
                      gap={2}
                      py={1}
                    >
                      <FormControlLabel
                        control={
                          <Checkbox checked={isSelected} onChange={() => toggleSelectedOffer(offer.offerId)} />
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
                            {offer.comment && (
                              <Typography variant="caption" color="primary.main" fontStyle="italic">
                                {offer.comment}
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
                          value={offerQuantities[offer.offerId] ?? "1"}
                          onChange={(e) => updateOfferQuantity(offer.offerId, e.target.value)}
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
