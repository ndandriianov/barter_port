import { useMemo } from "react";
import { useLocation, useNavigate, useSearchParams } from "react-router-dom";
import {
  Box,
  Button,
  ButtonGroup,
  Paper,
  Typography,
} from "@mui/material";
import AvailableReviewsWidget from "@/widgets/reviews/AvailableReviewsWidget.tsx";
import MyReviewsWidget from "@/widgets/reviews/MyReviewsWidget.tsx";
import AboutMeReviewsWidget from "@/widgets/reviews/AboutMeReviewsWidget.tsx";
import { appRoutes } from "@/shared/config/appRoutes.ts";

export type ReviewsTab = "available" | "mine" | "about-me";

interface ReviewsPageProps {
  forcedTab?: ReviewsTab;
  title?: string;
  description?: string;
  hideTabs?: boolean;
  hideBackButton?: boolean;
}

const tabMeta: Record<ReviewsTab, { title: string; description: string }> = {
  available: {
    title: "Доступные отзывы",
    description: "Все позиции из завершенных сделок, по которым вы можете оставить отзыв как получатель.",
  },
  mine: {
    title: "Мои отзывы",
    description: "Отзывы, которые вы уже оставили. Здесь их можно посмотреть, отредактировать или удалить.",
  },
  "about-me": {
    title: "Отзывы обо мне",
    description: "Как другие участники сделок оценивают вас как поставщика товара или услуги.",
  },
};

function isReviewsTab(value: string | null): value is ReviewsTab {
  return value === "available" || value === "mine" || value === "about-me";
}

function ReviewsPage({
  forcedTab,
  title,
  description,
  hideTabs = false,
  hideBackButton = false,
}: ReviewsPageProps) {
  const location = useLocation();
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const rawTab = searchParams.get("tab");
  const tab: ReviewsTab = forcedTab ?? (isReviewsTab(rawTab) ? rawTab : "available");
  const selectedDealId = searchParams.get("dealId");
  const shouldShowBackButton = !hideBackButton && location.state?.fromLayoutReviewsButton !== true;

  const meta = useMemo(() => tabMeta[tab], [tab]);

  const handleTabChange = (nextTab: ReviewsTab) => {
    const nextParams = new URLSearchParams();
    nextParams.set("tab", nextTab);

    if (nextTab === "available" && selectedDealId) {
      nextParams.set("dealId", selectedDealId);
    }

    setSearchParams(nextParams);
  };

  return (
    <Box maxWidth={900} mx="auto">
      {shouldShowBackButton && (
        <Button
          size="small"
          variant="text"
          onClick={() => window.history.length > 1 ? navigate(-1) : navigate(appRoutes.market.catalog)}
          sx={{ mb: 2 }}
        >
          ← Назад
        </Button>
      )}

      <Typography variant="h4" fontWeight={700} mb={1}>
        {title ?? "Отзывы"}
      </Typography>
      <Typography variant="body1" color="text.secondary" mb={3}>
        {description ?? meta.description}
      </Typography>

      {!hideTabs && (
        <Paper variant="outlined" sx={{ p: 1, mb: 3 }}>
          <ButtonGroup fullWidth variant="text" sx={{ display: "grid", gridTemplateColumns: { xs: "1fr", md: "repeat(3, 1fr)" } }}>
            <Button
              variant={tab === "mine" ? "contained" : "text"}
              onClick={() => handleTabChange("mine")}
            >
              Мои отзывы
            </Button>
            <Button
              variant={tab === "available" ? "contained" : "text"}
              onClick={() => handleTabChange("available")}
            >
              Доступные отзывы
            </Button>
            <Button
              variant={tab === "about-me" ? "contained" : "text"}
              onClick={() => handleTabChange("about-me")}
            >
              Отзывы обо мне
            </Button>
          </ButtonGroup>
        </Paper>
      )}

      <Typography variant="h5" fontWeight={700} mb={2}>
        {meta.title}
      </Typography>

      {tab === "mine" && <MyReviewsWidget />}
      {tab === "available" && <AvailableReviewsWidget selectedDealId={selectedDealId} />}
      {tab === "about-me" && <AboutMeReviewsWidget />}
    </Box>
  );
}

export default ReviewsPage;
