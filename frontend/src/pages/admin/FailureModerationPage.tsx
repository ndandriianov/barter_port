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
        <Typography variant="body1" color="text.secondary">
          Participant-side голосование остаётся внутри сделки, а админская развязка вынесена в отдельную очередь.
        </Typography>
      </div>
      <FailureModerationQueue />
    </Stack>
  );
}

export default FailureModerationPage;
