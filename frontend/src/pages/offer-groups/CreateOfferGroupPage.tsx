import { Box, Typography } from "@mui/material";
import CreateOfferGroupForm from "@/widgets/offer-groups/CreateOfferGroupForm.tsx";

function CreateOfferGroupPage() {
  return (
    <Box maxWidth={900} mx="auto">
      <Typography variant="h4" fontWeight={700} mb={3}>
        Новое композитное объявление
      </Typography>
      <Typography variant="body1" color="text.secondary" mb={3}>
        Название можно задать вручную или оставить пустым, чтобы сервер собрал его автоматически из выбранных offers.
      </Typography>
      <CreateOfferGroupForm />
    </Box>
  );
}

export default CreateOfferGroupPage;
