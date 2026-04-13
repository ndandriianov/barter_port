import { Avatar, Box, Typography } from "@mui/material";
import PersonOutlineIcon from "@mui/icons-material/PersonOutline";

interface UserAvatarLabelProps {
  name?: string | null;
  avatarUrl?: string | null;
  size?: number;
  textVariant?: "caption" | "body2" | "body1";
  fontWeight?: number;
}

function UserAvatarLabel({
  name,
  avatarUrl,
  size = 28,
  textVariant = "body2",
  fontWeight = 500,
}: UserAvatarLabelProps) {
  const normalizedName = name?.trim() || "Имя не указано";
  const initial = normalizedName[0]?.toUpperCase();

  return (
    <Box display="flex" alignItems="center" gap={1}>
      <Avatar
        src={avatarUrl || undefined}
        alt={normalizedName}
        sx={{ width: size, height: size, bgcolor: "action.selected", fontSize: Math.max(12, size / 2.1) }}
      >
        {avatarUrl ? initial : <PersonOutlineIcon sx={{ fontSize: Math.max(16, size * 0.7) }} color="action" />}
      </Avatar>
      <Typography variant={textVariant} fontWeight={fontWeight}>
        {normalizedName}
      </Typography>
    </Box>
  );
}

export default UserAvatarLabel;
