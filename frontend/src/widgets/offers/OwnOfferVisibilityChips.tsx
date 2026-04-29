import { Chip } from "@mui/material";
import type { Offer } from "@/features/offers/model/types.ts";

interface OwnOfferVisibilityChipsProps {
  offer: Pick<Offer, "hiddenByAuthor" | "isHidden">;
  size?: "small" | "medium";
}

function OwnOfferVisibilityChips({ offer, size = "small" }: OwnOfferVisibilityChipsProps) {
  return (
    <>
      {offer.hiddenByAuthor && (
        <Chip
          label="Скрыто автором"
          size={size}
          color="warning"
          variant="filled"
        />
      )}
      {offer.isHidden && (
        <Chip
          label="Скрыто модератором"
          size={size}
          color="error"
          variant="filled"
        />
      )}
    </>
  );
}

export default OwnOfferVisibilityChips;
