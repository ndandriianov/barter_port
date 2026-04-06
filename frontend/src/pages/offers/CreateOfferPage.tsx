import { Box, Typography } from "@mui/material";
import CreateOfferForm from "@/widgets/offers/CreateOfferForm";

function CreateOfferPage() {
  return (
    <Box maxWidth={600} mx="auto">
      <Typography variant="h4" fontWeight={700} mb={3}>
        Новое объявление
      </Typography>
      <CreateOfferForm />
    </Box>
  );
}

export default CreateOfferPage;
