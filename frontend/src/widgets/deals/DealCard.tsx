import {
  Box,
  Card,
  CardContent,
  Chip,
  Divider,
  List,
  ListItem,
  ListItemText,
  Typography,
} from "@mui/material";
import type { Deal } from "@/features/deals/model/types";

const formatDateTime = (value: string) =>
  new Intl.DateTimeFormat("ru-RU", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(value));

interface DealCardProps {
  deal: Deal;
}

function DealCard({ deal }: DealCardProps) {
  return (
    <Card variant="outlined">
      <CardContent>
        <Typography variant="h6" fontWeight={600} gutterBottom>
          {deal.name ?? "Сделка"}
        </Typography>

        {deal.description && (
          <Typography variant="body2" color="text.secondary" mb={2}>
            {deal.description}
          </Typography>
        )}

        <Box display="flex" gap={2} mb={2} flexWrap="wrap">
          <Typography variant="caption" color="text.disabled">
            ID: {deal.id}
          </Typography>
          <Typography variant="caption" color="text.disabled">
            Создана: {formatDateTime(deal.createdAt)}
          </Typography>
          {deal.updatedAt && (
            <Typography variant="caption" color="text.disabled">
              Обновлена: {formatDateTime(deal.updatedAt)}
            </Typography>
          )}
        </Box>

        <Divider sx={{ mb: 2 }} />

        <Typography variant="subtitle2" fontWeight={600} mb={1}>
          Позиции сделки
        </Typography>

        {deal.items.length === 0 ? (
          <Typography variant="body2" color="text.secondary">
            Позиции отсутствуют
          </Typography>
        ) : (
          <List dense disablePadding>
            {deal.items.map((item) => (
              <ListItem
                key={item.id}
                disableGutters
                sx={{ borderBottom: "1px solid", borderColor: "divider", pb: 1, mb: 1 }}
              >
                <ListItemText
                  primary={
                    <Box display="flex" alignItems="center" gap={1}>
                      <Typography variant="body2" fontWeight={500}>
                        {item.name}
                      </Typography>
                      <Typography variant="caption" color="text.secondary">
                        x{item.quantity}
                      </Typography>
                      <Chip label={item.type} size="small" variant="outlined" />
                    </Box>
                  }
                  secondary={item.description}
                />
              </ListItem>
            ))}
          </List>
        )}
      </CardContent>
    </Card>
  );
}

export default DealCard;
