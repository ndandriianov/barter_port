import { Stack, Typography } from "@mui/material";
import FailureModerationQueue from "@/widgets/deals/FailureModerationQueue.tsx";

function FailureModerationPage() {
  return (
    <Stack spacing={3}>
      <div>
        <Typography variant="overline" color="secondary.main">
          Модерация / Провалы сделок
        </Typography>
        <Typography variant="h4" fontWeight={800} mb={1}>
          Разбор провалов сделок
        </Typography>
      </div>
      <FailureModerationQueue />
    </Stack>
  );
}

export default FailureModerationPage;
