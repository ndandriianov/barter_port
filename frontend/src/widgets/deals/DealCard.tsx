import { useEffect, useMemo, useState } from "react";
import {
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  CircularProgress,
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
import { useAppDispatch, useAppSelector } from "@/hooks/redux.ts";
import type { User } from "@/features/users/model/types.ts";

const formatDateTime = (value: string) =>
  new Intl.DateTimeFormat("ru-RU", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(value));

// ─── Edit content dialog ────────────────────────────────────────────────────

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
    if (Object.keys(body).length > 0) {
      await updateDealItem({ dealId, itemId: item.id, body });
    }
    onClose();
  };

  const quantityError = quantity !== "" && (isNaN(parseInt(quantity, 10)) || parseInt(quantity, 10) < 1);

  return (
    <Dialog open onClose={onClose} fullWidth maxWidth="sm">
      <DialogTitle>Редактировать позицию</DialogTitle>
      <DialogContent sx={{ display: "flex", flexDirection: "column", gap: 2, pt: 2 }}>
        <TextField label="Название" value={name} onChange={(e) => setName(e.target.value)} fullWidth size="small" />
        <TextField
          label="Описание" value={description} onChange={(e) => setDescription(e.target.value)}
          fullWidth size="small" multiline minRows={2}
        />
        <TextField
          label="Количество" value={quantity} onChange={(e) => setQuantity(e.target.value)}
          type="number" inputProps={{ min: 1 }} fullWidth size="small"
          error={quantityError} helperText={quantityError ? "Минимум 1" : undefined}
        />
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose} disabled={isLoading}>Отмена</Button>
        <Button onClick={handleSave} variant="contained" disabled={isLoading || quantityError}>Сохранить</Button>
      </DialogActions>
    </Dialog>
  );
}

// ─── Role row (provider or receiver) ────────────────────────────────────────

interface RoleRowProps {
  label: string;
  userId: string | undefined;
  myId: string | undefined;
  isParticipant: boolean;
  getUserName: (id: string) => string;
  onClaim: () => void;
  onRelease: () => void;
  isLoading: boolean;
  canClaim?: boolean;
}

function RoleRow({ label, userId, myId, isParticipant, getUserName, onClaim, onRelease, isLoading, canClaim = true }: RoleRowProps) {
  const isMe = userId !== undefined && userId === myId;
  const isEmpty = userId === undefined;

  return (
    <Box display="flex" alignItems="center" gap={1} mt={0.5}>
      <Typography variant="caption" color="text.secondary" sx={{ minWidth: 80 }}>
        {label}:
      </Typography>
      <Typography variant="caption" fontWeight={isMe ? 600 : 400}>
        {userId ? getUserName(userId) : "не назначен"}
      </Typography>
      {isLoading ? (
        <CircularProgress size={14} />
      ) : isMe ? (
        <Button size="small" variant="text" color="error" sx={{ minWidth: 0, p: 0, fontSize: 11 }} onClick={onRelease}>
          снять себя
        </Button>
      ) : isEmpty && isParticipant && canClaim ? (
        <Button size="small" variant="text" sx={{ minWidth: 0, p: 0, fontSize: 11 }} onClick={onClaim}>
          стать
        </Button>
      ) : null}
    </Box>
  );
}

// ─── Item row ───────────────────────────────────────────────────────────────

interface ItemRowProps {
  item: Item;
  dealId: string;
  myId: string | undefined;
  isParticipant: boolean;
  getUserName: (id: string) => string;
  onEditClick: () => void;
}

