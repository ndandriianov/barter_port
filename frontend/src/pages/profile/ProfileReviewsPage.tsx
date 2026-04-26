import { Box, Button, ButtonGroup, Stack } from "@mui/material";
import { Link as RouterLink } from "react-router-dom";
import MyReviewsWidget from "@/widgets/reviews/MyReviewsWidget.tsx";
import AboutMeReviewsWidget from "@/widgets/reviews/AboutMeReviewsWidget.tsx";
import usePendingReviews from "@/features/reviews/model/usePendingReviews.ts";
import { appRoutes } from "@/shared/config/appRoutes.ts";
import ProfileSectionShell from "@/widgets/profile/ProfileSectionShell.tsx";

interface ProfileReviewsPageProps {
  mode: "mine" | "about-me";
}

function ProfileReviewsPage({ mode }: ProfileReviewsPageProps) {
  const { pendingCount, isDealsLoading, isPendingLoading } = usePendingReviews();
  const pendingCountLabel = isDealsLoading || isPendingLoading ? "..." : String(pendingCount);

  return (
    <ProfileSectionShell
      title={mode === "mine" ? "Мои отзывы" : "Отзывы обо мне"}
      description=""
      actions={
        <Button component={RouterLink} to={appRoutes.deals.reviews} variant="contained">
          {`Оставить отзыв (${pendingCountLabel})`}
        </Button>
      }
    >
      <Stack spacing={3}>
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
            to={appRoutes.profile.reviewsMine}
            variant={mode === "mine" ? "contained" : "text"}
          >
            Мои отзывы
          </Button>
          <Button
            component={RouterLink}
            to={appRoutes.profile.reviewsAboutMe}
            variant={mode === "about-me" ? "contained" : "text"}
          >
            Отзывы обо мне
          </Button>
        </ButtonGroup>

        <Box maxWidth={900}>
          {mode === "mine" ? <MyReviewsWidget /> : <AboutMeReviewsWidget />}
        </Box>
      </Stack>
    </ProfileSectionShell>
  );
}

export default ProfileReviewsPage;
