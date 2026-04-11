import { Box, Card, CardContent, LinearProgress, Rating, Stack, Typography } from "@mui/material";
import type { ReviewSummary } from "@/features/reviews/model/types.ts";

interface ReviewSummaryCardProps {
  title: string;
  summary: ReviewSummary;
}

const breakdownOrder = [
  { key: "rating5", label: "5" },
  { key: "rating4", label: "4" },
  { key: "rating3", label: "3" },
  { key: "rating2", label: "2" },
  { key: "rating1", label: "1" },
] as const;

function ReviewSummaryCard({ title, summary }: ReviewSummaryCardProps) {
  return (
    <Card variant="outlined">
      <CardContent>
        <Typography variant="h6" fontWeight={700} mb={2}>
          {title}
        </Typography>

        <Box display="flex" alignItems="center" gap={2} flexWrap="wrap" mb={2}>
          <Typography variant="h3" fontWeight={700}>
            {summary.avgRating.toFixed(1)}
          </Typography>
          <Box>
            <Rating value={summary.avgRating} precision={0.1} readOnly />
            <Typography variant="body2" color="text.secondary">
              {summary.count} {summary.count === 1 ? "отзыв" : summary.count < 5 ? "отзыва" : "отзывов"}
            </Typography>
          </Box>
        </Box>

        <Stack spacing={1}>
          {breakdownOrder.map(({ key, label }) => {
            const value = summary.ratingBreakdown[key];
            const progress = summary.count > 0 ? (value / summary.count) * 100 : 0;

            return (
              <Box key={key} display="flex" alignItems="center" gap={1.5}>
                <Typography variant="body2" sx={{ width: 18 }}>
                  {label}
                </Typography>
                <LinearProgress
                  variant="determinate"
                  value={progress}
                  sx={{ flexGrow: 1, height: 8, borderRadius: 999 }}
                />
                <Typography variant="caption" color="text.secondary" sx={{ width: 24, textAlign: "right" }}>
                  {value}
                </Typography>
              </Box>
            );
          })}
        </Stack>
      </CardContent>
    </Card>
  );
}

export default ReviewSummaryCard;
