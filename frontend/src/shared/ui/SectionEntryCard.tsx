import { Card, CardActionArea, CardContent, Chip, Stack, Typography } from "@mui/material";
import { Link as RouterLink } from "react-router-dom";
import type { ReactNode } from "react";

interface SectionEntryCardProps {
  to: string;
  icon: ReactNode;
  title: string;
  description: string;
  badge?: string | number | null;
  accent?: "primary" | "secondary" | "info" | "success" | "warning";
}

function SectionEntryCard({
  to,
  icon,
  title,
  description,
  badge,
  accent = "primary",
}: SectionEntryCardProps) {
  return (
    <Card
      variant="outlined"
      sx={{
        height: "100%",
        background:
          accent === "secondary"
            ? "linear-gradient(180deg, rgba(194,109,31,0.12) 0%, rgba(194,109,31,0.03) 100%)"
            : accent === "info"
              ? "linear-gradient(180deg, rgba(14,116,144,0.12) 0%, rgba(14,116,144,0.03) 100%)"
              : accent === "success"
                ? "linear-gradient(180deg, rgba(22,163,74,0.12) 0%, rgba(22,163,74,0.03) 100%)"
                : accent === "warning"
                  ? "linear-gradient(180deg, rgba(234,88,12,0.12) 0%, rgba(234,88,12,0.03) 100%)"
                  : "linear-gradient(180deg, rgba(15,118,110,0.12) 0%, rgba(15,118,110,0.03) 100%)",
      }}
    >
      <CardActionArea component={RouterLink} to={to} sx={{ height: "100%", alignItems: "stretch" }}>
        <CardContent sx={{ height: "100%" }}>
          <Stack spacing={2} height="100%">
            <Stack direction="row" justifyContent="space-between" alignItems="flex-start" gap={2}>
              <Typography color={`${accent}.main`}>{icon}</Typography>
              {badge ? <Chip label={badge} color={accent} size="small" /> : null}
            </Stack>
            <div>
              <Typography variant="h6" fontWeight={800} mb={1}>
                {title}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                {description}
              </Typography>
            </div>
          </Stack>
        </CardContent>
      </CardActionArea>
    </Card>
  );
}

export default SectionEntryCard;
