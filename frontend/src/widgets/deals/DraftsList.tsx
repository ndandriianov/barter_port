import { Link as RouterLink } from "react-router-dom";
import {
  Alert,
  Box,
  CircularProgress,
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

function DraftsList() {
  const { data, isLoading, error, refetch, isFetching } = dealsApi.useGetMyDraftDealsQuery({
    createdByMe: false,
    participating: true,
  });

  if (isLoading) {
    return (
      <Box display="flex" justifyContent="center" py={6}>
        <CircularProgress />
      </Box>
    );
  }

  if (error) {
    return <Alert severity="error">Не удалось загрузить черновики</Alert>;
  }

  if (!data) {
    return <Alert severity="info">Черновики недоступны</Alert>;
  }

  return (
    <Box>
      <Box display="flex" justifyContent="flex-end" mb={1}>
        <Tooltip title="Обновить">
          <span>
            <IconButton onClick={() => refetch()} disabled={isFetching} size="small">
              <RefreshIcon />
            </IconButton>
          </span>
        </Tooltip>
      </Box>

      {data.length === 0 ? (
        <Typography color="text.secondary" textAlign="center" py={4}>
          У вас пока нет черновых договоров
        </Typography>
      ) : (
        <List disablePadding>
          {data.map((draft) => (
            <ListItem key={draft.id} disablePadding divider>
              <ListItemButton component={RouterLink} to={`/deals/drafts/${draft.id}`}>
                <ListItemText
                  primary={draft.id}
                  secondary={draft.participants.length > 0 ? `Участники: ${draft.participants.join(", ")}` : "Без участников"}
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

export default DraftsList;
