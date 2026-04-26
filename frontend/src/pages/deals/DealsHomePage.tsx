import { useMemo } from "react";
import { Box, Grid, Stack, Typography } from "@mui/material";
import AssignmentTurnedInOutlinedIcon from "@mui/icons-material/AssignmentTurnedInOutlined";
import AutoModeOutlinedIcon from "@mui/icons-material/AutoModeOutlined";
import HistoryOutlinedIcon from "@mui/icons-material/HistoryOutlined";
import PlaylistAddCheckCircleOutlinedIcon from "@mui/icons-material/PlaylistAddCheckCircleOutlined";
import dealsApi from "@/features/deals/api/dealsApi.ts";
import useDealActionQueue from "@/features/deals/model/useDealActionQueue.ts";
import SectionEntryCard from "@/shared/ui/SectionEntryCard.tsx";
import { appRoutes } from "@/shared/config/appRoutes.ts";

function DealsHomePage() {
  const { data: deals = [] } = dealsApi.useGetDealsQuery({ my: true });
  const { totalActionCount, draftCount, pendingReviewCount, joinRequestCount } = useDealActionQueue();

  const activeCount = useMemo(
    () => deals.filter((deal) => ["LookingForParticipants", "Discussion", "Confirmed"].includes(deal.status)).length,
    [deals],
  );
  const historyCount = useMemo(
    () => deals.filter((deal) => ["Completed", "Cancelled", "Failed"].includes(deal.status)).length,
    [deals],
  );

  return (
    <Stack spacing={4}>
      <Box
        sx={{
          p: { xs: 3, md: 4 },
          borderRadius: 5,
          color: "common.white",
          background:
            "radial-gradient(circle at top left, rgba(255,255,255,0.24), transparent 30%), linear-gradient(135deg, #1e293b 0%, #0f766e 52%, #f59e0b 100%)",
        }}
      >
        <Typography variant="overline" sx={{ opacity: 0.82, letterSpacing: 1.2 }}>
          Сделки / Home
        </Typography>
        <Typography variant="h3" fontWeight={900} mb={1.5}>
          Все что касается сделок
        </Typography>
        <Typography variant="body1" sx={{ maxWidth: 760, opacity: 0.92 }}>
          Просматривать и управлять черновиками, активными сделками, просмотр истории
          и оставить отзывы после завершения сделки
        </Typography>
      </Box>

      <Grid container spacing={2.5}>
        <Grid size={{ xs: 12, md: 6, xl: 3 }}>
          <SectionEntryCard
            to={appRoutes.deals.tasks}
            icon={<AssignmentTurnedInOutlinedIcon />}
            title="Требуются действия"
            description="Входящие черновики, запросы на присоединение, напоминание об отзыве на товар после сделки"
            badge={totalActionCount}
            accent="warning"
          />
        </Grid>
        <Grid size={{ xs: 12, md: 6, xl: 3 }}>
          <SectionEntryCard
            to={appRoutes.deals.drafts}
            icon={<PlaylistAddCheckCircleOutlinedIcon />}
            title="Черновики"
            description="Входящие и исходящие черновики"
            badge={draftCount}
            accent="primary"
          />
        </Grid>
        <Grid size={{ xs: 12, md: 6, xl: 3 }}>
          <SectionEntryCard
            to={appRoutes.deals.active}
            icon={<AutoModeOutlinedIcon />}
            title="Активные"
            description="Сделки в которых активно идет обсуждение и обмен"
            badge={activeCount}
            accent="info"
          />
        </Grid>
        <Grid size={{ xs: 12, md: 6, xl: 3 }}>
          <SectionEntryCard
            to={appRoutes.deals.history}
            icon={<HistoryOutlinedIcon />}
            title="История"
            description="Завершённые, отменённые и проваленные сделки"
            badge={historyCount}
            accent="secondary"
          />
        </Grid>
      </Grid>

      <Box display="flex" gap={1} flexWrap="wrap">
        <Typography variant="body2" color="text.secondary">
          Сейчас требуют внимания: черновики {draftCount}, запросы на присоединение {joinRequestCount}, отзывы {pendingReviewCount}.
        </Typography>
      </Box>
    </Stack>
  );
}

export default DealsHomePage;
