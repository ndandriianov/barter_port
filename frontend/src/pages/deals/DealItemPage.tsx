import {useEffect, useMemo, useState} from "react";
import {useNavigate, useParams} from "react-router-dom";
import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  CircularProgress,
  Dialog,
  DialogContent,
  Divider,
  IconButton,
  ImageList,
  ImageListItem,
  Tooltip,
  Typography,
} from "@mui/material";
import EditIcon from "@mui/icons-material/Edit";
import dealsApi from "@/features/deals/api/dealsApi";
import usersApi from "@/features/users/api/usersApi.ts";
import {useAppDispatch, useAppSelector} from "@/hooks/redux.ts";
import type {User} from "@/features/users/model/types.ts";
import UserAvatarLabel from "@/shared/UserAvatarLabel.tsx";
import DealItemEditDialog from "@/widgets/deals/DealItemEditDialog.tsx";

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

function RoleRow({
                   label,
                   userId,
                   myId,
                   isParticipant,
                   getUserName,
                   getUserAvatarUrl,
                   onClaim,
                   onRelease,
                   isLoading,
                   canClaim = true
                 }: RoleRowProps) {
  const isMe = userId !== undefined && userId === myId;
  const isEmpty = userId === undefined;

  return (
    <Box display="flex" alignItems="center" gap={1}>
      <Typography variant="body2" color="text.secondary" sx={{minWidth: 100}}>
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
        <CircularProgress size={14}/>
      ) : isMe ? (
        <Button size="small" variant="text" color="error" sx={{minWidth: 0, p: 0, fontSize: 11}} onClick={onRelease}>
          снять себя
        </Button>
      ) : isEmpty && isParticipant && canClaim ? (
        <Button size="small" variant="text" sx={{minWidth: 0, p: 0, fontSize: 11}} onClick={onClaim}>
          стать
        </Button>
      ) : null}
    </Box>
  );
}

// ─── Page ─────────────────────────────────────────────────────────────────────

function DealItemPage() {
  const {dealId, itemId} = useParams<{ dealId: string; itemId: string }>();
  const navigate = useNavigate();
  const dispatch = useAppDispatch();

  const {data: deal, isLoading, error} = dealsApi.useGetDealByIdQuery(dealId ?? "", {
    skip: !dealId,
  });
  const [updateDealItem, {isLoading: isUpdating}] = dealsApi.useUpdateDealItemMutation();

  const item = deal?.items.find((i) => i.id === itemId);
  const [isEditDialogOpen, setIsEditDialogOpen] = useState(false);
  const [openedPhotoUrl, setOpenedPhotoUrl] = useState<string | null>(null);

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

  const {data: me} = usersApi.useGetCurrentUserQuery();

  const getUserName = (id: string) => usersById[id]?.name?.trim() || "имя не указано";
  const getUserAvatarUrl = (id: string) => usersById[id]?.avatarUrl;

  if (!dealId || !itemId) return <Alert severity="warning">Позиция не найдена</Alert>;

  if (isLoading) {
    return (
      <Box display="flex" justifyContent="center" py={6}>
        <CircularProgress/>
      </Box>
    );
  }

  if (error || !deal) return <Alert severity="error">Не удалось загрузить сделку</Alert>;
  if (!item) return <Alert severity="warning">Позиция не найдена</Alert>;

  const isParticipant = me ? deal.participants.includes(me.id) : false;
  const isEditableStatus = deal.status === "LookingForParticipants" || deal.status === "Discussion";
  const canEdit = !!me && me.id === item.authorId && isEditableStatus;
  const canClaimProvider = !!me && item.receiverId !== me.id && isEditableStatus;
  const canClaimReceiver = !!me && item.providerId !== me.id && isEditableStatus;

  const handleClaim = (field: "claimProvider" | "claimReceiver") => () =>
    updateDealItem({dealId, itemId: item.id, body: {[field]: true}});

  const handleRelease = (field: "releaseProvider" | "releaseReceiver") => () =>
    updateDealItem({dealId, itemId: item.id, body: {[field]: true}});

  return (
    <Box maxWidth={700} mx="auto">
      <Button
        size="small"
        variant="text"
        onClick={() => window.history.length > 1 ? navigate(-1) : navigate(`/deals/${dealId}`)}
        sx={{mb: 2}}
      >
        ← Назад
      </Button>

      <Box display="flex" alignItems="center" gap={1} mb={3}>
        <Typography variant="h4" fontWeight={700}>
          Позиция сделки
        </Typography>
        {canEdit && (
          <Tooltip title="Редактировать">
            <IconButton
              size="small"
              onClick={() => setIsEditDialogOpen(true)}
            >
              <EditIcon/>
            </IconButton>
          </Tooltip>
        )}
      </Box>

      <Card variant="outlined">
        <CardContent sx={{display: "flex", flexDirection: "column", gap: 2}}>

          {item.photoUrls.length > 0 && (
            <Box
              component="img"
              src={item.photoUrls[0]}
              alt={item.name}
              onClick={() => setOpenedPhotoUrl(item.photoUrls[0])}
              sx={{
                width: "100%",
                height: 280,
                objectFit: "cover",
                borderRadius: 2,
                border: 1,
                borderColor: "divider",
                cursor: "zoom-in",
              }}
            />
          )}

          {item.photoUrls.length > 1 && (
            <Box>
              <Typography variant="h6" fontWeight={600} mb={1.5}>
                Ещё фото
              </Typography>
              <ImageList cols={2} gap={12} sx={{m: 0}}>
                {item.photoUrls.slice(1).map((photoUrl) => (
                  <ImageListItem key={photoUrl} sx={{borderRadius: 2, overflow: "hidden"}}>
                    <Box
                      component="img"
                      src={photoUrl}
                      alt={item.name}
                      onClick={() => setOpenedPhotoUrl(photoUrl)}
                      sx={{
                        width: "100%",
                        height: 220,
                        objectFit: "cover",
                        display: "block",
                        cursor: "zoom-in",
                      }}
                    />
                  </ImageListItem>
                ))}
              </ImageList>
            </Box>
          )}

          <Box>
            <Typography variant="caption" color="text.secondary">Название</Typography>
            <Typography variant="body1" fontWeight={500}>{item.name}</Typography>
          </Box>

          <Box>
            <Typography variant="caption" color="text.secondary">Описание</Typography>
            <Typography variant="body2">{item.description}</Typography>
          </Box>

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

          <Divider/>

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
              <Divider/>
              <Typography variant="caption" color="text.disabled">
                Обновлено: {formatDateTime(item.updatedAt)}
              </Typography>
            </>
          )}
        </CardContent>
      </Card>

      <Dialog
        open={openedPhotoUrl !== null}
        onClose={() => setOpenedPhotoUrl(null)}
        maxWidth="lg"
        fullWidth
      >
        <DialogContent sx={{p: 1.5, bgcolor: "common.black"}}>
          {openedPhotoUrl && (
            <Box
              component="img"
              src={openedPhotoUrl}
              alt={item.name}
              sx={{
                width: "100%",
                maxHeight: "85vh",
                objectFit: "contain",
                display: "block",
              }}
            />
          )}
        </DialogContent>
      </Dialog>

      <DealItemEditDialog
        item={item}
        dealId={dealId}
        open={isEditDialogOpen}
        onClose={() => setIsEditDialogOpen(false)}
      />
    </Box>
  );
}

export default DealItemPage;
