import {Provider} from "react-redux";
import {store} from "../store/store";
import * as React from "react";

interface StateProviderProps {
  children: React.ReactNode;
}

function StoreProvider({children}: StateProviderProps) {
  return <Provider store={store}>{children}</Provider>;
};

export default StoreProvider;