import { Box, Grid, Stack, Typography } from "@mui/material";
import AssessmentOutlinedIcon from "@mui/icons-material/AssessmentOutlined";
import BadgeOutlinedIcon from "@mui/icons-material/BadgeOutlined";
import ConnectWithoutContactOutlinedIcon from "@mui/icons-material/ConnectWithoutContactOutlined";
import ReviewsOutlinedIcon from "@mui/icons-material/ReviewsOutlined";
import WorkspacePremiumOutlinedIcon from "@mui/icons-material/WorkspacePremiumOutlined";
import SectionEntryCard from "@/shared/ui/SectionEntryCard.tsx";
import { appRoutes } from "@/shared/config/appRoutes.ts";

function ProfileHomePage() {
  return (
    <Stack spacing={3.5}>
      <Box>
        <Typography variant="overline" color="primary.main">
          Профиль / Home
        </Typography>
        <Typography variant="h4" fontWeight={800} mb={1}>
          Данные о себе
        </Typography>
        <Typography variant="body1" color="text.secondary" maxWidth={820}>
          Просмотр и редактирование личных данных, репутации, подписок, отзывов, статистики
        </Typography>
      </Box>

      <Grid container spacing={2.5}>
        <Grid size={{ xs: 12, md: 6, xl: 4 }}>
          <SectionEntryCard
            to={appRoutes.profile.account}
            icon={<BadgeOutlinedIcon />}
            title="Личные данные"
            description="Профиль, аватар, телефон, биография, текущая точка и смена пароля."
            accent="primary"
          />
        </Grid>
        <Grid size={{ xs: 12, md: 6, xl: 4 }}>
          <SectionEntryCard
            to={appRoutes.profile.reputation}
            icon={<WorkspacePremiumOutlinedIcon />}
            title="Репутация"
            description="Текущие баллы, summary по источникам и отдельная история событий репутации."
            accent="secondary"
          />
        </Grid>
        <Grid size={{ xs: 12, md: 6, xl: 4 }}>
          <SectionEntryCard
            to={appRoutes.profile.networkSubscriptions}
            icon={<ConnectWithoutContactOutlinedIcon />}
            title="Подписки"
            description="Подписки и подписчики как отдельный social subflow, влияющий на direct chat."
            accent="info"
          />
        </Grid>
        <Grid size={{ xs: 12, md: 6, xl: 4 }}>
          <SectionEntryCard
            to={appRoutes.profile.reviewsMine}
            icon={<ReviewsOutlinedIcon />}
            title="Отзывы"
            description="Мои отзывы и отзывы обо мне как персональная история, а не action queue."
            accent="success"
          />
        </Grid>
        <Grid size={{ xs: 12, md: 6, xl: 4 }}>
          <SectionEntryCard
            to={appRoutes.profile.statistics}
            icon={<AssessmentOutlinedIcon />}
            title="Статистика"
            description="Агрегированные метрики по сделкам, отзывам, просмотрам и жалобам."
            accent="warning"
          />
        </Grid>
      </Grid>
    </Stack>
  );
}

export default ProfileHomePage;
