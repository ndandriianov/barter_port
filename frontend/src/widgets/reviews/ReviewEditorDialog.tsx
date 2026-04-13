import { useState } from "react";
import {
  Alert,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Rating,
  Stack,
  TextField,
  Typography,
} from "@mui/material";

interface ReviewEditorDialogProps {
  open: boolean;
  title: string;
  submitLabel: string;
  initialRating?: number;
  initialComment?: string;
  isLoading?: boolean;
  errorMessage?: string | null;
  onClose: () => void;
  onSubmit: (value: { rating: number; comment: string }) => Promise<void>;
}

interface ReviewEditorDialogFormProps {
  submitLabel: string;
  initialRating: number;
  initialComment: string;
  isLoading: boolean;
  errorMessage?: string | null;
  onClose: () => void;
  onSubmit: (value: { rating: number; comment: string }) => Promise<void>;
}

function ReviewEditorDialogForm({
  submitLabel,
  initialRating,
  initialComment,
  isLoading,
  errorMessage,
  onClose,
  onSubmit,
}: ReviewEditorDialogFormProps) {
  const [rating, setRating] = useState(initialRating);
  const [comment, setComment] = useState(initialComment);

  const handleSubmit = async () => {
    if (!rating) {
      return;
    }

    await onSubmit({ rating, comment });
  };

  return (
    <>
      <DialogContent>
        <Stack spacing={2} sx={{ pt: 1 }}>
          <div>
            <Typography variant="body2" color="text.secondary" mb={0.75}>
              Оценка
            </Typography>
            <Rating
              value={rating}
              onChange={(_event, value) => setRating(value ?? 0)}
              max={5}
              size="large"
            />
          </div>

          <TextField
            label="Комментарий"
            value={comment}
            onChange={(event) => setComment(event.target.value)}
            multiline
            minRows={4}
            fullWidth
            placeholder="Что было хорошо или что можно улучшить"
          />

          {errorMessage && <Alert severity="error">{errorMessage}</Alert>}
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose} disabled={isLoading}>
          Отмена
        </Button>
        <Button onClick={() => void handleSubmit()} variant="contained" disabled={isLoading || rating < 1}>
          {submitLabel}
        </Button>
      </DialogActions>
    </>
  );
}

function ReviewEditorDialog({
  open,
  title,
  submitLabel,
  initialRating = 5,
  initialComment = "",
  isLoading = false,
  errorMessage,
  onClose,
  onSubmit,
}: ReviewEditorDialogProps) {
  return (
    <Dialog open={open} onClose={isLoading ? undefined : onClose} fullWidth maxWidth="sm">
      <DialogTitle>{title}</DialogTitle>
      <ReviewEditorDialogForm
        key={JSON.stringify([open, initialRating, initialComment])}
        submitLabel={submitLabel}
        initialRating={initialRating}
        initialComment={initialComment}
        isLoading={isLoading}
        errorMessage={errorMessage}
        onClose={onClose}
        onSubmit={onSubmit}
      />
    </Dialog>
  );
}

export default ReviewEditorDialog;
