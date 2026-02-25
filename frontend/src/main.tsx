import ReactDOM from "react-dom/client";
import { StoreProvider } from "@/app/providers/StoreProvider";
import { App } from "@/app/App";

ReactDOM.createRoot(document.getElementById("root")!).render(
  <StoreProvider>
    <App />
  </StoreProvider>
);