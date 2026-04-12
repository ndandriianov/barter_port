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
  TextField,
  Typography,
} from "@mui/material";
import offerGroupsApi from "@/features/offer-groups/api/offerGroupsApi.ts";
import type { OfferGroup } from "@/features/offer-groups/model/types.ts";
import { getErrorMessage } from "@/shared/utils/getErrorMessage.ts";

interface RespondToOfferGroupModalProps {
  offerGroup: OfferGroup;
  isOpen: boolean;
  onClose: () => void;
}

function RespondToOfferGroupModal({
  offerGroup,
  isOpen,
  onClose,
}: RespondToOfferGroupModalProps) {
  const navigate = useNavigate();
  const [selectedByUnitId, setSelectedByUnitId] = useState<Record<string, string>>({});
  const [draftName, setDraftName] = useState("");
  const [draftDescription, setDraftDescription] = useState("");

  const [createDraftFromOfferGroup, { isLoading, error }] =
    offerGroupsApi.useCreateDraftFromOfferGroupMutation();

  const missingSelection = useMemo(
    () => offerGroup.units.some((unit) => !selectedByUnitId[unit.id]),
    [offerGroup.units, selectedByUnitId],
  );

  const closeModal = () => {
    setSelectedByUnitId({});
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
    if (missingSelection) {
      return;
    }

    const result = await createDraftFromOfferGroup({
      offerGroupId: offerGroup.id,
      body: {
        selectedOfferIds: offerGroup.units.map((unit) => selectedByUnitId[unit.id]),
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
        <Button variant="contained" onClick={submit} disabled={missingSelection || isLoading}>
          {isLoading ? <CircularProgress size={20} color="inherit" /> : "Создать черновик"}
        </Button>
      </DialogActions>
    </Dialog>
  );
}

export default RespondToOfferGroupModal;
