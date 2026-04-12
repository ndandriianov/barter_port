import { useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  Alert,
  Box,
  Button,
  Checkbox,
  CircularProgress,
  FormControl,
  Grid,
  IconButton,
  InputLabel,
  ListItemText,
  MenuItem,
  Paper,
  Select,
  TextField,
  Typography,
} from "@mui/material";
import AddCircleOutlineIcon from "@mui/icons-material/AddCircleOutline";
import DeleteOutlineIcon from "@mui/icons-material/DeleteOutline";
import offersApi from "@/features/offers/api/offersApi.ts";
import offerGroupsApi from "@/features/offer-groups/api/offerGroupsApi.ts";
import type { OfferAction } from "@/features/offers/model/types.ts";
import { getErrorMessage } from "@/shared/utils/getErrorMessage.ts";

const actionLabels: Record<OfferAction, string> = {
  give: "Отдаю",
  take: "Ищу",
};

function CreateOfferGroupForm() {
  const navigate = useNavigate();
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [units, setUnits] = useState<string[][]>([[]]);

  const { data, isLoading, error } = offersApi.useGetOffersQuery({
    sort: "ByTime",
    my: true,
    cursor_limit: 100,
  });
  const [createOfferGroup, { isLoading: isCreating, error: createError }] =
    offerGroupsApi.useCreateOfferGroupMutation();

  const offers = data?.offers ?? [];

  const selectedInOtherUnits = useMemo(() => {
    return units.map((_, unitIndex) => {
      const otherSelected = new Set<string>();

      units.forEach((unit, index) => {
        if (index === unitIndex) return;
        unit.forEach((offerId) => otherSelected.add(offerId));
      });

      return otherSelected;
    });
  }, [units]);

  const selectedActionByUnit = useMemo(() => {
    return units.map((unit) => {
      const firstOffer = offers.find((offer) => offer.id === unit[0]);
      return firstOffer?.action;
    });
  }, [offers, units]);

  const unitHasMixedActions = useMemo(() => {
    return units.map((unit) => {
      if (unit.length <= 1) {
        return false;
      }

      const actions = new Set(
        unit
          .map((offerId) => offers.find((offer) => offer.id === offerId)?.action)
          .filter((action): action is OfferAction => Boolean(action)),
      );

      return actions.size > 1;
    });
  }, [offers, units]);

  const hasMixedActions = unitHasMixedActions.some(Boolean);

  const updateUnit = (unitIndex: number, nextOfferIds: string[]) => {
    setUnits((current) => current.map((unit, index) => (index === unitIndex ? nextOfferIds : unit)));
  };

  const addUnit = () => {
    setUnits((current) => [...current, []]);
  };

  const removeUnit = (unitIndex: number) => {
    setUnits((current) => current.filter((_, index) => index !== unitIndex));
  };

  const canSubmit =
    name.trim().length > 0 &&
    units.length > 0 &&
    units.every((unit) => unit.length > 0) &&
    offers.length > 0 &&
    !hasMixedActions;

  const submit = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!canSubmit) {
      return;
    }

    const result = await createOfferGroup({
      name: name.trim(),
      description: description.trim() ? description.trim() : undefined,
      units: units.map((unit) => ({
        offers: unit.map((offerId) => ({ offerId })),
      })),
    }).unwrap();

    navigate(`/offer-groups/${result.id}`);
  };

  return (
    <Box component="form" onSubmit={submit} display="flex" flexDirection="column" gap={3}>
      <TextField
        label="Название композитного объявления"
        required
        fullWidth
        value={name}
        onChange={(event) => setName(event.target.value)}
      />

      <TextField
        label="Описание"
        fullWidth
        multiline
        minRows={3}
        value={description}
        onChange={(event) => setDescription(event.target.value)}
        helperText="Опишите общий сценарий обмена и логику выбора вариантов."
      />

      <Box display="flex" justifyContent="space-between" alignItems="center" gap={2} flexWrap="wrap">
        <Box>
          <Typography variant="h6" fontWeight={700}>
            AND-блоки
          </Typography>
          <Typography variant="body2" color="text.secondary">
            В каждом блоке соберите варианты OR из ваших обычных объявлений.
            Внутри одного unit можно смешивать только одинаковый `action`.
          </Typography>
        </Box>
        <Button variant="outlined" startIcon={<AddCircleOutlineIcon />} onClick={addUnit}>
          Добавить unit
        </Button>
      </Box>

      {isLoading && (
        <Box display="flex" justifyContent="center" py={4}>
          <CircularProgress />
        </Box>
      )}

      {error && <Alert severity="error">Не удалось загрузить ваши объявления</Alert>}

      {!isLoading && !error && offers.length === 0 && (
        <Alert severity="warning">
          Сначала создайте обычные объявления. Композитная группа собирается только из уже существующих offer.
        </Alert>
      )}

      {!isLoading && !error && offers.length > 0 && (
        <Grid container spacing={2}>
          {units.map((unit, unitIndex) => (
            <Grid key={`unit-${unitIndex}`} size={{ xs: 12 }}>
              <Paper variant="outlined" sx={{ p: 2.5 }}>
                <Box display="flex" justifyContent="space-between" alignItems="center" gap={2} mb={2}>
                  <Box>
                    <Typography variant="subtitle1" fontWeight={700}>
                      Unit {unitIndex + 1}
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      Пользователь при отклике выберет один вариант из этого блока.
                    </Typography>
                    {selectedActionByUnit[unitIndex] && (
                      <Typography variant="caption" color="primary.main">
                        Action unit: {actionLabels[selectedActionByUnit[unitIndex]]}
                      </Typography>
                    )}
                  </Box>

                  <IconButton
                    color="error"
                    onClick={() => removeUnit(unitIndex)}
                    disabled={units.length === 1}
                  >
                    <DeleteOutlineIcon />
                  </IconButton>
                </Box>

                <FormControl fullWidth required>
                  <InputLabel>Offers в unit</InputLabel>
                  <Select
                    multiple
                    value={unit}
                    label="Offers в unit"
                    onChange={(event) => updateUnit(unitIndex, event.target.value as string[])}
                    renderValue={(selected) =>
                      offers
                        .filter((offer) => selected.includes(offer.id))
                        .map((offer) => offer.name)
                        .join(", ")
                    }
                  >
                    {offers.map((offer) => {
                      const disabled =
                        (selectedInOtherUnits[unitIndex]?.has(offer.id) ?? false) ||
                        (
                          !unit.includes(offer.id) &&
                          !!selectedActionByUnit[unitIndex] &&
                          selectedActionByUnit[unitIndex] !== offer.action
                        );

                      const secondary =
                        selectedInOtherUnits[unitIndex]?.has(offer.id)
                          ? "Уже используется в другом unit"
                          : selectedActionByUnit[unitIndex] && selectedActionByUnit[unitIndex] !== offer.action && !unit.includes(offer.id)
                            ? `Нельзя смешивать с action "${actionLabels[selectedActionByUnit[unitIndex]]}"`
                            : `${actionLabels[offer.action]} · ${offer.description}`;

                      return (
                        <MenuItem key={offer.id} value={offer.id} disabled={disabled}>
                          <Checkbox checked={unit.includes(offer.id)} />
                          <ListItemText
                            primary={offer.name}
                            secondary={secondary}
                          />
                        </MenuItem>
                      );
                    })}
                  </Select>
                </FormControl>

                {unitHasMixedActions[unitIndex] && (
                  <Alert severity="error" sx={{ mt: 2 }}>
                    Внутри одного unit все выбранные offers должны иметь одинаковый action.
                  </Alert>
                )}
              </Paper>
            </Grid>
          ))}
        </Grid>
      )}

      {createError && (
        <Alert severity="error">
          {getErrorMessage(createError) ?? "Не удалось создать композитное объявление"}
        </Alert>
      )}

      <Button type="submit" variant="contained" size="large" disabled={!canSubmit || isCreating}>
        {isCreating ? <CircularProgress size={22} color="inherit" /> : "Создать композитное объявление"}
      </Button>
    </Box>
  );
}

export default CreateOfferGroupForm;
