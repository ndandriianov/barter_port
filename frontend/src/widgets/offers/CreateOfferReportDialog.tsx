import { useState } from "react";
import {
  Alert,
  Box,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import offersApi from "@/features/offers/api/offersApi.ts";
import { getStatusCode } from "@/shared/utils/getStatusCode.ts";
import { getErrorMessage } from "@/shared/utils/getErrorMessage.ts";

interface CreateOfferReportDialogProps {
  offerId: string;
  open: boolean;
  onClose: () => void;
  onSuccess?: () => void;
}

const recommendationItems = [
  "Кратко опишите, что именно в объявлении нарушает правила или вводит в заблуждение.",
  "Добавьте факты и детали, которые помогут модератору быстрее принять решение.",
  "Не используйте оскорбления: жалоба должна описывать проблему, а не эмоции.",
];

function CreateOfferReportDialog({
  offerId,
  open,
  onClose,
  onSuccess,
}: CreateOfferReportDialogProps) {
  const [message, setMessage] = useState("");
  const [createOfferReport, { isLoading, error }] = offersApi.useCreateOfferReportMutation();

  const trimmedMessage = message.trim();

  const handleClose = () => {
    setMessage("");
    onClose();
  };

  const handleSubmit = async () => {
    if (!trimmedMessage) {
      return;
    }

    try {
      await createOfferReport({
        offerId,
        body: { message: trimmedMessage },
      }).unwrap();
      onSuccess?.();
      onClose();
    } catch {
      // Error is rendered below.
    }
  };

  const statusCode = getStatusCode(error);
  const errorMessage =
    statusCode === 409
      ? "Вы уже отправили жалобу по этому объявлению в текущем разбирательстве."
      : statusCode === 403
        ? "Нельзя пожаловаться на собственное объявление."
        : getErrorMessage(error) ?? "Не удалось отправить жалобу.";

  return (
    <Dialog open={open} onClose={isLoading ? undefined : handleClose} fullWidth maxWidth="sm">
      <DialogTitle>Пожаловаться на объявление</DialogTitle>
      <DialogContent>
        <Stack spacing={2} sx={{ pt: 1 }}>
          <Alert severity="info" variant="outlined">
            <Typography variant="body2" fontWeight={600} mb={0.75}>
              Что написать в жалобе
            </Typography>
            <Stack spacing={0.75}>
              {recommendationItems.map((item) => (
                <Typography key={item} variant="body2">
                  {item}
                </Typography>
              ))}
            </Stack>
          </Alert>

          <TextField
            label="Текст жалобы"
            value={message}
            onChange={(event) => setMessage(event.target.value)}
            placeholder="Опишите проблему с объявлением"
            multiline
            minRows={5}
            fullWidth
            helperText={`${trimmedMessage.length} символов`}
          />

          {error && <Alert severity="error">{errorMessage}</Alert>}

          <Box sx={{ color: "text.secondary" }}>
            <Typography variant="caption">
              После отправки жалоба попадет на модерацию. Если по объявлению уже есть активная
              жалоба, ваше сообщение будет добавлено к текущему разбирательству.
            </Typography>
          </Box>
        </Stack>
      </DialogContent>
      <DialogActions sx={{ px: 3, pb: 2.5 }}>
        <Button onClick={handleClose} disabled={isLoading}>
          Отмена
        </Button>
        <Button variant="contained" onClick={handleSubmit} disabled={!trimmedMessage || isLoading}>
          {isLoading ? "Отправка..." : "Отправить жалобу"}
        </Button>
      </DialogActions>
    </Dialog>
  );
}

export default CreateOfferReportDialog;
