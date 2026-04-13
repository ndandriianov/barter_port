import { useEffect, useMemo, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  Alert,
  Box,
  Button,
  CircularProgress,
  FormControl,
  FormHelperText,
  InputLabel,
  MenuItem,
  Select,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import offersApi from "@/features/offers/api/offersApi";
import type { OfferAction, OfferType } from "@/features/offers/model/types";
import { getErrorMessage } from "@/shared/utils/getErrorMessage.ts";

const MAX_PHOTO_COUNT = 10;
const MAX_PHOTO_SIZE = 5 * 1024 * 1024;

function CreateOfferForm() {
  const [createOffer, { isLoading, error }] = offersApi.useCreateOfferMutation();
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [action, setAction] = useState<OfferAction>("give");
  const [type, setType] = useState<OfferType>("good");
  const [photos, setPhotos] = useState<File[]>([]);
  const [photoError, setPhotoError] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const navigate = useNavigate();
  const photoPreviewUrls = useMemo(
    () => photos.map((photo) => ({ file: photo, url: URL.createObjectURL(photo) })),
    [photos],
  );

  useEffect(() => {
    return () => {
      for (const { url } of photoPreviewUrls) {
        URL.revokeObjectURL(url);
      }
    };
  }, [photoPreviewUrls]);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    await createOffer({ name, description, action, type, photos }).unwrap();
    navigate("/offers");
  };

  const handleSelectPhotos = () => {
    fileInputRef.current?.click();
  };

  const handlePhotosChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const selectedFiles = Array.from(event.target.files ?? []);
    event.target.value = "";

    if (selectedFiles.length === 0) {
      return;
    }

    const nextPhotos = [...photos];

    for (const file of selectedFiles) {
      if (!file.type.startsWith("image/")) {
        setPhotoError("Можно загружать только изображения.");
        return;
      }

      if (file.size > MAX_PHOTO_SIZE) {
        setPhotoError("Размер одного фото не должен превышать 5 МБ.");
        return;
      }

      nextPhotos.push(file);
    }

    if (nextPhotos.length > MAX_PHOTO_COUNT) {
      setPhotoError("Можно загрузить не больше 10 фото.");
      return;
    }

    setPhotos(nextPhotos);
    setPhotoError(null);
  };

  const handleRemovePhoto = (index: number) => {
    setPhotos((current) => current.filter((_, currentIndex) => currentIndex !== index));
    setPhotoError(null);
  };

  const errorMessage = getErrorMessage(error) ?? "Ошибка при создании объявления";

  return (
    <Box component="form" onSubmit={submit} noValidate display="flex" flexDirection="column" gap={2}>
      <input
        ref={fileInputRef}
        type="file"
        accept="image/*"
        multiple
        hidden
        onChange={handlePhotosChange}
      />

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

      <Box>
        <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap" useFlexGap>
          <Button variant="outlined" onClick={handleSelectPhotos} disabled={isLoading}>
            Добавить фото
          </Button>
          <Typography variant="body2" color="text.secondary">
            До {MAX_PHOTO_COUNT} файлов, до 5 МБ каждый
          </Typography>
        </Stack>

        {photoPreviewUrls.length > 0 && (
          <Stack direction="row" spacing={1} mt={2} flexWrap="wrap" useFlexGap>
            {photoPreviewUrls.map(({ file, url }, index) => (
              <Box
                key={`${file.name}-${index}`}
                sx={{
                  width: 120,
                  border: 1,
                  borderColor: "divider",
                  borderRadius: 2,
                  overflow: "hidden",
                  bgcolor: "background.paper",
                }}
              >
                <Box
                  component="img"
                  src={url}
                  alt={file.name}
                  sx={{ width: "100%", height: 96, objectFit: "cover", display: "block" }}
                />
                <Box p={1}>
                  <Typography variant="caption" display="block" noWrap title={file.name}>
                    {file.name}
                  </Typography>
                  <Button
                    type="button"
                    size="small"
                    color="inherit"
                    onClick={() => handleRemovePhoto(index)}
                    sx={{ mt: 0.5 }}
                  >
                    Удалить
                  </Button>
                </Box>
              </Box>
            ))}
          </Stack>
        )}

        {photoError && <FormHelperText error>{photoError}</FormHelperText>}
      </Box>

      {error && <Alert severity="error">{errorMessage}</Alert>}

      <Button type="submit" variant="contained" size="large" disabled={isLoading}>
        {isLoading ? <CircularProgress size={24} color="inherit" /> : "Создать объявление"}
      </Button>
    </Box>
  );
}

export default CreateOfferForm;
