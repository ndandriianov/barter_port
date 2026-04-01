import {combineReducers} from "@reduxjs/toolkit";
import authReducer from "@/features/auth/model/authSlice.ts";
import authApi from "@/features/auth/api/authApi.ts";
import offersApi from "@/features/offers/api/offersApi.ts";

export const rootReducer = combineReducers({
  auth: authReducer,
  [authApi.reducerPath]: authApi.reducer,
  [offersApi.reducerPath]: offersApi.reducer,
})

export type RootState = ReturnType<typeof rootReducer>;