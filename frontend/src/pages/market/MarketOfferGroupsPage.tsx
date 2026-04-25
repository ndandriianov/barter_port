import { Box, Button, ButtonGroup, Stack, Typography } from "@mui/material";
import AddIcon from "@mui/icons-material/Add";
import HubOutlinedIcon from "@mui/icons-material/HubOutlined";
import Inventory2OutlinedIcon from "@mui/icons-material/Inventory2Outlined";
import { Link as RouterLink } from "react-router-dom";
import OfferGroupsListWidget from "@/widgets/offer-groups/OfferGroupsListWidget.tsx";
import { appRoutes } from "@/shared/config/appRoutes.ts";

interface MarketOfferGroupsPageProps {
  mode: "others" | "mine";
}

function MarketOfferGroupsPage({ mode }: MarketOfferGroupsPageProps) {
  return (
    <Stack spacing={3}>
      <Box display="flex" justifyContent="space-between" alignItems="flex-start" gap={2} flexWrap="wrap">
        <Box>
          <Typography variant="overline" color="secondary.main">
            Объявления / Группы объявлений
          </Typography>
          <Typography variant="h4" fontWeight={800} mb={1}>
            {mode === "mine" ? "Мои группы объявлений" : "Группы объявлений"}
          </Typography>
          <Typography variant="body1" color="text.secondary">
            {mode === "mine"
              ? "Здесь собраны ваши группы объявлений"
              : "Группы объявлений позволяют более гибко искать взаимовыгодный обмен"}
          </Typography>
        </Box>
        <Button
          component={RouterLink}
          to={appRoutes.market.createExchangeGroup}
          variant="contained"
          color="secondary"
          startIcon={<AddIcon />}
        >
          Создать группу
        </Button>
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
          to={appRoutes.market.exchangeGroups}
          variant={mode === "others" ? "contained" : "text"}
          startIcon={<HubOutlinedIcon />}
        >
          Все сценарии
        </Button>
        <Button
          component={RouterLink}
          to={appRoutes.market.exchangeGroupsMine}
          variant={mode === "mine" ? "contained" : "text"}
          startIcon={<Inventory2OutlinedIcon />}
        >
          Мои
        </Button>
      </ButtonGroup>

      <OfferGroupsListWidget mode={mode} />
    </Stack>
  );
}

export default MarketOfferGroupsPage;
