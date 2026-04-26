import { Alert, Box, Grid, Stack, Typography } from "@mui/material";
import Inventory2OutlinedIcon from "@mui/icons-material/Inventory2Outlined";
import ReportProblemOutlinedIcon from "@mui/icons-material/ReportProblemOutlined";
import ViewInArOutlinedIcon from "@mui/icons-material/ViewInArOutlined";
import offersApi from "@/features/offers/api/offersApi.ts";
import offerGroupsApi from "@/features/offer-groups/api/offerGroupsApi.ts";
import SectionEntryCard from "@/shared/ui/SectionEntryCard.tsx";
import { appRoutes } from "@/shared/config/appRoutes.ts";

function MyPublicationsPage() {
  const { data: offersData, isLoading: isOffersLoading } = offersApi.useGetOffersQuery({
    sort: "ByTime",
    my: true,
    cursor_limit: 100,
  });
  const { data: groups = [], isLoading: isGroupsLoading } = offerGroupsApi.useGetOfferGroupsQuery({ my: true });

  const myGroupsCount = groups.length;
  const myOffers = offersData?.offers ?? [];
  const moderationCount = myOffers.filter((offer) => offer.isHidden || offer.modificationBlocked).length;

  return (
    <Stack spacing={3.5}>
      <Box>
        <Typography variant="overline" color="info.main">
          Объявления / Мои публикации
        </Typography>
        <Typography variant="h4" fontWeight={800} mb={1}>
          Управление собственными материалами
        </Typography>
      </Box>

      <Grid container spacing={2.5}>
        <Grid size={{ xs: 12, md: 4 }}>
          <SectionEntryCard
            to={appRoutes.market.myPublicationOffers}
            icon={<Inventory2OutlinedIcon />}
            title="Мои объявления"
            description="Редактирование, удаление и отклики по моим публикациям"
            badge={isOffersLoading ? "..." : myOffers.length}
            accent="primary"
          />
        </Grid>
        <Grid size={{ xs: 12, md: 4 }}>
          <SectionEntryCard
            to={appRoutes.market.exchangeGroupsMine}
            icon={<ViewInArOutlinedIcon />}
            title="Мои группы"
            description="Редактирование, удаление и отклики по моим публикациям"
            badge={isGroupsLoading ? "..." : myGroupsCount}
            accent="secondary"
          />
        </Grid>
        <Grid size={{ xs: 12, md: 4 }}>
          <SectionEntryCard
            to={appRoutes.market.myPublicationModeration}
            icon={<ReportProblemOutlinedIcon />}
            title="Жалобы"
            description="Жалобы на мои публикации"
            badge={isOffersLoading ? "..." : moderationCount}
            accent="warning"
          />
        </Grid>
      </Grid>

      {moderationCount > 0 && (
        <Alert severity="warning">
          У части ваших публикаций есть модерационные ограничения. Откройте раздел «На модерации»,
          чтобы увидеть жалобы, скрытие или блокировку изменений.
        </Alert>
      )}
    </Stack>
  );
}

export default MyPublicationsPage;
