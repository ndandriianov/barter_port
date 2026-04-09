import { useEffect, useMemo, useState } from "react";
import {
  Alert,
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
  MenuItem,
  TextField,
  Tooltip,
  Typography,
} from "@mui/material";
import EditIcon from "@mui/icons-material/Edit";
import type { Deal, DealStatus, Item, UpdateDealItemRequest } from "@/features/deals/model/types";
import dealsApi from "@/features/deals/api/dealsApi";
import offersApi from "@/features/offers/api/offersApi";
import usersApi from "@/features/users/api/usersApi.ts";
import { useAppDispatch, useAppSelector } from "@/hooks/redux.ts";
import type { User } from "@/features/users/model/types.ts";
import type { Offer } from "@/features/offers/model/types";
import { getStatusCode } from "@/shared/utils/getStatusCode";
import type { FetchBaseQueryError } from "@reduxjs/toolkit/query";
import type { SerializedError } from "@reduxjs/toolkit";
import DealFailureSection from "@/widgets/deals/DealFailureSection.tsx";

function dealErrorMessage(
  error: FetchBaseQueryError | SerializedError | undefined,
  messages: Partial<Record<number, string>>,
  fallback: string
): string {
  const code = getStatusCode(error);
  return (code !== undefined && messages[code]) ? messages[code]! : fallback;
}

const formatDateTime = (value: string) =>
  new Intl.DateTimeFormat("ru-RU", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(value));

const dealStatusMeta: Record<
  Deal["status"],
  { label: string; color: "default" | "primary" | "secondary" | "success" | "error" | "info" | "warning" }
> = {
  LookingForParticipants: { label: "Поиск участников", color: "info" },
  Discussion: { label: "Обсуждение", color: "warning" },
  Confirmed: { label: "Подтверждена", color: "primary" },
  Completed: { label: "Завершена", color: "success" },
  Cancelled: { label: "Отменена", color: "default" },
  Failed: { label: "Провалена", color: "error" },
};

const nextStatusByCurrent: Partial<Record<DealStatus, DealStatus>> = {
  LookingForParticipants: "Discussion",
  Discussion: "Confirmed",
  Confirmed: "Completed",
};

const isFinalStatus = (status: DealStatus) => ["Completed", "Cancelled", "Failed"].includes(status);

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

interface AddItemDialogProps {
  dealId: string;
  onClose: () => void;
}

