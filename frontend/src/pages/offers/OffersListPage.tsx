import { Link as RouterLink, useSearchParams } from "react-router-dom";
import { Box, Button, ButtonGroup, Paper, Stack, Typography } from "@mui/material";
import AddIcon from "@mui/icons-material/Add";
import LayersOutlinedIcon from "@mui/icons-material/LayersOutlined";
import OffersListWidget from "@/widgets/offers/OffersListWidget.tsx";

type OffersTab = "others" | "mine";

const tabMeta: Record<OffersTab, { title: string; description: string }> = {
  others: {
    title: "Все объявления",
    description: "Все объявления в каталоге. Список уже приходит с сервера в нужном виде.",
  },
  mine: {
    title: "Только мои",
    description: "Ваши опубликованные объявления. Список уже приходит с сервера отдельно от общего каталога.",
  },
};

function isOffersTab(value: string | null): value is OffersTab {
  return value === "others" || value === "mine";
}

function OffersListPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const rawTab = searchParams.get("tab");
  const tab: OffersTab = isOffersTab(rawTab) ? rawTab : "others";

  const handleTabChange = (nextTab: OffersTab) => {
    const nextParams = new URLSearchParams();
    nextParams.set("tab", nextTab);
    setSearchParams(nextParams);
  };

  return (
    <Box maxWidth={1200} mx="auto">
      <Box display="flex" justifyContent="space-between" alignItems="flex-start" gap={2} mb={3} flexWrap="wrap">
        <div>
          <Typography variant="h4" fontWeight={700} mb={1}>
            Объявления
          </Typography>
          <Typography variant="body1" color="text.secondary">
            {tabMeta[tab].description}
          </Typography>
        </div>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          component={RouterLink}
          to="/offers/create"
        >
          Создать
        </Button>
        <Stack direction="row" spacing={1}>
          <Button
            variant="outlined"
            startIcon={<LayersOutlinedIcon />}
            component={RouterLink}
            to="/offer-groups"
          >
            Композитные
          </Button>
        </Stack>
      </Box>

      <Paper variant="outlined" sx={{ p: 1, mb: 3 }}>
        <ButtonGroup
          fullWidth
          variant="text"
          sx={{ display: "grid", gridTemplateColumns: { xs: "1fr", md: "repeat(2, 1fr)" } }}
        >
          <Button
            variant={tab === "others" ? "contained" : "text"}
            onClick={() => handleTabChange("others")}
          >
            Все объявления
          </Button>
          <Button
            variant={tab === "mine" ? "contained" : "text"}
            onClick={() => handleTabChange("mine")}
          >
            Только мои
          </Button>
        </ButtonGroup>
      </Paper>

      <Typography variant="h5" fontWeight={700} mb={2}>
        {tabMeta[tab].title}
      </Typography>

      {tab === "others" && <OffersListWidget mode="others" />}
      {tab === "mine" && <OffersListWidget mode="mine" />}
    </Box>
  );
}

export default OffersListPage;
