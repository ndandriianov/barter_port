import { useNavigate, useParams } from "react-router-dom";
import { Alert, Box, Button, CircularProgress, Stack, Typography } from "@mui/material";
import CheckCircleOutlineIcon from "@mui/icons-material/CheckCircleOutline";
import CancelOutlinedIcon from "@mui/icons-material/CancelOutlined";
import dealsApi from "@/features/deals/api/dealsApi";
import DraftCard from "@/widgets/deals/DraftCard";

function DraftPage() {
  const { draftId } = useParams<{ draftId: string }>();
  const navigate = useNavigate();

  const { data, isLoading, error } = dealsApi.useGetDraftDealByIdQuery(draftId ?? "", {
    skip: !draftId,
  });
  const [confirmDraftDeal, { isLoading: isConfirming }] = dealsApi.useConfirmDraftDealMutation();
  const [cancelDraftDeal, { isLoading: isCancelling }] = dealsApi.useCancelDraftDealMutation();

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
    await confirmDraftDeal(draftId);
  };

  const onCancel = async () => {
    await cancelDraftDeal(draftId);
  };

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
          disabled={isConfirming || isCancelling}
        >
          Подтвердить участие
        </Button>
        <Button
          variant="outlined"
          color="error"
          startIcon={isCancelling ? <CircularProgress size={18} color="inherit" /> : <CancelOutlinedIcon />}
          onClick={onCancel}
          disabled={isConfirming || isCancelling}
        >
          Отменить участие
        </Button>
      </Stack>
    </Box>
  );
}

export default DraftPage;
