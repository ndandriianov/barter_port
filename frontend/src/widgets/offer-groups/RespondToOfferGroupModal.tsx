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
  InputLabel,
  MenuItem,
  Radio,
  RadioGroup,
  Select,
  TextField,
  Typography,
} from "@mui/material";
import offerGroupsApi from "@/features/offer-groups/api/offerGroupsApi.ts";
import offersApi from "@/features/offers/api/offersApi.ts";
import type { OfferAction } from "@/features/offers/model/types.ts";
import type { OfferGroup } from "@/features/offer-groups/model/types.ts";
import { getOfferGroupUniformAction } from "@/features/offer-groups/model/utils.ts";
import { getErrorMessage } from "@/shared/utils/getErrorMessage.ts";

interface RespondToOfferGroupModalProps {
  offerGroup: OfferGroup;
  isOpen: boolean;
  onClose: () => void;
}

const actionLabels: Record<OfferAction, string> = {
  give: "Отдаю",
  take: "Ищу",
};

function RespondToOfferGroupModal({
  offerGroup,
  isOpen,
  onClose,
}: RespondToOfferGroupModalProps) {
  const navigate = useNavigate();
  const [selectedByUnitId, setSelectedByUnitId] = useState<Record<string, string>>({});
  const [responderOfferId, setResponderOfferId] = useState("");
  const [draftName, setDraftName] = useState("");
  const [draftDescription, setDraftDescription] = useState("");

  const [createDraftFromOfferGroup, { isLoading, error }] =
    offerGroupsApi.useCreateDraftFromOfferGroupMutation();
  const {
    data: myOffersData,
    isLoading: isLoadingMyOffers,
    error: myOffersError,
  } = offersApi.useGetOffersQuery(
    { sort: "ByTime", my: true, cursor_limit: 100 },
    { skip: !isOpen },
  );

  const uniformAction = useMemo(() => getOfferGroupUniformAction(offerGroup), [offerGroup]);
  const requiredResponderAction = uniformAction;
  const isResponderOfferRequired = !!uniformAction;

  const responderOffers = useMemo(() => {
    const offers = myOffersData?.offers ?? [];
    if (!requiredResponderAction) {
      return offers;
    }

    return offers.filter((offer) => offer.action === requiredResponderAction);
  }, [myOffersData?.offers, requiredResponderAction]);

  const missingSelection = useMemo(
    () => offerGroup.units.some((unit) => !selectedByUnitId[unit.id]),
    [offerGroup.units, selectedByUnitId],
  );

  const closeModal = () => {
    setSelectedByUnitId({});
    setResponderOfferId("");
    setDraftName("");
    setDraftDescription("");
    onClose();
  };

  const handleSelect = (unitId: string, offerId: string) => {
    setSelectedByUnitId((current) => ({
      ...current,
      [unitId]: offerId,
    }));
  };

  const submit = async () => {
    if (missingSelection || (isResponderOfferRequired && !responderOfferId)) {
      return;
    }

    const result = await createDraftFromOfferGroup({
      offerGroupId: offerGroup.id,
      body: {
        selectedOfferIds: offerGroup.units.map((unit) => selectedByUnitId[unit.id]),
        responderOfferId: responderOfferId || undefined,
        name: draftName.trim() || undefined,
        description: draftDescription.trim() || undefined,
      },
    }).unwrap();

    closeModal();
    navigate(`/deals/drafts/${result.id}`);
  };

  return (
    <Dialog open={isOpen} onClose={closeModal} maxWidth="md" fullWidth>
      <DialogTitle>Откликнуться на композитное объявление</DialogTitle>

      <DialogContent dividers>
        <Typography variant="body2" color="text.secondary" mb={3}>
          Выберите по одному offer из каждого unit. После этого будет создан обычный draft deal.
        </Typography>

        <Alert severity={isResponderOfferRequired ? "info" : "success"} sx={{ mb: 3 }}>
          {isResponderOfferRequired && requiredResponderAction
            ? `Во всех unit группы одинаковый action. Нужно приложить свой offer с action "${actionLabels[requiredResponderAction]}".`
            : "В группе встречаются разные action. Свой offer можно приложить, но это необязательно."}
        </Alert>

        <Box display="flex" flexDirection="column" gap={3}>
          {offerGroup.units.map((unit, index) => (
            <FormControl key={unit.id}>
              <Typography variant="subtitle1" fontWeight={700} mb={1}>
                Unit {index + 1}
              </Typography>
              <RadioGroup
                value={selectedByUnitId[unit.id] ?? ""}
                onChange={(event) => handleSelect(unit.id, event.target.value)}
              >
                {unit.offers.map((offer) => (
                  <Box
                    key={offer.id}
                    sx={{
                      border: "1px solid",
                      borderColor: selectedByUnitId[unit.id] === offer.id ? "primary.main" : "divider",
                      borderRadius: 2,
                      px: 2,
                      py: 1.5,
                      mb: 1,
                    }}
                  >
                    <FormControlLabel
                      value={offer.id}
                      control={<Radio />}
                      label={
                        <Box>
                          <Typography variant="body1" fontWeight={600}>
                            {offer.name}
                          </Typography>
                          <Typography variant="body2" color="text.secondary">
                            {offer.description}
                          </Typography>
                        </Box>
                      }
                    />
                  </Box>
                ))}
              </RadioGroup>
            </FormControl>
          ))}
        </Box>

        <Box mt={3}>
          {isLoadingMyOffers && (
            <Box display="flex" justifyContent="center" py={2}>
              <CircularProgress size={24} />
            </Box>
          )}

          {myOffersError && (
            <Alert severity="error">Не удалось загрузить ваши offers для отклика</Alert>
          )}

          {!isLoadingMyOffers && !myOffersError && (
            <FormControl fullWidth required={isResponderOfferRequired}>
              <InputLabel>
                {isResponderOfferRequired ? "Ваш offer для отклика" : "Ваш offer для отклика (необязательно)"}
              </InputLabel>
              <Select
                value={responderOfferId}
                label={isResponderOfferRequired ? "Ваш offer для отклика" : "Ваш offer для отклика (необязательно)"}
                onChange={(event) => setResponderOfferId(event.target.value)}
              >
                {!isResponderOfferRequired && <MenuItem value="">Не прикреплять</MenuItem>}
                {responderOffers.map((offer) => (
                  <MenuItem key={offer.id} value={offer.id}>
                    {offer.name} · {actionLabels[offer.action]}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
          )}

          {!isLoadingMyOffers && !myOffersError && responderOffers.length === 0 && (
            <Alert severity={isResponderOfferRequired ? "warning" : "info"} sx={{ mt: 2 }}>
              {isResponderOfferRequired && requiredResponderAction
                ? `У вас нет offer с action "${actionLabels[requiredResponderAction]}". Без него отклик на эту группу невозможен.`
                : "У вас пока нет offer, который можно дополнительно приложить к отклику."}
            </Alert>
          )}
        </Box>

        <Box mt={3} display="flex" flexDirection="column" gap={2}>
          <TextField
            label="Название draft deal"
            value={draftName}
            onChange={(event) => setDraftName(event.target.value)}
            helperText="Необязательно. Если оставить пустым, сервер сгенерирует имя автоматически."
          />
          <TextField
            label="Описание draft deal"
            multiline
            minRows={2}
            value={draftDescription}
            onChange={(event) => setDraftDescription(event.target.value)}
          />
        </Box>

        {error && (
          <Alert severity="error" sx={{ mt: 2 }}>
            {getErrorMessage(error) ?? "Не удалось создать черновик по композитному объявлению"}
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
          disabled={
            missingSelection ||
            isLoading ||
            isLoadingMyOffers ||
            (isResponderOfferRequired && (!responderOfferId || responderOffers.length === 0))
          }
        >
          {isLoading ? <CircularProgress size={20} color="inherit" /> : "Создать черновик"}
        </Button>
      </DialogActions>
    </Dialog>
  );
}

export default RespondToOfferGroupModal;
