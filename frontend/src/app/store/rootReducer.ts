import {combineReducers} from "@reduxjs/toolkit";
import authReducer from "@/features/auth/model/authSlice.ts";
import authApi from "@/features/auth/api/authApi.ts";
import offersApi from "@/features/offers/api/offersApi.ts";
import dealsApi from "@/features/deals/api/dealsApi.ts";
import usersApi from "@/features/users/api/usersApi.ts";

export const rootReducer = combineReducers({
  auth: authReducer,
  [authApi.reducerPath]: authApi.reducer,
  [offersApi.reducerPath]: offersApi.reducer,
  [dealsApi.reducerPath]: dealsApi.reducer,
  [usersApi.reducerPath]: usersApi.reducer,
})

export type RootState = ReturnType<typeof rootReducer>;