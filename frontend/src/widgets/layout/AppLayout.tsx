import { useMemo, useState } from "react";
import { Outlet, Link as RouterLink, useLocation, useNavigate } from "react-router-dom";
import { skipToken } from "@reduxjs/toolkit/query";
import {
  AppBar,
  Badge,
  Box,
  Button,
  Container,
  Divider,
  Drawer,
  IconButton,
  List,
  ListItemButton,
  ListItemText,
  Stack,
  Toolbar,
  Typography,
} from "@mui/material";
import MenuIcon from "@mui/icons-material/Menu";
import { useAppDispatch } from "@/hooks/redux";
import { performLogout } from "@/features/auth/model/logoutThunk";
import usersApi from "@/features/users/api/usersApi.ts";
import dealsApi from "@/features/deals/api/dealsApi.ts";
import offersApi from "@/features/offers/api/offersApi.ts";
import useDealActionQueue from "@/features/deals/model/useDealActionQueue.ts";
import { appRoutes, getSectionFromPathname, type AppSection } from "@/shared/config/appRoutes.ts";

const sectionMeta: Record<AppSection, { label: string; description: string }> = {
  market: {
    label: "Объявления",
    description: "Найти, откликнуться, опубликовать и управлять своими материалами.",
  },
  deals: {
    label: "Сделки",
    description: "Черновики, активные сделки, история сделок и отзывы на товар после завершения сделки",
  },
  messages: {
    label: "Сообщения",
    description: "Личные диалоги и переписка по сделкам в едином потоке.",
  },
  profile: {
    label: "Профиль",
    description: "Личные данные, репутация, отзывы, подписки и статистика.",
  },
  admin: {
    label: "Модерация",
    description: "Отдельная admin-only зона для очередей и системных сущностей.",
  },
};

function AppLayout() {
  const [drawerOpen, setDrawerOpen] = useState(false);
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const location = useLocation();
  const activeSection = getSectionFromPathname(location.pathname);
  const { data: currentUser } = usersApi.useGetCurrentUserQuery();
  const { totalActionCount } = useDealActionQueue();
  const { data: pendingReports = [] } = offersApi.useListAdminOfferReportsQuery(
    currentUser?.isAdmin ? "Pending" : skipToken,
  );
  const { data: failureDeals = [] } = dealsApi.useGetDealsForFailureReviewQuery(
    currentUser?.isAdmin ? undefined : skipToken,
  );

  const navLinks = useMemo(() => {
    const base = [
      { key: "market" as const, label: "Объявления", to: appRoutes.market.home, badge: 0 },
      { key: "deals" as const, label: "Сделки", to: appRoutes.deals.home, badge: totalActionCount },
      { key: "messages" as const, label: "Сообщения", to: appRoutes.messages.home, badge: 0 },
      { key: "profile" as const, label: "Профиль", to: appRoutes.profile.home, badge: 0 },
    ];

    if (currentUser?.isAdmin) {
      return [
        ...base,
        {
          key: "admin" as const,
          label: "Модерация",
          to: appRoutes.admin.home,
          badge: pendingReports.length + failureDeals.length,
        },
      ];
    }

    return base;
  }, [currentUser?.isAdmin, failureDeals.length, pendingReports.length, totalActionCount]);

  const handleLogout = async () => {
    await dispatch(performLogout());
    navigate(appRoutes.auth.login);
  };

  const renderNavLabel = (label: string, badge: number) => (
    <Badge color="warning" badgeContent={badge > 0 ? badge : 0} invisible={badge <= 0} max={99}>
      <Box component="span" sx={{ pr: badge > 0 ? 1 : 0 }}>
        {label}
      </Box>
    </Badge>
  );

  return (
    <Box
      sx={{
        minHeight: "100vh",
        bgcolor: "background.default",
        backgroundImage:
          "radial-gradient(circle at top left, rgba(15,118,110,0.08), transparent 25%), radial-gradient(circle at top right, rgba(194,109,31,0.08), transparent 22%)",
      }}
    >
      <AppBar position="sticky">
        <Toolbar sx={{ minHeight: 76 }}>
          <IconButton
            edge="start"
            sx={{ mr: 1, display: { md: "none" } }}
            onClick={() => setDrawerOpen(true)}
          >
            <MenuIcon />
          </IconButton>

          <Box
            component={RouterLink}
            to={appRoutes.market.home}
            sx={{ textDecoration: "none", color: "inherit", display: "flex", flexDirection: "column", mr: 3 }}
          >
            <Typography fontWeight={900} letterSpacing={-0.6}>
              Barter Port
            </Typography>
            <Typography variant="caption" color="text.secondary">
              Workflow-first marketplace
            </Typography>
          </Box>

          <Box sx={{ display: { xs: "none", md: "flex" }, gap: 1, flexGrow: 1 }}>
            {navLinks.map((link) => {
              const isActive = link.key === activeSection;

              return (
                <Button
                  key={link.key}
                  component={RouterLink}
                  to={link.to}
                  variant={isActive ? "contained" : "text"}
                  color={isActive ? "primary" : "inherit"}
                >
                  {renderNavLabel(link.label, link.badge)}
                </Button>
              );
            })}
          </Box>

          <Button onClick={handleLogout} variant="outlined" sx={{ borderColor: "rgba(23,33,43,0.12)" }}>
            Выйти
          </Button>
        </Toolbar>
      </AppBar>

      <Drawer open={drawerOpen} onClose={() => setDrawerOpen(false)}>
        <Box sx={{ width: 300, p: 2 }}>
          <Typography variant="h6" fontWeight={900} px={1} py={1.5}>
            Разделы
          </Typography>
          <List sx={{ display: "flex", flexDirection: "column", gap: 0.5 }}>
            {navLinks.map((link) => (
              <ListItemButton
                key={link.key}
                component={RouterLink}
                to={link.to}
                onClick={() => setDrawerOpen(false)}
                selected={link.key === activeSection}
                sx={{ borderRadius: 3 }}
              >
                <ListItemText primary={renderNavLabel(link.label, link.badge)} />
              </ListItemButton>
            ))}
          </List>
          <Divider sx={{ my: 2 }} />
          <Button fullWidth onClick={handleLogout} variant="outlined">
            Выйти
          </Button>
        </Box>
      </Drawer>

      <Box
        sx={{
          borderBottom: "1px solid rgba(15, 23, 42, 0.08)",
          background: "linear-gradient(180deg, rgba(255,255,255,0.72) 0%, rgba(255,255,255,0.36) 100%)",
        }}
      >
        <Container maxWidth="xl" sx={{ py: 3.5 }}>
          <Stack spacing={0.5}>
            <Typography variant="overline" color="primary.main">
              {sectionMeta[activeSection].label}
            </Typography>
            <Typography variant="h5">{sectionMeta[activeSection].description}</Typography>
          </Stack>
        </Container>
      </Box>

      <Container component="main" maxWidth="xl" sx={{ py: 4.5 }}>
        <Outlet />
      </Container>
    </Box>
  );
}

export default AppLayout;
