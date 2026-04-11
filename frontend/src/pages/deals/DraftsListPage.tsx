import { useSearchParams } from "react-router-dom";
import { Box, Button, ButtonGroup, Paper, Typography } from "@mui/material";
import DraftsList from "@/widgets/deals/DraftsList";

type DraftsTab = "all" | "others" | "mine";

const tabMeta: Record<DraftsTab, { title: string; description: string }> = {
  all: {
    title: "Все",
    description: "Все ваши черновики договоров: и входящие предложения, и ваши собственные исходящие приложения.",
  },
  others: {
    title: "Входящие предложения",
    description: "Черновики других пользователей, в которых участвуют ваши объявления.",
  },
  mine: {
    title: "Исходящие приложения",
    description: "Черновики, которые создали вы. Здесь удобно следить за собственными заготовками сделок.",
  },
};

function isDraftsTab(value: string | null): value is DraftsTab {
  return value === "all" || value === "others" || value === "mine";
}

function DraftsListPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const rawTab = searchParams.get("tab");
  const tab: DraftsTab = isDraftsTab(rawTab) ? rawTab : "all";

  const handleTabChange = (nextTab: DraftsTab) => {
    const nextParams = new URLSearchParams(searchParams);
    nextParams.set("tab", nextTab);
    setSearchParams(nextParams);
  };

  return (
    <Box maxWidth={1200} mx="auto">
      <Box mb={3}>
        <Typography variant="h4" fontWeight={700} mb={1}>
          Черновики договоров
        </Typography>
        <Typography variant="body1" color="text.secondary">
          {tabMeta[tab].description}
        </Typography>
      </Box>

      <Paper variant="outlined" sx={{ p: 1, mb: 3 }}>
        <ButtonGroup
          fullWidth
          variant="text"
          sx={{ display: "grid", gridTemplateColumns: { xs: "1fr", md: "repeat(3, 1fr)" } }}
        >
          <Button
            variant={tab === "all" ? "contained" : "text"}
            onClick={() => handleTabChange("all")}
          >
            Все
          </Button>
          <Button
            variant={tab === "others" ? "contained" : "text"}
            onClick={() => handleTabChange("others")}
          >
            Входящие предложения
          </Button>
          <Button
            variant={tab === "mine" ? "contained" : "text"}
            onClick={() => handleTabChange("mine")}
          >
            Исходящие приложения
          </Button>
        </ButtonGroup>
      </Paper>

      <Box mb={2}>
        <Typography variant="h5" fontWeight={700}>
          {tabMeta[tab].title}
        </Typography>
      </Box>

      <DraftsList
        mode={tab}
        selectedOfferId={searchParams.get("offerId") ?? ""}
        onSelectedOfferIdChange={(offerId) => {
          const nextParams = new URLSearchParams(searchParams);
          if (offerId) {
            nextParams.set("offerId", offerId);
          } else {
            nextParams.delete("offerId");
          }
          nextParams.set("tab", tab);
          setSearchParams(nextParams);
        }}
      />
    </Box>
  );
}

export default DraftsListPage;
