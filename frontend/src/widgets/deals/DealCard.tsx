import { useState } from "react";
import {
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Divider,
  IconButton,
  List,
  ListItem,
  ListItemText,
  TextField,
  Tooltip,
  Typography,
} from "@mui/material";
import EditIcon from "@mui/icons-material/Edit";
import type { Deal, Item, UpdateDealItemRequest } from "@/features/deals/model/types";
import dealsApi from "@/features/deals/api/dealsApi";
import usersApi from "@/features/users/api/usersApi.ts";

const formatDateTime = (value: string) =>
  new Intl.DateTimeFormat("ru-RU", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(value));

interface EditItemDialogProps {
  item: Item;
  dealId: string;
  onClose: () => void;
}

function EditItemDialog({ item, dealId, onClose }: EditItemDialogProps) {
  const [name, setName] = useState(item.name);
  const [description, setDescription] = useState(item.description);
  const [quantity, setQuantity] = useState(String(item.quantity));

  const [updateDealItem, { isLoading }] = dealsApi.useUpdateDealItemMutation();

  const handleSave = async () => {
    const body: UpdateDealItemRequest = {};
    if (name !== item.name) body.name = name;
    if (description !== item.description) body.description = description;
    const qty = parseInt(quantity, 10);
    if (!isNaN(qty) && qty !== item.quantity) body.quantity = qty;

    if (Object.keys(body).length === 0) {
      onClose();
      return;
    }

    await updateDealItem({ dealId, itemId: item.id, body });
    onClose();
  };

  const quantityError = quantity !== "" && (isNaN(parseInt(quantity, 10)) || parseInt(quantity, 10) < 1);

  return (
    <Dialog open onClose={onClose} fullWidth maxWidth="sm">
      <DialogTitle>Редактировать позицию</DialogTitle>
      <DialogContent sx={{ display: "flex", flexDirection: "column", gap: 2, pt: 2 }}>
        <TextField
          label="Название"
          value={name}
          onChange={(e) => setName(e.target.value)}
          fullWidth
          size="small"
        />
        <TextField
          label="Описание"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          fullWidth
          size="small"
          multiline
          minRows={2}
        />
        <TextField
          label="Количество"
          value={quantity}
          onChange={(e) => setQuantity(e.target.value)}
          type="number"
          inputProps={{ min: 1 }}
          fullWidth
          size="small"
          error={quantityError}
          helperText={quantityError ? "Минимум 1" : undefined}
        />
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose} disabled={isLoading}>
          Отмена
        </Button>
        <Button onClick={handleSave} variant="contained" disabled={isLoading || quantityError}>
          Сохранить
        </Button>
      </DialogActions>
    </Dialog>
  );
}

interface DealCardProps {
  deal: Deal;
}

function DealCard({ deal }: DealCardProps) {
  const [editingItem, setEditingItem] = useState<Item | null>(null);
  const { data: me } = usersApi.useGetCurrentUserQuery();

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
                secondaryAction={
                  me?.id === item.authorId ? (
                    <Tooltip title="Редактировать">
                      <IconButton size="small" onClick={() => setEditingItem(item)}>
                        <EditIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                  ) : undefined
                }
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

      {editingItem && (
        <EditItemDialog
          item={editingItem}
          dealId={deal.id}
          onClose={() => setEditingItem(null)}
        />
      )}
    </Card>
  );
}

export default DealCard;
