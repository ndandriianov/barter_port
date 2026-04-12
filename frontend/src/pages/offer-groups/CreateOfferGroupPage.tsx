import { Box, Typography } from "@mui/material";
import CreateOfferGroupForm from "@/widgets/offer-groups/CreateOfferGroupForm.tsx";

function CreateOfferGroupPage() {
  return (
    <Box maxWidth={900} mx="auto">
      <Typography variant="h4" fontWeight={700} mb={3}>
        Новое композитное объявление
      </Typography>
      <CreateOfferGroupForm />
    </Box>
  );
}

export default CreateOfferGroupPage;
