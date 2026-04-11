import { combineReducers } from "@reduxjs/toolkit";
import type { UnknownAction } from "@reduxjs/toolkit";
import authReducer from "@/features/auth/model/authSlice.ts";
import authApi from "@/features/auth/api/authApi.ts";
import offersApi from "@/features/offers/api/offersApi.ts";
import dealsApi from "@/features/deals/api/dealsApi.ts";
import usersApi from "@/features/users/api/usersApi.ts";
import chatsApi from "@/features/chats/api/chatsApi.ts";
import reviewsApi from "@/features/reviews/api/reviewsApi.ts";
import { resetAllApiCaches } from "@/app/store/resetCaches.ts";

const combinedReducer = combineReducers({
  auth: authReducer,
  [authApi.reducerPath]: authApi.reducer,
  [offersApi.reducerPath]: offersApi.reducer,
  [dealsApi.reducerPath]: dealsApi.reducer,
  [usersApi.reducerPath]: usersApi.reducer,
  [chatsApi.reducerPath]: chatsApi.reducer,
  [reviewsApi.reducerPath]: reviewsApi.reducer,
});

type CombinedState = ReturnType<typeof combinedReducer>;

export const rootReducer = (state: CombinedState | undefined, action: UnknownAction): CombinedState => {
  if (action.type === resetAllApiCaches.type) {
    // Сбрасываем все RTK Query кэши, передавая undefined в каждый API-редьюсер
    return combinedReducer({ auth: state?.auth } as CombinedState, action);
  }
  return combinedReducer(state, action);
};

export type RootState = ReturnType<typeof rootReducer>;
