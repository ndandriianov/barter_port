import { ThemeProvider, CssBaseline } from "@mui/material";
import AppRouter from "@/shared/config/Router";
import theme from "@/app/theme/theme";

function App() {
  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <AppRouter />
    </ThemeProvider>
  );
}

export default App;
