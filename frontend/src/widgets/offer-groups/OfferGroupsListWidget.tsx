import { useMemo } from "react";
import { Alert, Box, CircularProgress, Grid, Typography } from "@mui/material";
import usersApi from "@/features/users/api/usersApi.ts";
import offerGroupsApi from "@/features/offer-groups/api/offerGroupsApi.ts";
import type { OfferGroup } from "@/features/offer-groups/model/types.ts";
import { getOfferGroupOwnerId } from "@/features/offer-groups/model/utils.ts";
import OfferGroupCard from "@/widgets/offer-groups/OfferGroupCard.tsx";

interface OfferGroupsListWidgetProps {
  mode: "mine" | "others";
}

function OfferGroupsListWidget({ mode }: OfferGroupsListWidgetProps) {
  const { data: me } = usersApi.useGetCurrentUserQuery();
  const { data, isLoading, error } = offerGroupsApi.useGetOfferGroupsQuery(
    mode === "mine" ? { my: true } : undefined,
  );

  const items = useMemo(() => {
    if (!data) {
      return [] as OfferGroup[];
    }

    if (mode === "mine") {
      return data;
    }

    if (!me) {
      return data;
    }

    return data.filter((group) => {
      const ownerId = getOfferGroupOwnerId(group);
      if (!ownerId) {
        return false;
      }

      return ownerId !== me.id;
    });
  }, [data, me, mode]);

  if (isLoading) {
    return (
      <Box display="flex" justifyContent="center" py={6}>
        <CircularProgress />
      </Box>
    );
  }

  if (error) {
    return <Alert severity="error">Не удалось загрузить композитные объявления</Alert>;
  }

  if (!data) {
    return <Alert severity="info">Список композитных объявлений недоступен</Alert>;
  }

  if (items.length === 0) {
    return (
      <Typography color="text.secondary" textAlign="center" py={6}>
        {mode === "mine"
          ? "У вас пока нет композитных объявлений"
          : "Пока нет композитных объявлений других пользователей"}
      </Typography>
    );
  }

  return (
    <Grid container spacing={2}>
      {items.map((offerGroup) => (
        <Grid key={offerGroup.id} size={{ xs: 12, md: 6 }}>
          <OfferGroupCard offerGroup={offerGroup} href={`/offer-groups/${offerGroup.id}`} />
        </Grid>
      ))}
    </Grid>
  );
}

export default OfferGroupsListWidget;
