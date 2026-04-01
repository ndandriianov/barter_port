import { Outlet } from "react-router-dom";
import { Box, Paper } from "@mui/material";

function AuthLayout() {
  return (
    <Box
      sx={{
        minHeight: "100vh",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        bgcolor: "background.default",
        p: 2,
      }}
    >
      <Paper elevation={3} sx={{ p: 4, width: "100%", maxWidth: 440 }}>
        <Outlet />
      </Paper>
    </Box>
  );
}

export default AuthLayout;
