import { Box, Button, Grid, Stack, Typography } from "@mui/material";
import AddCircleOutlineOutlinedIcon from "@mui/icons-material/AddCircleOutlineOutlined";
import GridViewOutlinedIcon from "@mui/icons-material/GridViewOutlined";
import Inventory2OutlinedIcon from "@mui/icons-material/Inventory2Outlined";
import ViewInArOutlinedIcon from "@mui/icons-material/ViewInArOutlined";
import { Link as RouterLink } from "react-router-dom";
import SectionEntryCard from "@/shared/ui/SectionEntryCard.tsx";
import { appRoutes } from "@/shared/config/appRoutes.ts";

function MarketHomePage() {
  return (
    <Stack spacing={4}>
      <Box
        sx={{
          p: { xs: 3, md: 4 },
          borderRadius: 5,
          color: "common.white",
          background:
            "radial-gradient(circle at top right, rgba(255,255,255,0.24), transparent 28%), linear-gradient(135deg, #0b3c49 0%, #0f766e 48%, #c26d1f 100%)",
        }}
      >
        <Stack spacing={2.5}>
          <Box>
            <Typography variant="overline" sx={{ opacity: 0.82, letterSpacing: 1.2 }}>
              Объявления / Home
            </Typography>
            <Typography variant="h3" fontWeight={900} mb={1.5}>
              Отдавайте лишнее - получайте нужное
            </Typography>
            <Typography variant="body1" sx={{ maxWidth: 760, opacity: 0.92 }}>
              Вы можете опубликовать свои объявления, просматривать и откликаться на чужие
            </Typography>
          </Box>

          <Box display="flex" gap={1.5} flexWrap="wrap">
            <Button
              component={RouterLink}
              to={appRoutes.market.createOffer}
              variant="contained"
              color="secondary"
              startIcon={<AddCircleOutlineOutlinedIcon />}
            >
              Создать объявление
            </Button>
            <Button
              component={RouterLink}
              to={appRoutes.market.createExchangeGroup}
              variant="outlined"
              sx={{ color: "common.white", borderColor: "rgba(255,255,255,0.4)" }}
            >
              Создать группу объявлений
            </Button>
          </Box>
        </Stack>
      </Box>

      <Grid container spacing={2.5}>
        <Grid size={{ xs: 12, md: 4 }}>
          <SectionEntryCard
            to={appRoutes.market.catalog}
            icon={<GridViewOutlinedIcon />}
            title="Объявления"
            description="Все, подписки, избранное"
            accent="primary"
          />
        </Grid>
        <Grid size={{ xs: 12, md: 4 }}>
          <SectionEntryCard
            to={appRoutes.market.exchangeGroups}
            icon={<ViewInArOutlinedIcon />}
            title="Группы объявлений"
            description="Откликнуться сразу на несколько объявлений"
            accent="secondary"
          />
        </Grid>
        <Grid size={{ xs: 12, md: 4 }}>
          <SectionEntryCard
            to={appRoutes.market.myPublications}
            icon={<Inventory2OutlinedIcon />}
            title="Мои публикации"
            description="Мои объявления, мои группы и жалобы на мои объявления"
            accent="info"
          />
        </Grid>
      </Grid>
    </Stack>
  );
}

export default MarketHomePage;
