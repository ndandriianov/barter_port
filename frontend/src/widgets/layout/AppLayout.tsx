import { useState } from "react";
import { Outlet, Link as RouterLink, useNavigate } from "react-router-dom";
import {
  AppBar,
  Badge,
  Box,
  Button,
  Container,
  Drawer,
  IconButton,
  List,
  ListItem,
  ListItemButton,
  ListItemText,
  Toolbar,
  Typography,
} from "@mui/material";
import MenuIcon from "@mui/icons-material/Menu";
import { useAppDispatch } from "@/hooks/redux";
import { performLogout } from "@/features/auth/model/logoutThunk";
import dealsApi from "@/features/deals/api/dealsApi.ts";
import usePendingReviews from "@/features/reviews/model/usePendingReviews.ts";

const navLinks = [
  { label: "Админка", to: "/admin" },
  { label: "Объявления", to: "/offers" },
  { label: "Композиты", to: "/offer-groups" },
  { label: "Сделки", to: "/deals" },
  { label: "Отзывы", to: "/reviews" },
  { label: "Черновики", to: "/deals/drafts" },
  { label: "Профиль", to: "/profile" },
  { label: "Чаты", to: "/chats" },
];

function AppLayout() {
  const [drawerOpen, setDrawerOpen] = useState(false);
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const { pendingCount } = usePendingReviews();
  const { data: drafts } = dealsApi.useGetMyDraftDealsQuery({
    createdByMe: false,
    participating: true,
  });
  const draftCount = drafts?.length ?? 0;

  const handleLogout = async () => {
    await dispatch(performLogout());
    navigate("/login");
  };

  const renderNavLabel = (label: string, to: string) => {
    if (to === "/reviews") {
      return (
        <Badge badgeContent={pendingCount} color="error" max={99}>
          <Box component="span" sx={{ pr: pendingCount > 0 ? 1 : 0 }}>
            {label}
          </Box>
        </Badge>
      );
    }

    if (to === "/deals/drafts") {
      return (
        <Badge badgeContent={draftCount} color="warning" max={99}>
          <Box component="span" sx={{ pr: draftCount > 0 ? 1 : 0 }}>
            {label}
          </Box>
        </Badge>
      );
    }

    return label;
  };

  const getNavLinkState = (to: string) => {
    if (to === "/reviews") {
      return { fromLayoutReviewsButton: true };
    }

    return undefined;
  };

  const drawer = (
    <Box sx={{ width: 240 }} role="presentation" onClick={() => setDrawerOpen(false)}>
      <List>
        {navLinks.map((link) => (
          <ListItem key={link.to} disablePadding>
            <ListItemButton component={RouterLink} to={link.to} state={getNavLinkState(link.to)}>
              <ListItemText primary={renderNavLabel(link.label, link.to)} />
            </ListItemButton>
          </ListItem>
        ))}
        <ListItem disablePadding>
          <ListItemButton onClick={handleLogout}>
            <ListItemText primary="Выйти" />
          </ListItemButton>
        </ListItem>
      </List>
    </Box>
  );

  return (
    <Box sx={{ display: "flex", flexDirection: "column", minHeight: "100vh" }}>
      <AppBar position="sticky">
        <Toolbar>
          <IconButton
            color="inherit"
            edge="start"
            sx={{ mr: 1, display: { sm: "none" } }}
            onClick={() => setDrawerOpen(true)}
          >
            <MenuIcon />
          </IconButton>

          <Typography
            variant="h6"
            component={RouterLink}
            to="/offers"
            sx={{ flexGrow: 1, textDecoration: "none", color: "inherit", fontWeight: 700 }}
          >
            Barter Port
          </Typography>

          <Box sx={{ display: { xs: "none", sm: "flex" }, gap: 1 }}>
            {navLinks.map((link) => (
              <Button
                key={link.to}
                color="inherit"
                component={RouterLink}
                to={link.to}
                state={getNavLinkState(link.to)}
              >
                {renderNavLabel(link.label, link.to)}
              </Button>
            ))}
            <Button color="inherit" variant="outlined" onClick={handleLogout} sx={{ borderColor: "rgba(255,255,255,0.5)" }}>
              Выйти
            </Button>
          </Box>
        </Toolbar>
      </AppBar>

      <Drawer open={drawerOpen} onClose={() => setDrawerOpen(false)}>
        {drawer}
      </Drawer>

      <Container component="main" maxWidth="lg" sx={{ py: 4, flexGrow: 1 }}>
        <Outlet />
      </Container>
    </Box>
  );
}

export default AppLayout;
