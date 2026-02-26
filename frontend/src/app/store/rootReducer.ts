import {combineReducers} from "@reduxjs/toolkit";
import authReducer from "@/features/auth/model/authSlice.ts";
import authApi from "@/features/auth/api/authApi.ts";

export const rootReducer = combineReducers({
  auth: authReducer,
  [authApi.reducerPath]: authApi.reducer,
})

export type RootState = ReturnType<typeof rootReducer>;