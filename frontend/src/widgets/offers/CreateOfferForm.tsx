import { useEffect, useMemo, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  Alert,
  Box,
  Button,
  Chip,
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
import type { Offer, OfferAction, OfferType } from "@/features/offers/model/types";
import { normalizeOfferTags, parseOfferTagsInput } from "@/features/offers/model/tagUtils.ts";
import { getErrorMessage } from "@/shared/utils/getErrorMessage.ts";

const MAX_PHOTO_COUNT = 10;
const MAX_PHOTO_SIZE = 5 * 1024 * 1024;

interface CreateOfferFormProps {
  mode?: "create" | "edit";
  offer?: Offer;
}

function CreateOfferForm({ mode = "create", offer }: CreateOfferFormProps) {
  const [createOffer, createState] = offersApi.useCreateOfferMutation();
  const [updateOffer, updateState] = offersApi.useUpdateOfferMutation();
  const isEditMode = mode === "edit";
  const mutationState = isEditMode ? updateState : createState;
  const [name, setName] = useState(offer?.name ?? "");
  const [description, setDescription] = useState(offer?.description ?? "");
  const [action, setAction] = useState<OfferAction>(offer?.action ?? "give");
  const [type, setType] = useState<OfferType>(offer?.type ?? "good");
  const [tagsInput, setTagsInput] = useState((offer?.tags ?? []).join(", "));
  const [photos, setPhotos] = useState<File[]>([]);
  const [deletedPhotoIds, setDeletedPhotoIds] = useState<string[]>([]);
  const [photoError, setPhotoError] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const navigate = useNavigate();
  const { data: existingTags = [] } = offersApi.useListTagsQuery();
  const existingPhotos = useMemo(
    () => (offer ? offer.photoUrls.map((url, index) => ({ id: offer.photoIds[index] ?? `${offer.id}-${index}`, url })) : []),
    [offer],
  );
  const photoPreviewUrls = useMemo(
    () => photos.map((photo) => ({ file: photo, url: URL.createObjectURL(photo) })),
    [photos],
  );
  const activeExistingPhotoCount = existingPhotos.filter((photo) => !deletedPhotoIds.includes(photo.id)).length;
  const parsedTags = useMemo(() => parseOfferTagsInput(tagsInput), [tagsInput]);
  const suggestedTags = useMemo(
    () => existingTags.filter((tag) => !parsedTags.includes(tag)).slice(0, 12),
    [existingTags, parsedTags],
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
    if (isEditMode) {
      if (!offer) {
        return;
      }
      await updateOffer({
        offerId: offer.id,
        body: {
          name,
          description,
          action,
          type,
          tags: parsedTags,
          deletePhotoIds: deletedPhotoIds.length > 0 ? deletedPhotoIds : undefined,
          photos,
        },
      }).unwrap();
      navigate(`/offers/${offer.id}`);
      return;
    }

    await createOffer({ name, description, action, type, tags: parsedTags, photos }).unwrap();
    navigate("/offers");
  };

  const handleAddSuggestedTag = (tag: string) => {
    setTagsInput(normalizeOfferTags([...parsedTags, tag]).join(", "));
  };

  const handleRemoveTag = (tagToRemove: string) => {
    setTagsInput(normalizeOfferTags(parsedTags.filter((tag) => tag !== tagToRemove)).join(", "));
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

    if (activeExistingPhotoCount + nextPhotos.length > MAX_PHOTO_COUNT) {
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

  const handleToggleExistingPhotoDeletion = (photoId: string) => {
    setDeletedPhotoIds((current) => (
      current.includes(photoId)
        ? current.filter((currentId) => currentId !== photoId)
        : [...current, photoId]
    ));
    setPhotoError(null);
  };

  const errorMessage = getErrorMessage(mutationState.error) ?? (
    isEditMode ? "Ошибка при обновлении объявления" : "Ошибка при создании объявления"
  );

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

      <Box>
        <TextField
          label="Теги"
          fullWidth
          value={tagsInput}
          onChange={(e) => setTagsInput(e.target.value)}
          helperText="До 5 тегов. Разделяйте запятыми. Сервер нормализует теги в lowercase."
        />

        {parsedTags.length > 0 && (
          <Box display="flex" gap={1} flexWrap="wrap" mt={1.5}>
            {parsedTags.map((tag) => (
              <Chip key={tag} label={tag} size="small" clickable onClick={() => handleRemoveTag(tag)} />
            ))}
          </Box>
        )}

        {suggestedTags.length > 0 && (
          <Box mt={1.5}>
            <Typography variant="caption" color="text.secondary" display="block" mb={0.75}>
              Популярные теги
            </Typography>
            <Box display="flex" gap={1} flexWrap="wrap">
              {suggestedTags.map((tag) => (
                <Chip
                  key={tag}
                  label={tag}
                  size="small"
                  variant="outlined"
                  clickable
                  onClick={() => handleAddSuggestedTag(tag)}
                />
              ))}
            </Box>
          </Box>
        )}
      </Box>

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
          <Button variant="outlined" onClick={handleSelectPhotos} disabled={mutationState.isLoading}>
            Добавить фото
          </Button>
          <Typography variant="body2" color="text.secondary">
            До {MAX_PHOTO_COUNT} файлов, до 5 МБ каждый
          </Typography>
          {isEditMode && (
            <Typography variant="body2" color="text.secondary">
              Новые фото будут добавлены в конец списка.
            </Typography>
          )}
        </Stack>

        {(existingPhotos.length > 0 || photoPreviewUrls.length > 0) && (
          <Stack direction="row" spacing={1} mt={2} flexWrap="wrap" useFlexGap>
            {existingPhotos.map((photo) => {
              const isMarkedForDeletion = deletedPhotoIds.includes(photo.id);

              return (
                <Box
                  key={photo.id}
                  sx={{
                    width: 120,
                    border: 1,
                    borderColor: isMarkedForDeletion ? "error.main" : "divider",
                    borderRadius: 2,
                    overflow: "hidden",
                    bgcolor: "background.paper",
                    opacity: isMarkedForDeletion ? 0.55 : 1,
                  }}
                >
                  <Box
                    component="img"
                    src={photo.url}
                    alt={name || "offer photo"}
                    sx={{ width: "100%", height: 96, objectFit: "cover", display: "block" }}
                  />
                  <Box p={1}>
                    <Typography variant="caption" display="block" noWrap>
                      {isMarkedForDeletion ? "Будет удалено" : "Текущее фото"}
                    </Typography>
                    <Button
                      type="button"
                      size="small"
                      color={isMarkedForDeletion ? "inherit" : "error"}
                      onClick={() => handleToggleExistingPhotoDeletion(photo.id)}
                      sx={{ mt: 0.5 }}
                    >
                      {isMarkedForDeletion ? "Оставить" : "Удалить"}
                    </Button>
                  </Box>
                </Box>
              );
            })}

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

      {mutationState.error && <Alert severity="error">{errorMessage}</Alert>}

      <Button type="submit" variant="contained" size="large" disabled={mutationState.isLoading}>
        {mutationState.isLoading ? (
          <CircularProgress size={24} color="inherit" />
        ) : isEditMode ? (
          "Сохранить изменения"
        ) : (
          "Создать объявление"
        )}
      </Button>
    </Box>
  );
}

export default CreateOfferForm;
