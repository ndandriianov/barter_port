import { useParams } from "react-router-dom";
import { Alert, Box, CircularProgress, Typography } from "@mui/material";
import dealsApi from "@/features/deals/api/dealsApi";
import DealCard from "@/widgets/deals/DealCard";

function DealPage() {
  const { dealId } = useParams<{ dealId: string }>();

  const { data, isLoading, error } = dealsApi.useGetDealByIdQuery(dealId ?? "", {
    skip: !dealId,
  });

  if (!dealId) return <Alert severity="warning">Сделка не найдена</Alert>;

  if (isLoading) {
    return (
      <Box display="flex" justifyContent="center" py={6}>
        <CircularProgress />
      </Box>
    );
  }

  if (error) return <Alert severity="error">Не удалось загрузить сделку</Alert>;
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
