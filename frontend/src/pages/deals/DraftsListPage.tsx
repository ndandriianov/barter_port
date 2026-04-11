import { useSearchParams } from "react-router-dom";
import { Box, Typography } from "@mui/material";
import DraftsList from "@/widgets/deals/DraftsList";

function DraftsListPage() {
  const [searchParams, setSearchParams] = useSearchParams();

  return (
    <Box>
      <Typography variant="h4" fontWeight={700} mb={3}>
        Мои черновые договоры
      </Typography>
      <DraftsList
        selectedOfferId={searchParams.get("offerId") ?? ""}
        onSelectedOfferIdChange={(offerId) => {
          const nextParams = new URLSearchParams(searchParams);
          if (offerId) {
            nextParams.set("offerId", offerId);
          } else {
            nextParams.delete("offerId");
          }
          setSearchParams(nextParams);
        }}
      />
    </Box>
  );
}

export default DraftsListPage;
