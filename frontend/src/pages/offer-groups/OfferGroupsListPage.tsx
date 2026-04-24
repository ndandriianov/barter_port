import { Link as RouterLink, useSearchParams } from "react-router-dom";
import { Box, Button, ButtonGroup, Paper, Typography } from "@mui/material";
import AddIcon from "@mui/icons-material/Add";
import OfferGroupsListWidget from "@/widgets/offer-groups/OfferGroupsListWidget.tsx";
import { appRoutes } from "@/shared/config/appRoutes.ts";

export type OfferGroupsTab = "others" | "mine";

interface OfferGroupsListPageProps {
  forcedTab?: OfferGroupsTab;
  title?: string;
  description?: string;
  hideTabs?: boolean;
}

const tabMeta: Record<OfferGroupsTab, { title: string; description: string }> = {
  others: {
    title: "Чужие композитные объявления",
    description:
      "Группы offer с логикой AND/OR. На этапе отклика можно выбрать подходящий вариант из каждого блока.",
  },
  mine: {
    title: "Мои композитные объявления",
    description:
      "Собранные вами альтернативные сценарии обмена на базе обычных объявлений.",
  },
};

function isOfferGroupsTab(value: string | null): value is OfferGroupsTab {
  return value === "others" || value === "mine";
}

function OfferGroupsListPage({
  forcedTab,
  title,
  description,
  hideTabs = false,
}: OfferGroupsListPageProps) {
  const [searchParams, setSearchParams] = useSearchParams();
  const rawTab = searchParams.get("tab");
  const tab: OfferGroupsTab = forcedTab ?? (isOfferGroupsTab(rawTab) ? rawTab : "others");

  const handleTabChange = (nextTab: OfferGroupsTab) => {
    const nextParams = new URLSearchParams();
    nextParams.set("tab", nextTab);
    setSearchParams(nextParams);
  };

  return (
    <Box maxWidth={1200} mx="auto">
      <Box display="flex" justifyContent="space-between" alignItems="flex-start" gap={2} mb={3} flexWrap="wrap">
        <Box>
          <Typography variant="h4" fontWeight={700} mb={1}>
            {title ?? "Композитные объявления"}
          </Typography>
          <Typography variant="body1" color="text.secondary">
            {description ?? tabMeta[tab].description}
          </Typography>
        </Box>

        <Button variant="contained" startIcon={<AddIcon />} component={RouterLink} to={appRoutes.market.createExchangeGroup}>
          Создать группу
        </Button>
      </Box>

      {!hideTabs && (
        <Paper variant="outlined" sx={{ p: 1, mb: 3 }}>
          <ButtonGroup
            fullWidth
            variant="text"
            sx={{ display: "grid", gridTemplateColumns: { xs: "1fr", md: "repeat(2, 1fr)" } }}
          >
            <Button variant={tab === "others" ? "contained" : "text"} onClick={() => handleTabChange("others")}>
              Чужие группы
            </Button>
            <Button variant={tab === "mine" ? "contained" : "text"} onClick={() => handleTabChange("mine")}>
              Мои группы
            </Button>
          </ButtonGroup>
        </Paper>
      )}

      <Typography variant="h5" fontWeight={700} mb={2}>
        {tabMeta[tab].title}
      </Typography>

      <OfferGroupsListWidget mode={tab} />
    </Box>
  );
}

export default OfferGroupsListPage;
