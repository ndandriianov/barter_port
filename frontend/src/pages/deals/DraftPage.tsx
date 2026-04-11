import { useNavigate, useParams } from "react-router-dom";
import { Alert, Box, Button, CircularProgress, Stack, Typography } from "@mui/material";
import CheckCircleOutlineIcon from "@mui/icons-material/CheckCircleOutline";
import CancelOutlinedIcon from "@mui/icons-material/CancelOutlined";
import DeleteOutlineIcon from "@mui/icons-material/DeleteOutline";
import dealsApi from "@/features/deals/api/dealsApi";
import usersApi from "@/features/users/api/usersApi.ts";
import DraftCard from "@/widgets/deals/DraftCard";
import { getErrorMessage } from "@/shared/utils/getErrorMessage.ts";

function DraftPage() {
  const { draftId } = useParams<{ draftId: string }>();
  const navigate = useNavigate();

  const { data, isLoading, error } = dealsApi.useGetDraftDealByIdQuery(draftId ?? "", {
    skip: !draftId,
  });
  const { data: me } = usersApi.useGetCurrentUserQuery();
  const [confirmDraftDeal, { isLoading: isConfirming, error: confirmError }] = dealsApi.useConfirmDraftDealMutation();
  const [cancelDraftDeal, { isLoading: isCancelling, error: cancelError }] = dealsApi.useCancelDraftDealMutation();
  const [deleteDraftDeal, { isLoading: isDeleting, error: deleteError }] = dealsApi.useDeleteDraftDealMutation();

  if (!draftId) return <Alert severity="warning">Черновик не найден</Alert>;

  if (isLoading) {
    return (
      <Box display="flex" justifyContent="center" py={6}>
        <CircularProgress />
      </Box>
    );
  }

  if (error) return <Alert severity="error">Не удалось загрузить черновик</Alert>;
  if (!data) return <Alert severity="warning">Черновик не найден</Alert>;

  const onConfirm = async () => {
    try {
      await confirmDraftDeal(draftId).unwrap();
    } catch {
      return;
    }
  };

  const onCancel = async () => {
    try {
      await cancelDraftDeal(draftId).unwrap();
    } catch {
      return;
    }
  };

  const onDelete = async () => {
    try {
      await deleteDraftDeal(draftId).unwrap();
      navigate("/deals/drafts");
    } catch {
      return;
    }
  };

  const actionError = confirmError ?? cancelError ?? deleteError;
  const actionErrorMessage = getErrorMessage(actionError) ?? "Не удалось выполнить действие с черновиком";
  const deleteDraftLabel = me?.id === data.authorId ? "Отменить черновик" : "Отклонить черновик";

  return (
    <Box maxWidth={700} mx="auto">
      <Button
        size="small"
        variant="text"
        onClick={() => window.history.length > 1 ? navigate(-1) : navigate("/deals")}
        sx={{ mb: 2 }}
      >
        ← Назад
      </Button>

      <Typography variant="h4" fontWeight={700} mb={3}>
        Черновой договор
      </Typography>

      <DraftCard draft={data} />

      <Stack direction="row" spacing={2} mt={3}>
        <Button
          variant="contained"
          color="success"
          startIcon={isConfirming ? <CircularProgress size={18} color="inherit" /> : <CheckCircleOutlineIcon />}
          onClick={onConfirm}
          disabled={isConfirming || isCancelling || isDeleting}
        >
          Подтвердить участие
        </Button>
        <Button
          variant="outlined"
          color="error"
          startIcon={isCancelling ? <CircularProgress size={18} color="inherit" /> : <CancelOutlinedIcon />}
          onClick={onCancel}
          disabled={isConfirming || isCancelling || isDeleting}
        >
          Отменить участие
        </Button>
        <Button
          variant="outlined"
          color="warning"
          startIcon={isDeleting ? <CircularProgress size={18} color="inherit" /> : <DeleteOutlineIcon />}
          onClick={onDelete}
          disabled={isConfirming || isCancelling || isDeleting}
        >
          {deleteDraftLabel}
        </Button>
      </Stack>

      {actionError && (
        <Alert severity="error" sx={{ mt: 2 }}>
          {actionErrorMessage}
        </Alert>
      )}
    </Box>
  );
}

export default DraftPage;
