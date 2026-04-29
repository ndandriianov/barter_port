import { Box, Grid, Stack, Typography } from "@mui/material";
import GavelOutlinedIcon from "@mui/icons-material/GavelOutlined";
import ReportProblemOutlinedIcon from "@mui/icons-material/ReportProblemOutlined";
import SettingsSuggestOutlinedIcon from "@mui/icons-material/SettingsSuggestOutlined";
import dealsApi from "@/features/deals/api/dealsApi.ts";
import offersApi from "@/features/offers/api/offersApi.ts";
import SectionEntryCard from "@/shared/ui/SectionEntryCard.tsx";
import { appRoutes } from "@/shared/config/appRoutes.ts";

function ModerationHomePage() {
  const { data: pendingReports = [] } = offersApi.useListAdminOfferReportsQuery("Pending");
  const { data: failureDeals = [] } = dealsApi.useGetDealsForFailureReviewQuery();

  return (
    <Stack spacing={3.5}>
      <Box>
        <Typography variant="overline" color="warning.main">
          Модерация / Home
        </Typography>
        <Typography variant="h4" fontWeight={800} mb={1}>
          Рабочая зона админа
        </Typography>
      </Box>

      <Grid container spacing={2.5}>
        <Grid size={{ xs: 12, md: 4 }}>
          <SectionEntryCard
            to={appRoutes.admin.offerReports}
            icon={<ReportProblemOutlinedIcon />}
            title="Жалобы на объявления"
            description=""
            badge={pendingReports.length}
            accent="warning"
          />
        </Grid>
        <Grid size={{ xs: 12, md: 4 }}>
          <SectionEntryCard
            to={appRoutes.admin.failures}
            icon={<GavelOutlinedIcon />}
            title="Провалы сделок"
            description=""
            badge={failureDeals.length}
            accent="secondary"
          />
        </Grid>
        <Grid size={{ xs: 12, md: 4 }}>
          <SectionEntryCard
            to={appRoutes.admin.system}
            icon={<SettingsSuggestOutlinedIcon />}
            title="Статистика платформы"
            description=""
            accent="info"
          />
        </Grid>
      </Grid>
    </Stack>
  );
}

export default ModerationHomePage;
