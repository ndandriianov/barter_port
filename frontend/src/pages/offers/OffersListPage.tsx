import { Link as RouterLink } from "react-router-dom";
import { Box, Button, Typography } from "@mui/material";
import AddIcon from "@mui/icons-material/Add";
import OffersList from "@/widgets/offers/OffersList";

function OffersListPage() {
  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4" fontWeight={700}>
          Объявления
        </Typography>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          component={RouterLink}
          to="/offers/create"
        >
          Создать
        </Button>
      </Box>
      <OffersList />
    </Box>
  );
}

export default OffersListPage;
