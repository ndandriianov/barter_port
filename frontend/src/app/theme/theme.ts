import { createTheme } from "@mui/material";
import palette from "./palette";

const theme = createTheme({
  palette,
  typography: {
    fontFamily: '"Roboto", "Helvetica", "Arial", sans-serif',
  },
  shape: {
    borderRadius: 8,
  },
});

export default theme;
