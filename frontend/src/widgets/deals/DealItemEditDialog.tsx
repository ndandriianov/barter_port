import { useEffect, useMemo, useRef, useState, type ChangeEvent } from "react";
import {
  Alert,
  Box,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormHelperText,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import type { FetchBaseQueryError } from "@reduxjs/toolkit/query";
import type { SerializedError } from "@reduxjs/toolkit";
import dealsApi from "@/features/deals/api/dealsApi";
import type { Item, UpdateDealItemRequest } from "@/features/deals/model/types";
import { getStatusCode } from "@/shared/utils/getStatusCode";

const MAX_PHOTO_COUNT = 10;
const MAX_PHOTO_SIZE = 5 * 1024 * 1024;

function getDealItemErrorMessage(
  error: FetchBaseQueryError | SerializedError | undefined,
  fallback: string,
): string {
  const code = getStatusCode(error);

  switch (code) {
    case 400:
      return "Некорректные данные позиции";
    case 403:
      return "Редактирование позиции недоступно на данном этапе";
    case 404:
      return "Позиция сделки не найдена";
    default:
      return fallback;
  }
}

interface DealItemEditDialogProps {
  item: Item;
  dealId: string;
  open: boolean;
  onClose: () => void;
}

function DealItemEditDialog({ item, dealId, open, onClose }: DealItemEditDialogProps) {
  const [name, setName] = useState(item.name);
  const [description, setDescription] = useState(item.description);
  const [quantity, setQuantity] = useState(String(item.quantity));
  const [photos, setPhotos] = useState<File[]>([]);
  const [deletedPhotoIds, setDeletedPhotoIds] = useState<string[]>([]);
  const [photoError, setPhotoError] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const [updateDealItem, updateState] = dealsApi.useUpdateDealItemMutation();

  const existingPhotos = useMemo(
    () => item.photoUrls.map((url, index) => ({ id: item.photoIds[index] ?? `${item.id}-${index}`, url })),
    [item],
  );
  const photoPreviewUrls = useMemo(
    () => photos.map((photo) => ({ file: photo, url: URL.createObjectURL(photo) })),
    [photos],
  );
  const activeExistingPhotoCount = existingPhotos.filter((photo) => !deletedPhotoIds.includes(photo.id)).length;

  useEffect(() => {
    if (!open) {
      return;
    }

    setName(item.name);
    setDescription(item.description);
    setQuantity(String(item.quantity));
    setPhotos([]);
    setDeletedPhotoIds([]);
    setPhotoError(null);
  }, [item, open]);

  useEffect(() => {
    return () => {
      for (const { url } of photoPreviewUrls) {
        URL.revokeObjectURL(url);
      }
    };
  }, [photoPreviewUrls]);

  const parsedQuantity = parseInt(quantity, 10);
  const hasValidQuantity = Number.isInteger(parsedQuantity) && parsedQuantity >= 1;
  const quantityError = quantity === "" || !hasValidQuantity;
  const trimmedName = name.trim();
  const hasChanges =
    trimmedName !== item.name ||
    description !== item.description ||
    (hasValidQuantity && parsedQuantity !== item.quantity) ||
    deletedPhotoIds.length > 0 ||
    photos.length > 0;

  const handleSelectPhotos = () => {
    fileInputRef.current?.click();
  };

  const handlePhotosChange = (event: ChangeEvent<HTMLInputElement>) => {
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

  const handleRemoveNewPhoto = (index: number) => {
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

  const handleSave = async () => {
    if (!trimmedName || !hasValidQuantity) {
      return;
    }

    const body: UpdateDealItemRequest = {};

    if (trimmedName !== item.name) {
      body.name = trimmedName;
    }
    if (description !== item.description) {
      body.description = description;
    }
    if (parsedQuantity !== item.quantity) {
      body.quantity = parsedQuantity;
    }
    if (deletedPhotoIds.length > 0) {
      body.deletePhotoIds = deletedPhotoIds;
    }
    if (photos.length > 0) {
      body.photos = photos;
    }

    if (Object.keys(body).length === 0) {
      onClose();
      return;
    }

    await updateDealItem({ dealId, itemId: item.id, body }).unwrap();
    onClose();
  };

  return (
    <Dialog open={open} onClose={updateState.isLoading ? undefined : onClose} fullWidth maxWidth="sm">
      <DialogTitle>Редактировать позицию</DialogTitle>
      <DialogContent sx={{ display: "flex", flexDirection: "column", gap: 2, pt: 2 }}>
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
          value={name}
          onChange={(event) => setName(event.target.value)}
          fullWidth
          size="small"
        />

        <TextField
          label="Описание"
          value={description}
          onChange={(event) => setDescription(event.target.value)}
          fullWidth
          size="small"
          multiline
          minRows={2}
        />

        <TextField
          label="Количество"
          value={quantity}
          onChange={(event) => setQuantity(event.target.value)}
          type="number"
          inputProps={{ min: 1 }}
          fullWidth
          size="small"
          error={quantityError}
          helperText={quantityError ? "Минимум 1" : undefined}
        />

        <Box>
          <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap" useFlexGap>
            <Button variant="outlined" onClick={handleSelectPhotos} disabled={updateState.isLoading}>
              Добавить фото
            </Button>
            <Typography variant="body2" color="text.secondary">
              До {MAX_PHOTO_COUNT} файлов, до 5 МБ каждый
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Новые фото будут добавлены в конец списка.
            </Typography>
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
                      alt={trimmedName || item.name}
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
                      onClick={() => handleRemoveNewPhoto(index)}
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

        {updateState.error && (
          <Alert severity="error">
            {getDealItemErrorMessage(updateState.error, "Не удалось обновить позицию сделки")}
          </Alert>
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose} disabled={updateState.isLoading}>Отмена</Button>
        <Button
          onClick={() => void handleSave()}
          variant="contained"
          disabled={updateState.isLoading || !hasValidQuantity || !trimmedName || !!photoError || !hasChanges}
        >
          Сохранить
        </Button>
      </DialogActions>
    </Dialog>
  );
}

export default DealItemEditDialog;
