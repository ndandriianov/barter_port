import { Box, Button, ButtonGroup, Stack, Typography } from "@mui/material";
import HistoryOutlinedIcon from "@mui/icons-material/HistoryOutlined";
import PlayCircleOutlineOutlinedIcon from "@mui/icons-material/PlayCircleOutlineOutlined";
import { Link as RouterLink } from "react-router-dom";
import DealsList from "@/widgets/deals/DealsList.tsx";
import type { DealStatus } from "@/features/deals/model/types.ts";
import { appRoutes } from "@/shared/config/appRoutes.ts";

interface DealsStatusBoardPageProps {
  mode: "active" | "history";
}

const activeStatuses: DealStatus[] = ["LookingForParticipants", "Discussion", "Confirmed"];
const historyStatuses: DealStatus[] = ["Completed", "Cancelled", "Failed"];

function DealsStatusBoardPage({ mode }: DealsStatusBoardPageProps) {
  const isActiveMode = mode === "active";

  return (
    <Stack spacing={3}>
      <Box>
        <Typography variant="overline" color={isActiveMode ? "info.main" : "secondary.main"}>
          Сделки / {isActiveMode ? "Активные" : "История"}
        </Typography>
        <Typography variant="h4" fontWeight={800} mb={1}>
          {isActiveMode ? "Активные сделки" : "История сделок"}
        </Typography>
        <Typography variant="body1" color="text.secondary">
          {isActiveMode
            ? "Рабочая зона по текущим сделкам: открытые наборы участников, обсуждение и подтвержденные договоренности."
            : "Финальные состояния сделки собраны отдельно, чтобы не конкурировать с активной работой за внимание."}
        </Typography>
      </Box>

      <ButtonGroup
        variant="text"
        sx={{
          alignSelf: "flex-start",
          bgcolor: "background.paper",
          borderRadius: 999,
          p: 0.75,
          boxShadow: "0 10px 30px rgba(15, 23, 42, 0.08)",
        }}
      >
        <Button
          component={RouterLink}
          to={appRoutes.deals.active}
          variant={isActiveMode ? "contained" : "text"}
          startIcon={<PlayCircleOutlineOutlinedIcon />}
        >
          Активные
        </Button>
        <Button
          component={RouterLink}
          to={appRoutes.deals.history}
          variant={!isActiveMode ? "contained" : "text"}
          startIcon={<HistoryOutlinedIcon />}
        >
          История
        </Button>
      </ButtonGroup>

      <DealsList
        statusFilter="all"
        allowedStatuses={isActiveMode ? activeStatuses : historyStatuses}
        defaultMyOnly
      />
    </Stack>
  );
}

export default DealsStatusBoardPage;
