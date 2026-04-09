import { useEffect, useMemo, useState } from "react";
import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  CircularProgress,
  MenuItem,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import dealsApi from "@/features/deals/api/dealsApi";
import type { Deal, FailureResolution } from "@/features/deals/model/types";
import type { Me } from "@/features/users/model/types.ts";
import { getStatusCode } from "@/shared/utils/getStatusCode";

interface DealFailureSectionProps {
  deal: Deal;
  me?: Me;
  isParticipant: boolean;
  getUserName: (id: string) => string;
}

const votingStatuses = new Set(["LookingForParticipants", "Discussion", "Confirmed"]);

function formatResolutionText(resolution: FailureResolution, getUserName: (id: string) => string): {
  severity: "success" | "info" | "warning" | "error";
  title: string;
  description: string;
} {
  if (resolution.confirmed === undefined) {
    return {
      severity: "warning",
      title: "Сделка передана на разбор администратору",
      description: resolution.userId
        ? `Предварительно выбран виновник: ${getUserName(resolution.userId)}.`
        : "Голоса участников разошлись, предварительный виновник не определен.",
    };
  }

  if (resolution.confirmed) {
    const guilty = resolution.userId ? getUserName(resolution.userId) : "виновник не указан";
    const punishment =
      resolution.punishmentPoints !== undefined
        ? ` Штрафные баллы: ${resolution.punishmentPoints}.`
        : "";

    return {
      severity: "error",
      title: "Администратор подтвердил провал сделки",
      description: `Решение: ${guilty}.${punishment}${resolution.comment ? ` Комментарий: ${resolution.comment}` : ""}`,
    };
  }

  return {
    severity: "success",
    title: "Администратор не признал сделку проваленной",
    description: resolution.comment ? `Комментарий: ${resolution.comment}` : "Сделка не была признана проваленной.",
  };
}

