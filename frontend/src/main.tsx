import ReactDOM from "react-dom/client";
import App from "@/app/App";
import {Provider} from "react-redux";
import {store} from "@/app/store/store.ts";

ReactDOM.createRoot(document.getElementById("root")!).render(
  <Provider store={store}>
    <App />
  </Provider>
);