import { Box, Button, Card, CardContent, Chip, Stack, Typography } from "@mui/material";
import HubOutlinedIcon from "@mui/icons-material/HubOutlined";
import AltRouteOutlinedIcon from "@mui/icons-material/AltRouteOutlined";
import { Link as RouterLink } from "react-router-dom";
import type { OfferGroup } from "@/features/offer-groups/model/types.ts";
import usersApi from "@/features/users/api/usersApi.ts";
import {
  getOfferGroupOwnerId,
  getOfferGroupOwnerName,
  getOfferGroupVariantCount,
} from "@/features/offer-groups/model/utils.ts";
import UserAvatarLabel from "@/shared/UserAvatarLabel.tsx";

interface OfferGroupCardProps {
  offerGroup: OfferGroup;
  href?: string;
}

function OfferGroupCard({ offerGroup, href }: OfferGroupCardProps) {
  const ownerId = getOfferGroupOwnerId(offerGroup);
  const ownerName = getOfferGroupOwnerName(offerGroup);
  const variantCount = getOfferGroupVariantCount(offerGroup);
  const { data: author } = usersApi.useGetUserByIdQuery(ownerId ?? "", {
    skip: !ownerId,
  });

  return (
    <Card
      variant="outlined"
      sx={{
        height: "100%",
        display: "flex",
        flexDirection: "column",
        background:
          "linear-gradient(180deg, rgba(7, 116, 129, 0.04) 0%, rgba(7, 116, 129, 0.01) 100%)",
      }}
    >
      <CardContent sx={{ display: "flex", flexDirection: "column", gap: 2, flexGrow: 1 }}>
        <Box display="flex" gap={1} flexWrap="wrap">
          <Chip icon={<HubOutlinedIcon />} label={`${offerGroup.units.length} AND-блок(ов)`} size="small" />
          <Chip
            icon={<AltRouteOutlinedIcon />}
            label={`${variantCount} offer-вариантов`}
            size="small"
            color="info"
            variant="outlined"
          />
        </Box>

        <Box>
          <Typography variant="h6" fontWeight={700} gutterBottom>
            {offerGroup.name}
          </Typography>
          <UserAvatarLabel
            name={author?.name ?? ownerName}
            avatarUrl={author?.avatarUrl}
            size={30}
            textVariant="body2"
            fontWeight={400}
          />
        </Box>

        <Typography
          variant="body2"
          color="text.secondary"
          sx={{
            display: "-webkit-box",
            WebkitLineClamp: 3,
            WebkitBoxOrient: "vertical",
            overflow: "hidden",
            minHeight: 60,
          }}
        >
          {offerGroup.description?.trim() || "Описание не указано"}
        </Typography>

        <Stack spacing={1} sx={{ mt: "auto" }}>
          {offerGroup.units.map((unit, index) => (
            <Box
              key={unit.id}
              sx={{
                px: 1.5,
                py: 1,
                borderRadius: 2,
                bgcolor: "background.paper",
                border: "1px solid",
                borderColor: "divider",
              }}
            >
              <Typography variant="caption" color="text.secondary">
                Unit {index + 1}
              </Typography>
              <Typography variant="body2" fontWeight={600}>
                {unit.offers.map((offer) => offer.name).join(" / ")}
              </Typography>
            </Box>
          ))}
        </Stack>

        {href && (
          <Box display="flex" justifyContent="flex-start">
            <Button component={RouterLink} to={href} variant="outlined">
              Открыть
            </Button>
          </Box>
        )}
      </CardContent>
    </Card>
  );
}

export default OfferGroupCard;
