import { Link as RouterLink } from "react-router-dom";
import {
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  Divider,
  List,
  ListItem,
  ListItemText,
  Typography,
} from "@mui/material";
import type { Draft } from "@/features/deals/model/types";

const formatDateTime = (value: string) =>
  new Intl.DateTimeFormat("ru-RU", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(value));

function confirmChip(confirmed: boolean | undefined) {
  if (confirmed === undefined)
    return <Chip label="Ожидает подтверждения" size="small" color="default" />;
  if (confirmed) return <Chip label="Подтверждено" size="small" color="success" />;
  return <Chip label="Не подтверждено" size="small" color="error" />;
}

interface DraftCardProps {
  draft: Draft;
}

function DraftCard({ draft }: DraftCardProps) {
  return (
    <Card variant="outlined">
      <CardContent>
        <Typography variant="h6" fontWeight={600} gutterBottom>
          {draft.name ?? "Черновик сделки"}
        </Typography>

        {draft.description && (
          <Typography variant="body2" color="text.secondary" mb={2}>
            {draft.description}
          </Typography>
        )}

        <Box display="flex" gap={2} mb={2} flexWrap="wrap">
          <Typography variant="caption" color="text.disabled">
            Создан: {formatDateTime(draft.createdAt)}
          </Typography>
          {draft.updatedAt && (
            <Typography variant="caption" color="text.disabled">
              Обновлён: {formatDateTime(draft.updatedAt)}
            </Typography>
          )}
        </Box>

        <Divider sx={{ mb: 2 }} />

        <Typography variant="subtitle2" fontWeight={600} mb={1}>
          Объявления в черновике
        </Typography>

        {draft.offers.length === 0 ? (
          <Typography variant="body2" color="text.secondary">
            Пусто
          </Typography>
        ) : (
          <List dense disablePadding>
            {draft.offers.map((offer) => (
              <ListItem
                key={offer.id}
                disableGutters
                sx={{ borderBottom: "1px solid", borderColor: "divider", pb: 1, mb: 1 }}
              >
                <ListItemText
                  primary={
                    <Box display="flex" alignItems="center" gap={1} flexWrap="wrap">
                      <Typography variant="body2" fontWeight={500}>
                        {offer.name}
                      </Typography>
                      <Typography variant="caption" color="text.disabled">
                        ×{offer.quantity}
                      </Typography>
                      {confirmChip(offer.confirmed)}
                      <Button
                        component={RouterLink}
                        to={`/offers/${offer.id}`}
                        size="small"
                        variant="outlined"
                        sx={{ ml: "auto" }}
                      >
                        Открыть
                      </Button>
                    </Box>
                  }
                />
              </ListItem>
            ))}
          </List>
        )}
      </CardContent>
    </Card>
  );
}

export default DraftCard;
