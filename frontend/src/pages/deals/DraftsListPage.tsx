import { Box, Typography } from "@mui/material";
import DraftsList from "@/widgets/deals/DraftsList";

function DraftsListPage() {
  return (
    <Box>
      <Typography variant="h4" fontWeight={700} mb={3}>
        Мои черновые договоры
      </Typography>
      <DraftsList />
    </Box>
  );
}

export default DraftsListPage;
