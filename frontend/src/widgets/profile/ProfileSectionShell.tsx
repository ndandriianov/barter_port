import { Box, Stack, Typography } from "@mui/material";
import type { ReactNode } from "react";
import ProfileSectionNav from "@/widgets/profile/ProfileSectionNav.tsx";

interface ProfileSectionShellProps {
  title: string;
  description: string;
  actions?: ReactNode;
  children: ReactNode;
}

function ProfileSectionShell({
  title,
  description,
  actions,
  children,
}: ProfileSectionShellProps) {
  return (
    <Stack spacing={3}>
      <Box display="flex" justifyContent="space-between" alignItems="flex-start" gap={2} flexWrap="wrap">
        <Box>
          <Typography variant="overline" color="primary.main">
            Профиль
          </Typography>
          <Typography variant="h4" fontWeight={800} mb={1}>
            {title}
          </Typography>
          <Typography variant="body1" color="text.secondary" maxWidth={820}>
            {description}
          </Typography>
        </Box>
        {actions}
      </Box>

      <ProfileSectionNav />

      {children}
    </Stack>
  );
}

export default ProfileSectionShell;
