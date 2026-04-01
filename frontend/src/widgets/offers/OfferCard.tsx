import { Box, Card, CardContent, Chip, Typography } from "@mui/material";
import VisibilityOutlinedIcon from "@mui/icons-material/VisibilityOutlined";
import CalendarTodayOutlinedIcon from "@mui/icons-material/CalendarTodayOutlined";
import type { Offer, OfferAction, OfferType } from "@/features/offers/model/types";

const actionLabels: Record<OfferAction, string> = {
  give: "Отдаю",
  take: "Ищу",
};

const typeLabels: Record<OfferType, string> = {
  good: "Товар",
  service: "Услуга",
};

const actionColors: Record<OfferAction, "success" | "primary"> = {
  give: "success",
  take: "primary",
};

const formatCreatedAt = (value: string) =>
  new Intl.DateTimeFormat("ru-RU", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
  }).format(new Date(value));

interface OfferCardProps {
  offer: Offer;
}

function OfferCard({ offer }: OfferCardProps) {
  return (
    <Card variant="outlined" sx={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <CardContent sx={{ flexGrow: 1 }}>
        <Box display="flex" gap={1} mb={1} flexWrap="wrap">
          <Chip label={typeLabels[offer.type]} size="small" variant="outlined" />
          <Chip label={actionLabels[offer.action]} size="small" color={actionColors[offer.action]} />
        </Box>

        <Typography variant="h6" fontWeight={600} gutterBottom noWrap>
          {offer.name}
        </Typography>

        <Typography variant="body2" color="text.secondary" sx={{ mb: 2, display: "-webkit-box", WebkitLineClamp: 3, WebkitBoxOrient: "vertical", overflow: "hidden" }}>
          {offer.description}
        </Typography>

        <Box display="flex" justifyContent="space-between" alignItems="center" mt="auto">
          <Box display="flex" alignItems="center" gap={0.5} color="text.disabled">
            <VisibilityOutlinedIcon fontSize="small" />
            <Typography variant="caption">{offer.views}</Typography>
          </Box>
          <Box display="flex" alignItems="center" gap={0.5} color="text.disabled">
            <CalendarTodayOutlinedIcon fontSize="small" />
            <Typography variant="caption">{formatCreatedAt(offer.createdAt)}</Typography>
          </Box>
        </Box>
      </CardContent>
    </Card>
  );
}

export default OfferCard;