function ItemRow({ item, dealId, myId, isParticipant, getUserName, onEditClick }: ItemRowProps) {
  const [updateDealItem, { isLoading }] = dealsApi.useUpdateDealItemMutation();

  const canClaimProvider = myId !== undefined && item.receiverId !== myId;
  const canClaimReceiver = myId !== undefined && item.providerId !== myId;

  const handleClaim = (field: "claimProvider" | "claimReceiver") => () =>
    updateDealItem({ dealId, itemId: item.id, body: { [field]: true } });

  const handleRelease = (field: "releaseProvider" | "releaseReceiver") => () =>
    updateDealItem({ dealId, itemId: item.id, body: { [field]: true } });

  return (
    <ListItem
      disableGutters
      sx={{ borderBottom: "1px solid", borderColor: "divider", pb: 1, mb: 1, flexDirection: "column", alignItems: "flex-start" }}
      secondaryAction={
        myId === item.authorId ? (
          <Tooltip title="Редактировать">
            <IconButton size="small" onClick={onEditClick}>
              <EditIcon fontSize="small" />
            </IconButton>
          </Tooltip>
        ) : undefined
      }
    >
      <ListItemText
        primary={
          <Box display="flex" alignItems="center" gap={1}>
            <Typography variant="body2" fontWeight={500}>{item.name}</Typography>
            <Typography variant="caption" color="text.secondary">x{item.quantity}</Typography>
            <Chip label={item.type} size="small" variant="outlined" />
          </Box>
        }
        secondary={item.description}
      />
      <Box pl={0} mt={0.5}>
        <RoleRow
          label="Поставщик"
          userId={item.providerId}
          myId={myId}
          isParticipant={isParticipant}
          getUserName={getUserName}
          onClaim={handleClaim("claimProvider")}
          onRelease={handleRelease("releaseProvider")}
          isLoading={isLoading}
          canClaim={canClaimProvider}
        />
        <RoleRow
          label="Получатель"
          userId={item.receiverId}
          myId={myId}
          isParticipant={isParticipant}
          getUserName={getUserName}
          onClaim={handleClaim("claimReceiver")}
          onRelease={handleRelease("releaseReceiver")}
          isLoading={isLoading}
          canClaim={canClaimReceiver}
        />
      </Box>
    </ListItem>
  );
}

// ─── DealCard ────────────────────────────────────────────────────────────────

interface DealCardProps {
  deal: Deal;
}

function DealCard({ deal }: DealCardProps) {
  const dispatch = useAppDispatch();
  const [editingItem, setEditingItem] = useState<Item | null>(null);
  const { data: me } = usersApi.useGetCurrentUserQuery();

  // Collect all unique user IDs referenced in items
  const userIds = useMemo(
    () => [
      ...new Set(
        deal.items.flatMap((item) => [
          item.authorId,
          ...(item.providerId ? [item.providerId] : []),
          ...(item.receiverId ? [item.receiverId] : []),
        ]),
      ),
    ],
    [deal.items],
  );

  // Prefetch user info for name resolution
  useEffect(() => {
    if (userIds.length === 0) return;
    const subs = userIds.map((id) => dispatch(usersApi.endpoints.getUserById.initiate(id)));
    return () => subs.forEach((s) => s.unsubscribe());
  }, [dispatch, userIds]);

  const usersById = useAppSelector((state) =>
    userIds.reduce<Record<string, User | undefined>>((acc, id) => {
      acc[id] = usersApi.endpoints.getUserById.select(id)(state).data;
      return acc;
    }, {}),
  );

  const getUserName = (id: string) => usersById[id]?.name?.trim() || "имя не указано";

  // Current user is a participant if they authored at least one item
  const isParticipant = me ? deal.items.some((item) => item.authorId === me.id) : false;

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
          <Typography variant="caption" color="text.disabled">ID: {deal.id}</Typography>
          <Typography variant="caption" color="text.disabled">Создана: {formatDateTime(deal.createdAt)}</Typography>
          {deal.updatedAt && (
            <Typography variant="caption" color="text.disabled">Обновлена: {formatDateTime(deal.updatedAt)}</Typography>
          )}
        </Box>

        <Divider sx={{ mb: 2 }} />

        <Typography variant="subtitle2" fontWeight={600} mb={1}>
          Позиции сделки
        </Typography>

        {deal.items.length === 0 ? (
          <Typography variant="body2" color="text.secondary">Позиции отсутствуют</Typography>
        ) : (
          <List dense disablePadding>
            {deal.items.map((item) => (
              <ItemRow
                key={item.id}
                item={item}
                dealId={deal.id}
                myId={me?.id}
                isParticipant={isParticipant}
                getUserName={getUserName}
                onEditClick={() => setEditingItem(item)}
              />
            ))}
          </List>
        )}
      </CardContent>

      {editingItem && (
        <EditItemDialog item={editingItem} dealId={deal.id} onClose={() => setEditingItem(null)} />
      )}
    </Card>
  );
}

export default DealCard;
