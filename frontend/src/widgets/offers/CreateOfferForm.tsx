import { useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  Alert,
  Box,
  Button,
  CircularProgress,
  FormControl,
  InputLabel,
  MenuItem,
  Select,
  TextField,
} from "@mui/material";
import offersApi from "@/features/offers/api/offersApi";
import type { OfferAction, OfferType } from "@/features/offers/model/types";

function CreateOfferForm() {
  const [createOffer, { isLoading, error }] = offersApi.useCreateOfferMutation();
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [action, setAction] = useState<OfferAction>("give");
  const [type, setType] = useState<OfferType>("good");
  const navigate = useNavigate();

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    await createOffer({ name, description, action, type }).unwrap();
    navigate("/offers");
  };

  return (
    <Box component="form" onSubmit={submit} noValidate display="flex" flexDirection="column" gap={2}>
      <TextField
        label="Название"
        fullWidth
        required
        value={name}
        onChange={(e) => setName(e.target.value)}
      />

      <TextField
        label="Описание"
        fullWidth
        required
        multiline
        minRows={3}
        value={description}
        onChange={(e) => setDescription(e.target.value)}
      />

      <FormControl fullWidth>
        <InputLabel>Тип действия</InputLabel>
        <Select
          value={action}
          label="Тип действия"
          onChange={(e) => setAction(e.target.value as OfferAction)}
        >
          <MenuItem value="give">Отдаю</MenuItem>
          <MenuItem value="take">Ищу</MenuItem>
        </Select>
      </FormControl>

      <FormControl fullWidth>
        <InputLabel>Категория</InputLabel>
        <Select
          value={type}
          label="Категория"
          onChange={(e) => setType(e.target.value as OfferType)}
        >
          <MenuItem value="good">Товар</MenuItem>
          <MenuItem value="service">Услуга</MenuItem>
        </Select>
      </FormControl>

      {error && <Alert severity="error">Ошибка при создании объявления</Alert>}

      <Button type="submit" variant="contained" size="large" disabled={isLoading}>
        {isLoading ? <CircularProgress size={24} color="inherit" /> : "Создать объявление"}
      </Button>
    </Box>
  );
}

export default CreateOfferForm;
