import { useEffect, useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  CircularProgress,
  Divider,
  IconButton,
  TextField,
  Tooltip,
  Typography,
} from "@mui/material";
import EditIcon from "@mui/icons-material/Edit";
import dealsApi from "@/features/deals/api/dealsApi";
import usersApi from "@/features/users/api/usersApi.ts";
import { useAppDispatch, useAppSelector } from "@/hooks/redux.ts";
import type { User } from "@/features/users/model/types.ts";
import type { UpdateDealItemRequest } from "@/features/deals/model/types.ts";
import UserAvatarLabel from "@/shared/UserAvatarLabel.tsx";

const formatDateTime = (value: string) =>
  new Intl.DateTimeFormat("ru-RU", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(value));

// ─── Role row ────────────────────────────────────────────────────────────────

interface RoleRowProps {
  label: string;
  userId: string | undefined;
  myId: string | undefined;
  isParticipant: boolean;
  getUserName: (id: string) => string;
  getUserAvatarUrl: (id: string) => string | undefined;
  onClaim: () => void;
  onRelease: () => void;
  isLoading: boolean;
  canClaim?: boolean;
}

function RoleRow({ label, userId, myId, isParticipant, getUserName, getUserAvatarUrl, onClaim, onRelease, isLoading, canClaim = true }: RoleRowProps) {
  const isMe = userId !== undefined && userId === myId;
  const isEmpty = userId === undefined;

  return (
    <Box display="flex" alignItems="center" gap={1}>
      <Typography variant="body2" color="text.secondary" sx={{ minWidth: 100 }}>
        {label}:
      </Typography>
      {userId ? (
        <UserAvatarLabel
          userId={userId}
          name={getUserName(userId)}
          avatarUrl={getUserAvatarUrl(userId)}
          size={26}
          textVariant="body2"
          fontWeight={isMe ? 600 : 400}
        />
      ) : (
        <Typography variant="body2" fontWeight={isMe ? 600 : 400}>
          не назначен
        </Typography>
      )}
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

// ─── Page ─────────────────────────────────────────────────────────────────────

function DealItemPage() {
  const { dealId, itemId } = useParams<{ dealId: string; itemId: string }>();
  const navigate = useNavigate();
  const dispatch = useAppDispatch();

  const { data: deal, isLoading, error } = dealsApi.useGetDealByIdQuery(dealId ?? "", {
    skip: !dealId,
  });
  const [updateDealItem, { isLoading: isUpdating }] = dealsApi.useUpdateDealItemMutation();

  const item = deal?.items.find((i) => i.id === itemId);

  const [draft, setDraft] = useState<{
    name: string;
    description: string;
    quantity: string;
  } | null>(null);

  const userIds = useMemo(() => {
    if (!item) return [];
    return [...new Set([
      item.authorId,
      ...(item.providerId ? [item.providerId] : []),
      ...(item.receiverId ? [item.receiverId] : []),
    ])];
  }, [item]);

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

  const { data: me } = usersApi.useGetCurrentUserQuery();

  const getUserName = (id: string) => usersById[id]?.name?.trim() || "имя не указано";
  const getUserAvatarUrl = (id: string) => usersById[id]?.avatarUrl;

  if (!dealId || !itemId) return <Alert severity="warning">Позиция не найдена</Alert>;

  if (isLoading) {
    return (
      <Box display="flex" justifyContent="center" py={6}>
        <CircularProgress />
      </Box>
    );
  }

  if (error || !deal) return <Alert severity="error">Не удалось загрузить сделку</Alert>;
  if (!item) return <Alert severity="warning">Позиция не найдена</Alert>;

  const isEditing = draft !== null;
  const isParticipant = me ? deal.participants.includes(me.id) : false;
  const isEditableStatus = deal.status === "LookingForParticipants" || deal.status === "Discussion";
  const canEdit = !!me && me.id === item.authorId && isEditableStatus;
  const canClaimProvider = !!me && item.receiverId !== me.id && isEditableStatus;
  const canClaimReceiver = !!me && item.providerId !== me.id && isEditableStatus;

  const quantityError =
    draft !== null && draft.quantity !== "" && (isNaN(parseInt(draft.quantity, 10)) || parseInt(draft.quantity, 10) < 1);

  const handleSave = async () => {
    if (!draft) return;

    const body: UpdateDealItemRequest = {};
    if (draft.name !== item.name) body.name = draft.name;
    if (draft.description !== item.description) body.description = draft.description;
    const qty = parseInt(draft.quantity, 10);
    if (!isNaN(qty) && qty !== item.quantity) body.quantity = qty;
    if (Object.keys(body).length > 0) {
      await updateDealItem({ dealId, itemId: item.id, body });
    }
    setDraft(null);
  };

  const handleClaim = (field: "claimProvider" | "claimReceiver") => () =>
    updateDealItem({ dealId, itemId: item.id, body: { [field]: true } });

  const handleRelease = (field: "releaseProvider" | "releaseReceiver") => () =>
    updateDealItem({ dealId, itemId: item.id, body: { [field]: true } });

  return (
    <Box maxWidth={700} mx="auto">
      <Button
        size="small"
        variant="text"
        onClick={() => window.history.length > 1 ? navigate(-1) : navigate(`/deals/${dealId}`)}
        sx={{ mb: 2 }}
      >
        ← Назад
      </Button>

      <Box display="flex" alignItems="center" gap={1} mb={3}>
        <Typography variant="h4" fontWeight={700}>
          Позиция сделки
        </Typography>
        {canEdit && !isEditing && (
          <Tooltip title="Редактировать">
            <IconButton
              size="small"
              onClick={() => setDraft({
                name: item.name,
                description: item.description,
                quantity: String(item.quantity),
              })}
            >
              <EditIcon />
            </IconButton>
          </Tooltip>
        )}
      </Box>

      <Card variant="outlined">
        <CardContent sx={{ display: "flex", flexDirection: "column", gap: 2 }}>

          {/* Name */}
          {isEditing ? (
            <TextField
              label="Название"
              value={draft?.name ?? ""}
              onChange={(e) => setDraft((current) => current ? { ...current, name: e.target.value } : current)}
              fullWidth
              size="small"
            />
          ) : (
            <Box>
              <Typography variant="caption" color="text.secondary">Название</Typography>
              <Typography variant="body1" fontWeight={500}>{item.name}</Typography>
            </Box>
          )}

          {/* Description */}
          {isEditing ? (
            <TextField
              label="Описание"
              value={draft?.description ?? ""}
              onChange={(e) => setDraft((current) => current ? { ...current, description: e.target.value } : current)}
              fullWidth
              size="small"
              multiline
              minRows={2}
            />
          ) : (
            <Box>
              <Typography variant="caption" color="text.secondary">Описание</Typography>
              <Typography variant="body2">{item.description}</Typography>
            </Box>
          )}

          {/* Type + Quantity */}
          {isEditing ? (
            <TextField
              label="Количество"
              value={draft?.quantity ?? ""}
              onChange={(e) => setDraft((current) => current ? { ...current, quantity: e.target.value } : current)}
              type="number"
              inputProps={{ min: 1 }}
              fullWidth
              size="small"
              error={quantityError}
              helperText={quantityError ? "Минимум 1" : undefined}
            />
          ) : (
            <Box display="flex" alignItems="center" gap={2}>
              <Box>
                <Typography variant="caption" color="text.secondary">Тип</Typography>
                <Box>
                  <Chip
                    label={item.type === "good" ? "Товар" : "Услуга"}
                    size="small"
                    variant="outlined"
                  />
                </Box>
              </Box>
              <Box>
                <Typography variant="caption" color="text.secondary">Количество</Typography>
                <Typography variant="body2">{item.quantity}</Typography>
              </Box>
            </Box>
          )}

          {isEditing && (
            <Box display="flex" gap={1}>
              <Button
                variant="contained"
                size="small"
                onClick={() => void handleSave()}
                disabled={isUpdating || quantityError || !(draft?.name ?? "").trim()}
              >
                Сохранить
              </Button>
              <Button
                variant="outlined"
                size="small"
                onClick={() => setDraft(null)}
                disabled={isUpdating}
              >
                Отмена
              </Button>
            </Box>
          )}

          <Divider />

          {/* Participants */}
          <Box>
            <Typography variant="caption" color="text.secondary">Автор</Typography>
            <Box mt={0.5}>
              <UserAvatarLabel
                userId={item.authorId}
                name={getUserName(item.authorId)}
                avatarUrl={getUserAvatarUrl(item.authorId)}
                size={30}
                textVariant="body2"
              />
            </Box>
          </Box>

          <RoleRow
            label="Поставщик"
            userId={item.providerId}
            myId={me?.id}
            isParticipant={isParticipant}
            getUserName={getUserName}
            getUserAvatarUrl={getUserAvatarUrl}
            onClaim={handleClaim("claimProvider")}
            onRelease={handleRelease("releaseProvider")}
            isLoading={isUpdating}
            canClaim={canClaimProvider}
          />

          <RoleRow
            label="Получатель"
            userId={item.receiverId}
            myId={me?.id}
            isParticipant={isParticipant}
            getUserName={getUserName}
            getUserAvatarUrl={getUserAvatarUrl}
            onClaim={handleClaim("claimReceiver")}
            onRelease={handleRelease("releaseReceiver")}
            isLoading={isUpdating}
            canClaim={canClaimReceiver}
          />

          {item.updatedAt && (
            <>
              <Divider />
              <Typography variant="caption" color="text.disabled">
                Обновлено: {formatDateTime(item.updatedAt)}
              </Typography>
            </>
          )}
        </CardContent>
      </Card>
    </Box>
  );
}

export default DealItemPage;
