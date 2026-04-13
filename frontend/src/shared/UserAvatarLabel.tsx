import { useState } from "react";
import { Link as RouterLink } from "react-router-dom";
import { Avatar, Box, Dialog, DialogContent, Typography } from "@mui/material";
import PersonOutlineIcon from "@mui/icons-material/PersonOutline";
import usersApi from "@/features/users/api/usersApi.ts";

interface UserAvatarLabelProps {
  userId?: string;
  name?: string | null;
  avatarUrl?: string | null;
  size?: number;
  textVariant?: "caption" | "body2" | "body1";
  fontWeight?: number;
}

function UserAvatarLabel({
  userId,
  name,
  avatarUrl,
  size = 28,
  textVariant = "body2",
  fontWeight = 500,
}: UserAvatarLabelProps) {
  const [isAvatarOpen, setIsAvatarOpen] = useState(false);
  const { data: currentUser } = usersApi.useGetCurrentUserQuery();
  const normalizedName = name?.trim() || "Имя не указано";
  const initial = normalizedName[0]?.toUpperCase();
  const profileHref = userId
    ? currentUser?.id === userId
      ? "/profile"
      : `/users/${userId}`
    : undefined;

  return (
    <Box display="flex" alignItems="center" gap={1}>
      <Avatar
        src={avatarUrl || undefined}
        alt={normalizedName}
        onClick={avatarUrl ? () => setIsAvatarOpen(true) : undefined}
        sx={{
          width: size,
          height: size,
          bgcolor: "action.selected",
          fontSize: Math.max(12, size / 2.1),
          cursor: avatarUrl ? "zoom-in" : "default",
        }}
      >
        {avatarUrl ? initial : <PersonOutlineIcon sx={{ fontSize: Math.max(16, size * 0.7) }} color="action" />}
      </Avatar>

      {profileHref ? (
        <Typography
          component={RouterLink}
          to={profileHref}
          variant={textVariant}
          fontWeight={fontWeight}
          sx={{
            color: "inherit",
            textDecoration: "none",
            "&:hover": {
              textDecoration: "underline",
            },
          }}
        >
          {normalizedName}
        </Typography>
      ) : (
        <Typography variant={textVariant} fontWeight={fontWeight}>
          {normalizedName}
        </Typography>
      )}

      <Dialog
        open={isAvatarOpen}
        onClose={() => setIsAvatarOpen(false)}
        maxWidth="sm"
        fullWidth
      >
        <DialogContent sx={{ p: 1.5, bgcolor: "common.black" }}>
          {avatarUrl && (
            <Box
              component="img"
              src={avatarUrl}
              alt={normalizedName}
              sx={{
                width: "100%",
                maxHeight: "85vh",
                objectFit: "contain",
                display: "block",
              }}
            />
          )}
        </DialogContent>
      </Dialog>
    </Box>
  );
}

export default UserAvatarLabel;
