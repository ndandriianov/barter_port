import { Link, useParams } from "react-router-dom";
import { Alert, Box, Button, CircularProgress, Typography } from "@mui/material";
import dealsApi from "@/features/deals/api/dealsApi";
import DealCard from "@/widgets/deals/DealCard";
import { getStatusCode } from "@/shared/utils/getStatusCode";

function DealPage() {
  const { dealId } = useParams<{ dealId: string }>();

  const { data, isLoading, error } = dealsApi.useGetDealByIdQuery(dealId ?? "", {
    skip: !dealId,
    pollingInterval: 10_000,
  });

  if (!dealId) return <Alert severity="warning">Сделка не найдена</Alert>;

  if (isLoading) {
    return (
      <Box display="flex" justifyContent="center" py={6}>
        <CircularProgress />
      </Box>
    );
  }

  if (error) {
    const code = getStatusCode(error);
    if (code === 404) {
      return (
        <Box display="flex" flexDirection="column" alignItems="flex-start" gap={2}>
          <Alert severity="warning">Сделка не найдена</Alert>
          <Button component={Link} to="/deals" variant="outlined" size="small">
            К списку сделок
          </Button>
        </Box>
      );
    }
    if (code === 403) return <Alert severity="error">У вас нет доступа к этой сделке</Alert>;
    return <Alert severity="error">Не удалось загрузить сделку. Попробуйте позже</Alert>;
  }
  if (!data) return <Alert severity="warning">Сделка не найдена</Alert>;

  return (
    <Box maxWidth={700} mx="auto">
      <Typography variant="h4" fontWeight={700} mb={3}>
        Детали сделки
      </Typography>
      <DealCard deal={data} />
    </Box>
  );
}

export default DealPage;
