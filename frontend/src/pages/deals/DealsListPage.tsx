import { useEffect, useState } from "react";
import { Link as RouterLink, useSearchParams } from "react-router-dom";
import { Box, Button, ButtonGroup, Paper, Typography } from "@mui/material";
import type { DealStatus } from "@/features/deals/model/types.ts";
import DealsList from "@/widgets/deals/DealsList";

type DealsStatusTab = "all" | DealStatus;
type DealsStatusCounts = Partial<Record<DealsStatusTab, number>>;
const orderedStatusTabs: DealsStatusTab[] = [
  "all",
  "LookingForParticipants",
  "Discussion",
  "Confirmed",
  "Completed",
  "Cancelled",
  "Failed",
];

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
  const [availableTabs, setAvailableTabs] = useState<DealsStatusTab[] | null>(null);
  const [statusCounts, setStatusCounts] = useState<DealsStatusCounts>({});
  const rawStatus = searchParams.get("status");
  const statusTab: DealsStatusTab = isDealsStatusTab(rawStatus) ? rawStatus : "all";
  const visibleTabs = availableTabs ?? orderedStatusTabs;

  const handleStatusTabChange = (nextStatus: DealsStatusTab) => {
    const nextParams = new URLSearchParams(searchParams);
    nextParams.set("status", nextStatus);
    setSearchParams(nextParams);
  };

  useEffect(() => {
    if (!availableTabs || availableTabs.includes(statusTab)) {
      return;
    }

    const nextStatus = availableTabs[0];
    const nextParams = new URLSearchParams(searchParams);

    if (nextStatus) {
      nextParams.set("status", nextStatus);
    } else {
      nextParams.delete("status");
    }

    setSearchParams(nextParams);
  }, [availableTabs, searchParams, setSearchParams, statusTab]);

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

      {visibleTabs.length > 0 && (
        <Paper variant="outlined" sx={{ p: 1, mb: 3 }}>
          <ButtonGroup
            fullWidth
            variant="text"
            sx={{ display: "grid", gridTemplateColumns: { xs: "1fr", md: "repeat(3, 1fr)", xl: "repeat(4, 1fr)" } }}
          >
            {visibleTabs.map((tab) => (
              <Button
                key={tab}
                variant={statusTab === tab ? "contained" : "text"}
                onClick={() => handleStatusTabChange(tab)}
              >
                {statusTabMeta[tab].title} {statusCounts[tab] ?? 0}
              </Button>
            ))}
          </ButtonGroup>
        </Paper>
      )}

      <Typography variant="h5" fontWeight={700} mb={2}>
        {statusTabMeta[statusTab].title}
      </Typography>

      <DealsList
        statusFilter={statusTab}
        onAvailableStatusTabsChange={setAvailableTabs}
        onStatusCountsChange={setStatusCounts}
      />
    </Box>
  );
}

export default DealsListPage;
