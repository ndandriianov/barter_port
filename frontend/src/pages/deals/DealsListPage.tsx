import { useMemo } from "react";
import { Link as RouterLink, useSearchParams } from "react-router-dom";
import { Box, Button, ButtonGroup, Paper, Typography } from "@mui/material";
import dealsApi from "@/features/deals/api/dealsApi.ts";
import usersApi from "@/features/users/api/usersApi.ts";
import DealsList from "@/widgets/deals/DealsList";
import { appRoutes } from "@/shared/config/appRoutes.ts";
import { dealsListModeConfig, type DealsListMode } from "@/pages/deals/dealsListModes.ts";

interface DealsListPageProps {
  mode: DealsListMode;
}

function DealsListPage({ mode }: DealsListPageProps) {
  const [searchParams, setSearchParams] = useSearchParams();
  const currentPage = dealsListModeConfig[mode];
  const selections = currentPage.selections ?? [];
  const rawSelectionKey = searchParams.get("status");
  const { data: currentUser } = usersApi.useGetCurrentUserQuery();
  const {
    data: deals = [],
    isLoading,
    isFetching,
    error,
    refetch,
  } = dealsApi.useGetDealsQuery({
    my: currentPage.query.myOnly || undefined,
    open: currentPage.query.openOnly || undefined,
  });

  const scopedDeals = useMemo(() => {
    if (!currentPage.query.excludeCurrentUser || !currentUser) {
      return deals;
    }

    return deals.filter((deal) => !deal.participants.includes(currentUser.id));
  }, [currentPage.query.excludeCurrentUser, currentUser, deals]);

  const selectedSelection = selections.find((selection) => selection.key === rawSelectionKey) ?? selections[0];
  const selectedStatuses = selectedSelection?.statuses ?? currentPage.defaultStatuses;
  const filteredDeals = useMemo(
    () => scopedDeals.filter((deal) => selectedStatuses.includes(deal.status)),
    [scopedDeals, selectedStatuses],
  );
  const selectionCounts = useMemo(
    () =>
      selections.reduce<Record<string, number>>((acc, selection) => {
        acc[selection.key] = scopedDeals.filter((deal) => selection.statuses.includes(deal.status)).length;
        return acc;
      }, {}),
    [scopedDeals, selections],
  );

  const handleSelectionChange = (nextSelectionKey: string) => {
    const nextParams = new URLSearchParams(searchParams);
    nextParams.set("status", nextSelectionKey);
    setSearchParams(nextParams);
  };
  const pageTitle = selectedSelection?.title ?? currentPage.title;
  const pageDescription = selectedSelection?.description ?? currentPage.description;
  const emptyMessage = mode === "joinable"
    ? "Подходящих сделок для присоединения пока нет"
    : `Сделок в разделе «${pageTitle}» пока нет`;

  return (
    <Box maxWidth={1200} mx="auto">
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3} flexWrap="wrap" gap={1}>
        <Box>
          <Typography variant="overline" color={mode === "history" ? "secondary.main" : "info.main"}>
            {currentPage.eyebrow}
          </Typography>
          <Typography variant="h4" fontWeight={700} mb={1}>
            {currentPage.title}
          </Typography>
          <Typography variant="body1" color="text.secondary">
            {pageDescription}
          </Typography>
        </Box>
        <Box display="flex" gap={1}>
          <Button
            variant={mode === "active" ? "contained" : "outlined"}
            component={RouterLink}
            to={appRoutes.deals.active}
          >
            Активные
          </Button>
          <Button
            variant={mode === "joinable" ? "contained" : "outlined"}
            component={RouterLink}
            to={appRoutes.deals.joinable}
          >
            Можно присоединиться
          </Button>
          <Button
            variant={mode === "history" ? "contained" : "outlined"}
            component={RouterLink}
            to={appRoutes.deals.history}
          >
            История
          </Button>
        </Box>
      </Box>

      {selections.length > 1 && (
        <Paper variant="outlined" sx={{ p: 1, mb: 3 }}>
          <ButtonGroup
            fullWidth
            variant="text"
            sx={{ display: "grid", gridTemplateColumns: { xs: "1fr", md: "repeat(3, 1fr)", xl: "repeat(4, 1fr)" } }}
          >
            {selections.map((selection) => (
              <Button
                key={selection.key}
                variant={selectedSelection?.key === selection.key ? "contained" : "text"}
                onClick={() => handleSelectionChange(selection.key)}
              >
                {selection.title} {selectionCounts[selection.key] ?? 0}
              </Button>
            ))}
          </ButtonGroup>
        </Paper>
      )}

      <Typography variant="h5" fontWeight={700} mb={2}>
        {pageTitle}
      </Typography>

      <DealsList
        deals={filteredDeals}
        isLoading={isLoading}
        isFetching={isFetching}
        hasError={Boolean(error)}
        onRefresh={refetch}
        emptyMessage={emptyMessage}
        showStatusSections={selectedStatuses.length > 1}
      />
    </Box>
  );
}

export default DealsListPage;
