import {configureStore} from "@reduxjs/toolkit";
import authApi from "@/features/auth/api/authApi";
import offersApi from "@/features/offers/api/offersApi";
import {rootReducer} from "@/app/store/rootReducer.ts";

export const store = configureStore({
  reducer: rootReducer,
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware().concat(authApi.middleware, offersApi.middleware),
});

export type AppDispatch = typeof store.dispatch;