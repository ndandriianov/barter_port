import {combineReducers} from "@reduxjs/toolkit";
import authReducer from "@/features/auth/model/authSlice.ts";
import authApi from "@/features/auth/api/authApi.ts";
import itemsApi from "@/features/items/api/itemsApi.ts";

export const rootReducer = combineReducers({
  auth: authReducer,
  [authApi.reducerPath]: authApi.reducer,
  [itemsApi.reducerPath]: itemsApi.reducer,
})

export type RootState = ReturnType<typeof rootReducer>;