function AddItemDialog({ dealId, onClose }: AddItemDialogProps) {
  const [offerId, setOfferId] = useState("");
  const [quantity, setQuantity] = useState("1");
  const { data, isLoading, error } = offersApi.useGetOffersQuery({ sort: "ByTime", my: true, cursor_limit: 100 });
  const [addDealItem, { isLoading: isAdding, error: addError }] = dealsApi.useAddDealItemMutation();

  const offers: Offer[] = data?.offers ?? [];
  const quantityError = quantity !== "" && (isNaN(parseInt(quantity, 10)) || parseInt(quantity, 10) < 1);

  const handleSubmit = async () => {
    const parsedQuantity = parseInt(quantity, 10);
    if (!offerId || !Number.isInteger(parsedQuantity) || parsedQuantity < 1) return;

    await addDealItem({
      dealId,
      body: {
        offerId,
        quantity: parsedQuantity,
      },
    }).unwrap();

    onClose();
  };

  return (
    <Dialog open onClose={onClose} fullWidth maxWidth="sm">
      <DialogTitle>Добавить позицию в сделку</DialogTitle>
      <DialogContent sx={{ display: "flex", flexDirection: "column", gap: 2, pt: 2 }}>
        {isLoading ? (
          <Box display="flex" justifyContent="center" py={2}>
            <CircularProgress size={20} />
          </Box>
        ) : error ? (
          <Alert severity="error">Не удалось загрузить ваши объявления</Alert>
        ) : offers.length === 0 ? (
          <Alert severity="info">У вас нет доступных объявлений для добавления</Alert>
        ) : (
          <TextField
            select
            label="Ваше объявление"
            value={offerId}
            onChange={(e) => setOfferId(e.target.value)}
            fullWidth
            size="small"
          >
            {offers.map((offer) => (
              <MenuItem key={offer.id} value={offer.id}>
                {offer.name}
              </MenuItem>
            ))}
          </TextField>
        )}

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

        {addError && (
          <Alert severity="error">
            {dealErrorMessage(addError, {
              400: "Некорректные данные позиции",
              403: "Добавление позиций недоступно на данном этапе",
              404: "Сделка не найдена",
            }, "Не удалось добавить позицию в сделку")}
          </Alert>
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose} disabled={isAdding}>Отмена</Button>
        <Button
          onClick={() => void handleSubmit()}
          variant="contained"
          disabled={isAdding || !offerId || quantityError || offers.length === 0 || isLoading}
        >
          Добавить
        </Button>
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
  const [isAddItemDialogOpen, setIsAddItemDialogOpen] = useState(false);
  const [isEditingName, setIsEditingName] = useState(false);
  const [nameInput, setNameInput] = useState(deal.name ?? "");
  const { data: me } = usersApi.useGetCurrentUserQuery();
  const [updateDeal, { isLoading: isUpdateDealLoading }] = dealsApi.useUpdateDealMutation();
  const [changeDealStatus, { isLoading: isStatusLoading, error: changeStatusError }] = dealsApi.useChangeDealStatusMutation();
  const [joinDeal, { isLoading: isJoinLoading, error: joinError }] = dealsApi.useJoinDealMutation();
  const [leaveDeal, { isLoading: isLeaveLoading, error: leaveError }] = dealsApi.useLeaveDealMutation();
  const [processJoinRequest, { isLoading: isProcessJoinLoading, error: processJoinError }] = dealsApi.useProcessJoinRequestMutation();

  const isParticipant = me ? deal.participants.includes(me.id) : false;
  const canAccessFailureResolution = Boolean(me && (isParticipant || me.isAdmin));
  const { data: failureResolution } = dealsApi.useGetModeratorResolutionForFailureQuery(deal.id, {
    skip: !canAccessFailureResolution,
    pollingInterval: 10_000,
  });
  const hasFailureResolution = failureResolution !== undefined;
  const isFailurePending = hasFailureResolution && failureResolution.confirmed === undefined;
  const canSeeStatusVotes = Boolean(me && isParticipant && !isFinalStatus(deal.status));
  const {
    data: statusVotes,
    isLoading: isStatusVotesLoading,
    error: statusVotesError,
  } = dealsApi.useGetDealStatusVotesQuery(deal.id, {
    skip: !canSeeStatusVotes,
    pollingInterval: 10_000,
  });
  const canSeeJoinRequests = Boolean(me && isParticipant && deal.status === "LookingForParticipants");
  const {
    data: joinRequests,
    isLoading: isJoinRequestsLoading,
    error: joinRequestsError,
  } = dealsApi.useGetDealJoinsQuery(deal.id, {
    skip: !canSeeJoinRequests,
    pollingInterval: 10_000,
  });

  // Collect all unique user IDs referenced in items
  const userIds = useMemo(
    () => [
      ...new Set(
        [
          ...deal.items.flatMap((item) => [
            item.authorId,
            ...(item.providerId ? [item.providerId] : []),
            ...(item.receiverId ? [item.receiverId] : []),
          ]),
          ...(statusVotes?.map((statusVote) => statusVote.userId) ?? []),
          ...(joinRequests?.flatMap((request) => [request.userId, ...request.voters]) ?? []),
        ],
      ),
    ],
    [deal.items, joinRequests, statusVotes],
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

  const groupedStatusVotes = useMemo(() => {
    if (!statusVotes || statusVotes.length === 0) return [] as Array<{ status: DealStatus; voters: string[] }>;

    const grouped = new Map<DealStatus, string[]>();
    statusVotes.forEach(({ vote, userId }) => {
      const voters = grouped.get(vote) ?? [];
      voters.push(userId);
      grouped.set(vote, voters);
    });

    return [...grouped.entries()].map(([status, voters]) => ({ status, voters }));
  }, [statusVotes]);

  const nextStatus: DealStatus | undefined = nextStatusByCurrent[deal.status as DealStatus];
  const canVoteForNextStatus = isParticipant && nextStatus !== undefined && !isFailurePending;
  const canCancelDeal = isParticipant && !isFinalStatus(deal.status) && !isFailurePending;
  const canJoinDeal = Boolean(me && !isParticipant && deal.status === "LookingForParticipants");
  const canLeaveDeal = Boolean(me && isParticipant && !isFinalStatus(deal.status) && !isFailurePending);
  const canVoteJoinRequests = Boolean(me && isParticipant && deal.status === "LookingForParticipants" && !isFailurePending);
  const canAddItems = Boolean(me && isParticipant && deal.status === "LookingForParticipants" && !isFailurePending);
  const hasActions = canVoteForNextStatus || canCancelDeal || canJoinDeal || canLeaveDeal;

  const handleChangeStatus = async (expectedStatus: DealStatus) => {
    await changeDealStatus({ dealId: deal.id, body: { expectedStatus } }).unwrap();
  };

  const handleJoinDeal = async () => {
    await joinDeal(deal.id).unwrap();
  };

  const handleLeaveDeal = async () => {
    await leaveDeal(deal.id).unwrap();
  };

  const handleProcessJoin = async (userId: string, accept: boolean) => {
    await processJoinRequest({ dealId: deal.id, userId, accept }).unwrap();
  };

  const handleSaveName = async () => {
    const trimmed = nameInput.trim();
    if (trimmed && trimmed !== deal.name) {
      await updateDeal({ dealId: deal.id, name: trimmed });
    }
    setIsEditingName(false);
  };

  const handleCancelEditName = () => {
    setNameInput(deal.name ?? "");
    setIsEditingName(false);
  };

  return (
    <Card variant="outlined">
      <CardContent>
        {isEditingName ? (
          <Box display="flex" alignItems="center" gap={1} mb={1}>
            <TextField
              value={nameInput}
              onChange={(e) => setNameInput(e.target.value)}
              size="small"
              fullWidth
              autoFocus
              onKeyDown={(e) => {
                if (e.key === "Enter") void handleSaveName();
                if (e.key === "Escape") handleCancelEditName();
              }}
            />
            <Button size="small" variant="contained" onClick={() => void handleSaveName()} disabled={isUpdateDealLoading || !nameInput.trim()}>
              Сохранить
            </Button>
            <Button size="small" onClick={handleCancelEditName} disabled={isUpdateDealLoading}>
              Отмена
            </Button>
          </Box>
        ) : (
          <Box display="flex" alignItems="center" gap={1} mb={1}>
            <Typography variant="h6" fontWeight={600}>
              {deal.name ?? "Сделка"}
            </Typography>
            {isParticipant && !isFinalStatus(deal.status) && (
              <Tooltip title="Переименовать">
                <IconButton size="small" onClick={() => { setNameInput(deal.name ?? ""); setIsEditingName(true); }}>
                  <EditIcon fontSize="small" />
                </IconButton>
              </Tooltip>
            )}
          </Box>
        )}

        <Box mb={1.5}>
          <Chip
            size="small"
            label={`Статус: ${dealStatusMeta[deal.status].label}`}
            color={dealStatusMeta[deal.status].color}
            variant="outlined"
          />

          {hasActions && (
            <Box mt={1.5} display="flex" gap={1} flexWrap="wrap">
              {canVoteForNextStatus && nextStatus && (
                <Button
                  size="small"
                  variant="contained"
                  onClick={() => void handleChangeStatus(nextStatus)}
                  disabled={isStatusLoading}
                >
                  Голосовать за "{dealStatusMeta[nextStatus as Deal["status"]].label}"
                </Button>
              )}

              {canCancelDeal && (
                <Button
                  size="small"
                  variant="outlined"
                  color="error"
                  onClick={() => void handleChangeStatus("Cancelled")}
                  disabled={isStatusLoading || deal.status === "Cancelled" || isFailurePending}
                >
                  Отменить сделку
                </Button>
              )}

              {canJoinDeal && (
                <Button
                  size="small"
                  variant="contained"
                  color="success"
                  onClick={() => void handleJoinDeal()}
                  disabled={isJoinLoading}
                >
                  Откликнуться
                </Button>
              )}

              {canLeaveDeal && (
                <Button
                  size="small"
                  variant="outlined"
                  color="warning"
                  onClick={() => void handleLeaveDeal()}
                  disabled={isLeaveLoading}
                >
                  Покинуть сделку
                </Button>
              )}
            </Box>
          )}

          {changeStatusError && (
            <Alert severity="error" sx={{ mt: 1.5 }}>
              {dealErrorMessage(changeStatusError, {
                400: "Статус сделки изменился. Обновите страницу",
                403: isFailurePending ? "Сделка уже передана на разбор по провалу" : "Недостаточно прав для смены статуса",
                404: "Сделка не найдена",
              }, "Не удалось отправить голос за статус сделки")}
            </Alert>
          )}

          {joinError && (
            <Alert severity="error" sx={{ mt: 1.5 }}>
              {dealErrorMessage(joinError, {
                400: "Сделка больше не принимает участников",
                403: "Вы не можете присоединиться к этой сделке",
                404: "Сделка не найдена",
              }, "Не удалось откликнуться на сделку")}
            </Alert>
          )}

          {leaveError && (
            <Alert severity="error" sx={{ mt: 1.5 }}>
              {dealErrorMessage(leaveError, {
                400: "Невозможно покинуть сделку на данном этапе",
                403: isFailurePending ? "Сделка уже передана на разбор по провалу" : "Вы не можете покинуть эту сделку",
                404: "Сделка не найдена",
              }, "Не удалось покинуть сделку")}
            </Alert>
          )}

          {processJoinError && (
            <Alert severity="error" sx={{ mt: 1.5 }}>
              {dealErrorMessage(processJoinError, {
                400: "Невозможно обработать заявку",
                403: isFailurePending ? "Сделка уже передана на разбор по провалу" : "Недостаточно прав для обработки заявки",
                404: "Пользователь или сделка не найдены",
              }, "Не удалось обработать заявку на вступление")}
            </Alert>
          )}

          {isFailurePending && (
            <Alert severity="warning" sx={{ mt: 1.5 }}>
              По сделке достигнут порог голосов о провале. Изменение состава, позиций и статуса
              заблокировано до решения администратора.
            </Alert>
          )}

          {canSeeStatusVotes && (
            <Box mt={1.5}>
              <Typography variant="subtitle2" fontWeight={600} mb={0.5}>
                Голоса за статус
              </Typography>

              {isStatusVotesLoading ? (
                <Box display="flex" justifyContent="center" py={1}>
                  <CircularProgress size={18} />
                </Box>
              ) : statusVotesError && getStatusCode(statusVotesError) !== 401 ? (
                <Alert severity="error">
                  {dealErrorMessage(statusVotesError, {
                    404: "Данные не найдены",
                  }, "Не удалось загрузить голоса по статусу")}
                </Alert>
              ) : groupedStatusVotes.length === 0 ? (
                <Typography variant="body2" color="text.secondary">
                  Голосов пока нет
                </Typography>
              ) : (
                <Box display="flex" flexDirection="column" gap={0.5}>
                  {groupedStatusVotes.map(({ status, voters }) => (
                    <Typography key={status} variant="caption" color="text.secondary">
                      За "{dealStatusMeta[status].label}": {voters.map(getUserName).join(", ")}
                    </Typography>
                  ))}
                </Box>
              )}
            </Box>
          )}
        </Box>

        {deal.description && (
          <Typography variant="body2" color="text.secondary" mb={2}>
            {deal.description}
          </Typography>
        )}

        <Box display="flex" gap={2} mb={2} flexWrap="wrap">
          <Typography variant="caption" color="text.disabled">Создана: {formatDateTime(deal.createdAt)}</Typography>
          {deal.updatedAt && (
            <Typography variant="caption" color="text.disabled">Обновлена: {formatDateTime(deal.updatedAt)}</Typography>
          )}
        </Box>

        <Divider sx={{ mb: 2 }} />

        <Box mb={2}>
          <Typography variant="subtitle2" fontWeight={600} mb={1}>
            Участники ({deal.participants.length})
          </Typography>

          {deal.participants.length === 0 ? (
            <Typography variant="body2" color="text.secondary">
              Участников пока нет
            </Typography>
          ) : (
            <Box display="flex" flexDirection="column" gap={0.5}>
              {deal.participants.map((participantId) => (
                <Typography key={participantId} variant="body2">
                  • {getUserName(participantId)}
                </Typography>
              ))}
            </Box>
          )}
        </Box>

        <Divider sx={{ mb: 2 }} />

        {canSeeJoinRequests && (
          <Box mb={2}>
            <Typography variant="subtitle2" fontWeight={600} mb={1}>
              Заявки на вступление
            </Typography>

            {isJoinRequestsLoading ? (
              <Box display="flex" justifyContent="center" py={1}>
                <CircularProgress size={18} />
              </Box>
            ) : joinRequestsError && getStatusCode(joinRequestsError) !== 401 ? (
              <Alert severity="error" sx={{ mb: 1.5 }}>
                {dealErrorMessage(joinRequestsError, {
                  404: "Данные не найдены",
                }, "Не удалось загрузить заявки на вступление")}
              </Alert>
            ) : !joinRequests || joinRequests.length === 0 ? (
              <Typography variant="body2" color="text.secondary">
                Заявок пока нет
              </Typography>
            ) : (
              <Box display="flex" flexDirection="column" gap={1}>
                {joinRequests.map((request) => {
                  const hasVoted = Boolean(me && request.voters.includes(me.id));

                  return (
                    <Box
                      key={request.userId}
                      sx={{ border: "1px solid", borderColor: "divider", borderRadius: 1, p: 1.5 }}
                    >
                      <Typography variant="body2" fontWeight={600}>
                        {getUserName(request.userId)}
                      </Typography>

                      <Typography variant="caption" color="text.secondary" display="block" mt={0.5}>
                        Голоса: {request.voters.length > 0 ? request.voters.map(getUserName).join(", ") : "пока нет"}
                      </Typography>

                      {canVoteJoinRequests && request.userId !== me?.id && (
                        <Box display="flex" gap={1} mt={1}>
                          <Button
                            size="small"
                            variant="outlined"
                            color="success"
                            onClick={() => void handleProcessJoin(request.userId, true)}
                            disabled={isProcessJoinLoading || hasVoted}
                          >
                            Принять
                          </Button>
                          <Button
                            size="small"
                            variant="outlined"
                            color="error"
                            onClick={() => void handleProcessJoin(request.userId, false)}
                            disabled={isProcessJoinLoading || hasVoted}
                          >
                            Отклонить
                          </Button>
                        </Box>
                      )}
                    </Box>
                  );
                })}
              </Box>
            )}

            <Divider sx={{ mt: 2 }} />
          </Box>
        )}

        <Box display="flex" alignItems="center" justifyContent="space-between" gap={1} mb={1}>
          <Typography variant="subtitle2" fontWeight={600}>
            Позиции сделки
          </Typography>
          {canAddItems && (
            <Button size="small" variant="outlined" onClick={() => setIsAddItemDialogOpen(true)}>
              Добавить позицию
            </Button>
          )}
        </Box>

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

        <DealFailureSection deal={deal} me={me} isParticipant={isParticipant} getUserName={getUserName} />
      </CardContent>

      {editingItem && (
        <EditItemDialog item={editingItem} dealId={deal.id} onClose={() => setEditingItem(null)} />
      )}

      {isAddItemDialogOpen && (
        <AddItemDialog dealId={deal.id} onClose={() => setIsAddItemDialogOpen(false)} />
      )}
    </Card>
  );
}

export default DealCard;
