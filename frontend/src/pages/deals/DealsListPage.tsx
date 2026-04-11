import { Link as RouterLink, useSearchParams } from "react-router-dom";
import { Box, Button, ButtonGroup, Paper, Typography } from "@mui/material";
import type { DealStatus } from "@/features/deals/model/types.ts";
import DealsList from "@/widgets/deals/DealsList";

type DealsStatusTab = "all" | DealStatus;

const statusTabMeta: Record<DealsStatusTab, { title: string; description: string }> = {
  all: {
    title: "Все сделки",
    description: "Общий список сделок по всем статусам.",
  },
  LookingForParticipants: {
    title: "В поиске участников",
    description: "Сделки, в которые ещё можно добирать участников.",
  },
  Discussion: {
    title: "Обсуждение",
    description: "Сделки, которые уже перешли к согласованию условий.",
  },
  Confirmed: {
    title: "Подтверждены",
    description: "Сделки, где участники подтвердили договорённости.",
  },
  Completed: {
    title: "Завершены",
    description: "Успешно завершённые сделки.",
  },
  Cancelled: {
    title: "Отменены",
    description: "Сделки, которые были отменены.",
  },
  Failed: {
    title: "Не состоялись",
    description: "Сделки, завершившиеся неуспешно.",
  },
};

function isDealsStatusTab(value: string | null): value is DealsStatusTab {
  return value === "all" ||
    value === "LookingForParticipants" ||
    value === "Discussion" ||
    value === "Confirmed" ||
    value === "Completed" ||
    value === "Cancelled" ||
    value === "Failed";
}

function DealsListPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const rawStatus = searchParams.get("status");
  const statusTab: DealsStatusTab = isDealsStatusTab(rawStatus) ? rawStatus : "all";

  const handleStatusTabChange = (nextStatus: DealsStatusTab) => {
    const nextParams = new URLSearchParams(searchParams);
    nextParams.set("status", nextStatus);
    setSearchParams(nextParams);
  };

  return (
    <Box maxWidth={1200} mx="auto">
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3} flexWrap="wrap" gap={1}>
        <Box>
          <Typography variant="h4" fontWeight={700} mb={1}>
            Сделки
          </Typography>
          <Typography variant="body1" color="text.secondary">
            {statusTabMeta[statusTab].description}
          </Typography>
        </Box>
        <Box display="flex" gap={1}>
          <Button variant="outlined" component={RouterLink} to="/deals/drafts">
            Мои черновики
          </Button>
        </Box>
      </Box>

      <Paper variant="outlined" sx={{ p: 1, mb: 3 }}>
        <ButtonGroup
          fullWidth
          variant="text"
          sx={{ display: "grid", gridTemplateColumns: { xs: "1fr", md: "repeat(3, 1fr)", xl: "repeat(4, 1fr)" } }}
        >
          <Button variant={statusTab === "all" ? "contained" : "text"} onClick={() => handleStatusTabChange("all")}>
            Все
          </Button>
          <Button
            variant={statusTab === "LookingForParticipants" ? "contained" : "text"}
            onClick={() => handleStatusTabChange("LookingForParticipants")}
          >
            В поиске участников
          </Button>
          <Button
            variant={statusTab === "Discussion" ? "contained" : "text"}
            onClick={() => handleStatusTabChange("Discussion")}
          >
            Обсуждение
          </Button>
          <Button
            variant={statusTab === "Confirmed" ? "contained" : "text"}
            onClick={() => handleStatusTabChange("Confirmed")}
          >
            Подтверждены
          </Button>
          <Button
            variant={statusTab === "Completed" ? "contained" : "text"}
            onClick={() => handleStatusTabChange("Completed")}
          >
            Завершены
          </Button>
          <Button
            variant={statusTab === "Cancelled" ? "contained" : "text"}
            onClick={() => handleStatusTabChange("Cancelled")}
          >
            Отменены
          </Button>
          <Button
            variant={statusTab === "Failed" ? "contained" : "text"}
            onClick={() => handleStatusTabChange("Failed")}
          >
            Не состоялись
          </Button>
        </ButtonGroup>
      </Paper>

      <Typography variant="h5" fontWeight={700} mb={2}>
        {statusTabMeta[statusTab].title}
      </Typography>

      <DealsList statusFilter={statusTab} />
    </Box>
  );
}

export default DealsListPage;