function DealFailureSection({ deal, me, isParticipant, getUserName }: DealFailureSectionProps) {
  const canAccessFailureData = Boolean(me && (isParticipant || me.isAdmin));
  const canVoteForFailure = Boolean(isParticipant && votingStatuses.has(deal.status));

  const {
    data: votes = [],
    isLoading: isVotesLoading,
    error: votesError,
  } = dealsApi.useGetFailureVotesQuery(deal.id, {
    skip: !canAccessFailureData,
    pollingInterval: 10_000,
  });

  const {
    data: resolution,
    isLoading: isResolutionLoading,
    error: resolutionError,
  } = dealsApi.useGetModeratorResolutionForFailureQuery(deal.id, {
    skip: !canAccessFailureData,
    pollingInterval: 10_000,
  });

  const [voteForFailure, { isLoading: isVoting, error: voteError }] = dealsApi.useVoteForFailureMutation();
  const [revokeVoteForFailure, { isLoading: isRevoking, error: revokeError }] = dealsApi.useRevokeVoteForFailureMutation();

  const ownVote = useMemo(
    () => votes.find((vote) => vote.userId === me?.id),
    [votes, me?.id],
  );

  const [selectedUserId, setSelectedUserId] = useState("");

  useEffect(() => {
    setSelectedUserId(ownVote?.vote ?? "");
  }, [ownVote?.vote]);

  if (!me || !canAccessFailureData) {
    return null;
  }

  const isResolutionMissing = getStatusCode(resolutionError) === 403;
  const hasResolution = resolution !== undefined;
  const hasFailureRecord = hasResolution && !isResolutionMissing;
  const isFailurePending = hasResolution && resolution.confirmed === undefined;
  const canSubmitVote = canVoteForFailure && !hasFailureRecord && Boolean(selectedUserId);
  const canRevokeVote = canVoteForFailure && !hasFailureRecord && Boolean(ownVote);

  const resolutionMeta = resolution ? formatResolutionText(resolution, getUserName) : null;

  return (
    <Card
      variant="outlined"
      sx={{
        mt: 2,
        borderRadius: 3,
        background:
          "linear-gradient(180deg, rgba(183,28,28,0.05) 0%, rgba(183,28,28,0.015) 100%)",
      }}
    >
      <CardContent>
        <Box display="flex" justifyContent="space-between" alignItems="flex-start" gap={2} mb={2} flexWrap="wrap">
          <Box>
            <Typography variant="subtitle1" fontWeight={700}>
              Провал сделки
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Участники могут голосовать за виновника, а администратор принимает итоговое решение.
            </Typography>
          </Box>
          {isFailurePending ? (
            <Chip color="warning" label="На разборе" />
          ) : resolution?.confirmed === true ? (
            <Chip color="error" label="Провал подтвержден" />
          ) : resolution?.confirmed === false ? (
            <Chip color="success" label="Провал отклонен" />
          ) : (
            <Chip variant="outlined" label="Голосование открыто" />
          )}
        </Box>

        {resolutionMeta && (
          <Alert severity={resolutionMeta.severity} sx={{ mb: 2 }}>
            <Typography variant="subtitle2" fontWeight={700}>
              {resolutionMeta.title}
            </Typography>
            <Typography variant="body2">{resolutionMeta.description}</Typography>
          </Alert>
        )}

        {canVoteForFailure && (
          <Stack direction={{ xs: "column", sm: "row" }} spacing={1.5} alignItems={{ sm: "flex-start" }} mb={2}>
            <TextField
              select
              label="Кого считаете виновным"
              value={selectedUserId}
              onChange={(event) => setSelectedUserId(event.target.value)}
              size="small"
              fullWidth
              disabled={isFailurePending || resolution?.confirmed !== undefined || isVoting || isRevoking}
              helperText={
                isFailurePending
                  ? "После достижения порога голосование заморожено до решения администратора"
                  : ownVote
                    ? "Вы можете изменить свой голос до достижения порога"
                    : "Можно голосовать и за себя"
              }
            >
              {deal.participants.map((participantId) => (
                <MenuItem key={participantId} value={participantId}>
                  {getUserName(participantId)}
                </MenuItem>
              ))}
            </TextField>

            <Stack direction="row" spacing={1} flexShrink={0}>
              <Button
                variant="contained"
                color="error"
                disabled={!canSubmitVote || isVoting}
                onClick={() => void voteForFailure({ dealId: deal.id, body: { userId: selectedUserId } })}
              >
                Проголосовать
              </Button>
              <Button
                variant="outlined"
                disabled={!canRevokeVote || isRevoking}
                onClick={() => void revokeVoteForFailure(deal.id)}
              >
                Отозвать
              </Button>
            </Stack>
          </Stack>
        )}

        {voteError && (
          <Alert severity="error" sx={{ mb: 2 }}>
            {getStatusCode(voteError) === 403
              ? "Сейчас голосование по провалу недоступно для этой сделки"
              : getStatusCode(voteError) === 404
                ? "Сделка не найдена"
                : "Не удалось сохранить голос"}
          </Alert>
        )}

        {revokeError && (
          <Alert severity="error" sx={{ mb: 2 }}>
            {getStatusCode(revokeError) === 403
              ? "Сейчас голос нельзя отозвать"
              : getStatusCode(revokeError) === 404
                ? "Сделка не найдена"
                : "Не удалось отозвать голос"}
          </Alert>
        )}

        <Typography variant="subtitle2" fontWeight={700} mb={1}>
          Текущие голоса
        </Typography>

        {isVotesLoading || isResolutionLoading ? (
          <Box display="flex" justifyContent="center" py={2}>
            <CircularProgress size={20} />
          </Box>
        ) : votesError && getStatusCode(votesError) !== 403 ? (
          <Alert severity="error">Не удалось загрузить голоса по провалу сделки</Alert>
        ) : votes.length === 0 ? (
          <Typography variant="body2" color="text.secondary">
            Голосов пока нет
          </Typography>
        ) : (
          <Stack spacing={0.75}>
            {votes.map((vote) => (
              <Typography key={`${vote.userId}-${vote.vote}`} variant="body2">
                {getUserName(vote.userId)} считает виновным: <strong>{getUserName(vote.vote)}</strong>
              </Typography>
            ))}
          </Stack>
        )}
      </CardContent>
    </Card>
  );
}

export default DealFailureSection;
