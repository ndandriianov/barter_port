import { Box, Button, ButtonGroup, Stack, Typography } from "@mui/material";
import AddIcon from "@mui/icons-material/Add";
import FavoriteBorderOutlinedIcon from "@mui/icons-material/FavoriteBorderOutlined";
import NotificationsActiveOutlinedIcon from "@mui/icons-material/NotificationsActiveOutlined";
import StorefrontOutlinedIcon from "@mui/icons-material/StorefrontOutlined";
import { Link as RouterLink } from "react-router-dom";
import usersApi from "@/features/users/api/usersApi.ts";
import OffersListWidget from "@/widgets/offers/OffersListWidget.tsx";
import { appRoutes } from "@/shared/config/appRoutes.ts";

type MarketCatalogMode = "others" | "subscriptions" | "favorites";

interface MarketCatalogPageProps {
  mode: MarketCatalogMode;
}

const meta: Record<MarketCatalogMode, { title: string; description: string }> = {
  others: {
    title: "Все объявления",
    description: "Открытые объявления с сортировками и фильтрами по тэгам",
  },
  subscriptions: {
    title: "Подписки",
    description: "Лента публикаций пользователей, за которыми вы следите",
  },
  favorites: {
    title: "Избранное",
    description: "Сохраненные объявления",
  },
};

function MarketCatalogPage({ mode }: MarketCatalogPageProps) {
  const { data: currentUser } = usersApi.useGetCurrentUserQuery();
  const effectiveMode = currentUser?.isAdmin ? "others" : mode;

  return (
    <Stack spacing={3}>
      <Box display="flex" justifyContent="space-between" alignItems="flex-start" gap={2} flexWrap="wrap">
        <Box>
          <Typography variant="overline" color="primary.main">
            Объявления / Каталог
          </Typography>
          <Typography variant="h4" fontWeight={800} mb={1}>
            {meta[effectiveMode].title}
          </Typography>
          <Typography variant="body1" color="text.secondary">
            {meta[effectiveMode].description}
          </Typography>
        </Box>
        <Button
          component={RouterLink}
          to={appRoutes.market.createOffer}
          variant="contained"
          startIcon={<AddIcon />}
        >
          Создать
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
          to={appRoutes.market.catalog}
          variant={effectiveMode === "others" ? "contained" : "text"}
          startIcon={<StorefrontOutlinedIcon />}
        >
          Все
        </Button>
        {!currentUser?.isAdmin ? (
          <Button
            component={RouterLink}
            to={appRoutes.market.catalogSubscriptions}
            variant={effectiveMode === "subscriptions" ? "contained" : "text"}
            startIcon={<NotificationsActiveOutlinedIcon />}
          >
            Подписки
          </Button>
        ) : null}
        {!currentUser?.isAdmin ? (
          <Button
            component={RouterLink}
            to={appRoutes.market.catalogFavorites}
            variant={effectiveMode === "favorites" ? "contained" : "text"}
            startIcon={<FavoriteBorderOutlinedIcon />}
          >
            Избранное
          </Button>
        ) : null}
      </ButtonGroup>

      <OffersListWidget mode={effectiveMode} />
    </Stack>
  );
}

export default MarketCatalogPage;
