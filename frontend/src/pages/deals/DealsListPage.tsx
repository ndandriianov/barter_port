import { Link as RouterLink } from "react-router-dom";
import { Box, Button, Typography } from "@mui/material";
import DealsList from "@/widgets/deals/DealsList";

function DealsListPage() {
  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3} flexWrap="wrap" gap={1}>
        <Typography variant="h4" fontWeight={700}>
          Сделки
        </Typography>
        <Box display="flex" gap={1}>
          <Button variant="outlined" component={RouterLink} to="/deals/drafts">
            Мои черновики
          </Button>
        </Box>
      </Box>
      <DealsList />
    </Box>
  );
}

export default DealsListPage;
