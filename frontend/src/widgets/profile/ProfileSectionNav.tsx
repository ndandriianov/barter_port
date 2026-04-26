import { Button, ButtonGroup } from "@mui/material";
import { Link as RouterLink, useLocation } from "react-router-dom";
import { appRoutes } from "@/shared/config/appRoutes.ts";

const navItems = [
  {
    label: "Обзор",
    to: appRoutes.profile.home,
    isActive: (pathname: string) => pathname === appRoutes.profile.home,
  },
  {
    label: "Личные данные",
    to: appRoutes.profile.account,
    isActive: (pathname: string) => pathname.startsWith(appRoutes.profile.account),
  },
  {
    label: "Репутация",
    to: appRoutes.profile.reputation,
    isActive: (pathname: string) => pathname.startsWith("/app/profile/reputation"),
  },
  {
    label: "Подписки",
    to: appRoutes.profile.networkSubscriptions,
    isActive: (pathname: string) => pathname.startsWith("/app/profile/network"),
  },
  {
    label: "Отзывы",
    to: appRoutes.profile.reviewsMine,
    isActive: (pathname: string) => pathname.startsWith("/app/profile/reviews"),
  },
  {
    label: "Статистика",
    to: appRoutes.profile.statistics,
    isActive: (pathname: string) => pathname === appRoutes.profile.statistics,
  },
] as const;

function ProfileSectionNav() {
  const location = useLocation();

  return (
    <ButtonGroup
      variant="text"
      sx={{
        alignSelf: "flex-start",
        bgcolor: "background.paper",
        borderRadius: 999,
        p: 0.75,
        boxShadow: "0 10px 30px rgba(15, 23, 42, 0.08)",
        flexWrap: "wrap",
        rowGap: 0.75,
      }}
    >
      {navItems.map((item) => (
        <Button
          key={item.to}
          component={RouterLink}
          to={item.to}
          variant={item.isActive(location.pathname) ? "contained" : "text"}
        >
          {item.label}
        </Button>
      ))}
    </ButtonGroup>
  );
}

export default ProfileSectionNav;
