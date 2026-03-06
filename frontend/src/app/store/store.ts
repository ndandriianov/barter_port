import {configureStore} from "@reduxjs/toolkit";
import authApi from "@/features/auth/api/authApi";
import itemsApi from "@/features/items/api/itemsApi";
import {rootReducer} from "@/app/store/rootReducer.ts";

export const store = configureStore({
  reducer: rootReducer,
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware().concat(authApi.middleware, itemsApi.middleware),
});

export type AppDispatch = typeof store.dispatch;