import { useMemo, useState } from "react";
import { Link as RouterLink } from "react-router-dom";
import {
  Alert,
  Box,
  CircularProgress,
  FormControl,
  Grid,
  IconButton,
  InputLabel,
  MenuItem,
  Select,
  Tooltip,
  Typography,
} from "@mui/material";
import RefreshIcon from "@mui/icons-material/Refresh";
import offersApi from "@/features/offers/api/offersApi";
import usersApi from "@/features/users/api/usersApi.ts";
import type { Offer, SortType } from "@/features/offers/model/types";
import useDraftOfferCounts from "@/features/deals/model/useDraftOfferCounts.ts";
import OfferCard from "@/widgets/offers/OfferCard";

interface OffersListWidgetProps {
  mode: "mine" | "others";
}

function OffersListWidget({ mode }: OffersListWidgetProps) {
  const [sortType, setSortType] = useState<SortType>("ByTime");
  const { data: currentUser } = usersApi.useGetCurrentUserQuery();
  const { countsByOfferId } = useDraftOfferCounts({ enabled: mode === "mine" });
  const {
    data,
    isLoading,
    isFetching,
    error,
    refetch,
  } = offersApi.useGetOffersQuery({
    sort: sortType,
    cursor_limit: 20,
    ...(mode === "mine" ? { my: true } : {}),
  });

  const offers = useMemo(() => {
    if (!data?.offers) {
      return [] as Offer[];
    }

    if (mode === "mine") {
      return data.offers;
    }

    if (!currentUser) {
      return data.offers;
    }

    return data.offers.filter((offer) => offer.authorId !== currentUser.id);
  }, [currentUser, data?.offers, mode]);

  if (isLoading) {
    return (
      <Box display="flex" justifyContent="center" py={6}>
        <CircularProgress />
      </Box>
    );
  }

  if (error) {
    return <Alert severity="error">Не удалось загрузить список объявлений</Alert>;
  }

  if (!data) {
    return <Alert severity="info">Список объявлений недоступен</Alert>;
  }

  return (
    <Box>
      <Box display="flex" alignItems="center" gap={2} mb={3} flexWrap="wrap">
        <FormControl size="small" sx={{ minWidth: 200 }}>
          <InputLabel>Сортировка</InputLabel>
          <Select
            value={sortType}
            label="Сортировка"
            onChange={(e) => setSortType(e.target.value as SortType)}
          >
            <MenuItem value="ByTime">Сначала новые</MenuItem>
            <MenuItem value="ByPopularity">По популярности</MenuItem>
          </Select>
        </FormControl>

        <Tooltip title="Обновить">
          <span>
            <IconButton onClick={() => refetch()} disabled={isFetching}>
              <RefreshIcon />
            </IconButton>
          </span>
        </Tooltip>
      </Box>

      {offers.length === 0 ? (
        <Typography color="text.secondary" textAlign="center" py={6}>
          {mode === "mine" ? "У вас пока нет объявлений" : "Пока нет чужих объявлений"}
        </Typography>
      ) : (
        <Grid container spacing={2}>
          {offers.map((offer) => (
            <Grid key={offer.id} size={{ xs: 12, sm: 6, md: 4, lg: 3 }}>
              <RouterLink to={`/offers/${offer.id}`} style={{ textDecoration: "none" }}>
                <OfferCard
                  offer={offer}
                  showRating
                  draftCount={mode === "mine" ? (countsByOfferId[offer.id] ?? 0) : 0}
                />
              </RouterLink>
            </Grid>
          ))}
        </Grid>
      )}
    </Box>
  );
}

export default OffersListWidget;
