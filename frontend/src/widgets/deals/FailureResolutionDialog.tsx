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
  List,
  ListItem,
  ListItemText,
  MenuItem,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import dealsApi from "@/features/deals/api/dealsApi";
import usersApi from "@/features/users/api/usersApi.ts";
import { useAppDispatch, useAppSelector } from "@/hooks/redux.ts";
import type { User } from "@/features/users/model/types.ts";
import { getStatusCode } from "@/shared/utils/getStatusCode";

interface FailureResolutionDialogProps {
  dealId: string;
  onClose: () => void;
}

function FailureResolutionDialog({ dealId, onClose }: FailureResolutionDialogProps) {
  const dispatch = useAppDispatch();
  const [confirmed, setConfirmed] = useState<"true" | "false">("true");
  const [userId, setUserId] = useState("");
  const [punishmentPoints, setPunishmentPoints] = useState("");
  const [comment, setComment] = useState("");

  const { data: materials, isLoading: isMaterialsLoading, error: materialsError } =
    dealsApi.useGetFailureMaterialsQuery(dealId);
  const { data: votes = [], isLoading: isVotesLoading, error: votesError } =
    dealsApi.useGetFailureVotesQuery(dealId);
  const [resolveFailure, { isLoading: isResolving, error: resolveError }] =
    dealsApi.useModeratorResolutionForFailureMutation();

  const userIds = useMemo(
    () =>
      [
        ...(materials?.deal.participants ?? []),
        ...votes.flatMap((vote) => [vote.userId, vote.vote]),
      ].filter((value, index, array) => array.indexOf(value) === index),
    [materials?.deal.participants, votes],
  );

  useEffect(() => {
    if (userIds.length === 0) return;
    const subscriptions = userIds.map((id) =>
      dispatch(usersApi.endpoints.getUserById.initiate(id)),
    );
    return () => subscriptions.forEach((subscription) => subscription.unsubscribe());
  }, [dispatch, userIds]);

  const usersById = useAppSelector((state) =>
    userIds.reduce<Record<string, User | undefined>>((acc, id) => {
      acc[id] = usersApi.endpoints.getUserById.select(id)(state).data;
      return acc;
    }, {}),
  );

  const getUserName = (id: string) => usersById[id]?.name?.trim() || "имя не указано";
  const isConfirmed = confirmed === "true";
  const punishmentError =
    punishmentPoints !== "" &&
    (!Number.isInteger(Number(punishmentPoints)) || Number(punishmentPoints) < 0);

  const handleSubmit = async () => {
    await resolveFailure({
      dealId,
      body: {
        confirmed: isConfirmed,
        userId: isConfirmed && userId ? userId : undefined,
        punishmentPoints: isConfirmed && punishmentPoints !== "" ? Number(punishmentPoints) : undefined,
        comment: comment.trim() ? comment.trim() : undefined,
      },
    }).unwrap();

    onClose();
  };

  return (
    <Dialog open onClose={onClose} fullWidth maxWidth="md">
      <DialogTitle>Разбор провала сделки</DialogTitle>
      <DialogContent sx={{ pt: 2 }}>
        {isMaterialsLoading || isVotesLoading ? (
          <Box display="flex" justifyContent="center" py={4}>
            <CircularProgress />
          </Box>
        ) : materialsError ? (
          <Alert severity="error">Не удалось загрузить материалы по сделке</Alert>
        ) : !materials ? (
          <Alert severity="warning">Материалы по сделке недоступны</Alert>
        ) : (
          <Stack spacing={2.5}>
            <Card variant="outlined" sx={{ borderRadius: 3 }}>
              <CardContent>
                <Stack direction={{ xs: "column", md: "row" }} spacing={2} justifyContent="space-between">
                  <Box>
                    <Typography variant="h6" fontWeight={700}>
                      {materials.deal.name || "Сделка без названия"}
                    </Typography>
                    {materials.deal.description && (
                      <Typography variant="body2" color="text.secondary" mt={0.5}>
                        {materials.deal.description}
                      </Typography>
                    )}
                  </Box>
                  <Stack direction="row" spacing={1} flexWrap="wrap">
                    <Chip label={`Статус: ${materials.deal.status}`} variant="outlined" />
                    {materials.chatId ? (
                      <Chip color="info" label={`chatId: ${materials.chatId}`} />
                    ) : (
                      <Chip label="Чат не создан" variant="outlined" />
                    )}
                  </Stack>
                </Stack>
              </CardContent>
            </Card>

            <Box>
              <Typography variant="subtitle1" fontWeight={700} mb={1}>
                Участники
              </Typography>
              <Stack spacing={0.5}>
                {materials.deal.participants.map((participantId) => (
                  <Typography key={participantId} variant="body2">
                    • {getUserName(participantId)}
                  </Typography>
                ))}
              </Stack>
            </Box>

            <Divider />

            <Box>
              <Typography variant="subtitle1" fontWeight={700} mb={1}>
                Голоса по провалу
              </Typography>
              {votesError ? (
                <Alert severity="error">Не удалось загрузить голоса участников</Alert>
              ) : votes.length === 0 ? (
                <Typography variant="body2" color="text.secondary">
                  Голоса не найдены
                </Typography>
              ) : (
                <List dense disablePadding>
                  {votes.map((vote) => (
                    <ListItem key={`${vote.userId}-${vote.vote}`} disableGutters>
                      <ListItemText
                        primary={getUserName(vote.userId)}
                        secondary={`Считает виновным: ${getUserName(vote.vote)}`}
                      />
                    </ListItem>
                  ))}
                </List>
              )}
            </Box>

            <Divider />

            <Box>
              <Typography variant="subtitle1" fontWeight={700} mb={1.5}>
                Решение администратора
              </Typography>

              <Stack spacing={2}>
                <TextField
                  select
                  label="Решение"
                  value={confirmed}
                  onChange={(event) => setConfirmed(event.target.value as "true" | "false")}
                  size="small"
                >
                  <MenuItem value="true">Подтвердить провал</MenuItem>
                  <MenuItem value="false">Не считать сделку проваленной</MenuItem>
                </TextField>

                {isConfirmed && (
                  <>
                    <TextField
                      select
                      label="Виновник"
                      value={userId}
                      onChange={(event) => setUserId(event.target.value)}
                      size="small"
                      helperText="Можно оставить пустым, если виновник не определен"
                    >
                      <MenuItem value="">
                        <em>Не указывать виновника</em>
                      </MenuItem>
                      {materials.deal.participants.map((participantId) => (
                        <MenuItem key={participantId} value={participantId}>
                          {getUserName(participantId)}
                        </MenuItem>
                      ))}
                    </TextField>

                    <TextField
                      label="Штрафные баллы"
                      type="number"
                      size="small"
                      value={punishmentPoints}
                      onChange={(event) => setPunishmentPoints(event.target.value)}
                      error={punishmentError}
                      helperText={punishmentError ? "Нужно неотрицательное целое число" : undefined}
                    />
                  </>
                )}

                <TextField
                  label="Комментарий"
                  value={comment}
                  onChange={(event) => setComment(event.target.value)}
                  multiline
                  minRows={3}
                  size="small"
                />
              </Stack>
            </Box>

            {resolveError && (
              <Alert severity="error">
                {getStatusCode(resolveError) === 400
                  ? "Проверьте поля решения: при отклонении провала нельзя указывать виновника или штраф"
                  : getStatusCode(resolveError) === 403
                    ? "Сделка больше не находится на рассмотрении или доступ запрещен"
                    : getStatusCode(resolveError) === 404
                      ? "Сделка не найдена"
                      : getStatusCode(resolveError) === 409
                        ? "Решение по этой сделке уже принято"
                        : "Не удалось сохранить решение администратора"}
              </Alert>
            )}
          </Stack>
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose} disabled={isResolving}>
          Закрыть
        </Button>
        <Button
          variant="contained"
          onClick={() => void handleSubmit()}
          disabled={isResolving || isMaterialsLoading || isVotesLoading || Boolean(materialsError) || punishmentError}
        >
          Сохранить решение
        </Button>
      </DialogActions>
    </Dialog>
  );
}

export default FailureResolutionDialog;
