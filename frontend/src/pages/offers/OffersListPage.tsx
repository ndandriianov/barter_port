import { Link as RouterLink, useSearchParams } from "react-router-dom";
import { Box, Button, ButtonGroup, Paper, Typography } from "@mui/material";
import AddIcon from "@mui/icons-material/Add";
import OffersListWidget from "@/widgets/offers/OffersListWidget.tsx";

type OffersTab = "others" | "mine";

const tabMeta: Record<OffersTab, { title: string; description: string }> = {
  others: {
    title: "Чужие объявления",
    description: "Объявления других пользователей, на которые можно откликнуться или посмотреть рейтинг.",
  },
  mine: {
    title: "Мои объявления",
    description: "Ваши опубликованные объявления. Здесь удобно следить за тем, как они выглядят в общем каталоге.",
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
            Чужие объявления
          </Button>
          <Button
            variant={tab === "mine" ? "contained" : "text"}
            onClick={() => handleTabChange("mine")}
          >
            Мои объявления
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
