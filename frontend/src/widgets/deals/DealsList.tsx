import { useState } from "react";
import { Link as RouterLink } from "react-router-dom";
import {
  Alert,
  Box,
  Checkbox,
  CircularProgress,
  FormControlLabel,
  FormGroup,
  IconButton,
  List,
  ListItem,
  ListItemButton,
  ListItemText,
  Tooltip,
  Typography,
} from "@mui/material";
import RefreshIcon from "@mui/icons-material/Refresh";
import dealsApi from "@/features/deals/api/dealsApi";

function DealsList() {
  const [myOnly, setMyOnly] = useState(false);
  const [openOnly, setOpenOnly] = useState(false);

  const { data, isLoading, isFetching, error, refetch } = dealsApi.useGetDealsQuery({
    my: myOnly || undefined,
    open: openOnly || undefined,
  });

  if (isLoading) {
    return (
      <Box display="flex" justifyContent="center" py={6}>
        <CircularProgress />
      </Box>
    );
  }

  if (error) {
    return <Alert severity="error">Не удалось загрузить сделки</Alert>;
  }

  if (!data) {
    return <Alert severity="info">Список сделок недоступен</Alert>;
  }

  return (
    <Box>
      <Box display="flex" alignItems="center" gap={2} mb={2} flexWrap="wrap">
        <FormGroup row>
          <FormControlLabel
            control={
              <Checkbox
                checked={myOnly}
                onChange={(e) => setMyOnly(e.target.checked)}
                size="small"
              />
            }
            label="Только мои"
          />
          <FormControlLabel
            control={
              <Checkbox
                checked={openOnly}
                onChange={(e) => setOpenOnly(e.target.checked)}
                size="small"
              />
            }
            label="Только открытые"
          />
        </FormGroup>

        <Tooltip title="Обновить">
          <span>
            <IconButton onClick={() => refetch()} disabled={isFetching} size="small">
              <RefreshIcon />
            </IconButton>
          </span>
        </Tooltip>
      </Box>

      {data.data.length === 0 ? (
        <Typography color="text.secondary" textAlign="center" py={4}>
          Сделок пока нет
        </Typography>
      ) : (
        <List disablePadding>
          {data.data.map((dealId) => (
            <ListItem key={dealId} disablePadding divider>
              <ListItemButton component={RouterLink} to={`/deals/${dealId}`}>
                <ListItemText
                  primary={dealId}
                  primaryTypographyProps={{ variant: "body2", fontFamily: "monospace" }}
                />
              </ListItemButton>
            </ListItem>
          ))}
        </List>
      )}
    </Box>
  );
}

export default DealsList;
