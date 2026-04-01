import { useState } from "react";
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
import type { SortType } from "@/features/offers/model/types";
import OfferCard from "@/widgets/offers/OfferCard";

function OffersList() {
  const [sortType, setSortType] = useState<SortType>("ByTime");
  const { data, isLoading, isFetching, error, refetch } = offersApi.useGetOffersQuery({
    sort: sortType,
    cursor_limit: 20,
  });

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
      <Box display="flex" alignItems="center" gap={2} mb={3}>
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

      {data.offers.length === 0 ? (
        <Typography color="text.secondary" textAlign="center" py={6}>
          Пока нет объявлений
        </Typography>
      ) : (
        <Grid container spacing={2}>
          {data.offers.map((offer) => (
            <Grid key={offer.id} size={{ xs: 12, sm: 6, md: 4, lg: 3 }}>
              <RouterLink to={`/offers/${offer.id}`} style={{ textDecoration: "none" }}>
                <OfferCard offer={offer} />
              </RouterLink>
            </Grid>
          ))}
        </Grid>
      )}
    </Box>
  );
}

export default OffersList;
