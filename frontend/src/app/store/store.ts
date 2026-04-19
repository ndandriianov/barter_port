import {configureStore} from "@reduxjs/toolkit";
import authApi from "@/features/auth/api/authApi";
import offersApi from "@/features/offers/api/offersApi";
import dealsApi from "@/features/deals/api/dealsApi.ts";
import offerGroupsApi from "@/features/offer-groups/api/offerGroupsApi.ts";
import usersApi from "@/features/users/api/usersApi.ts";
import chatsApi from "@/features/chats/api/chatsApi.ts";
import reviewsApi from "@/features/reviews/api/reviewsApi.ts";
import statisticsApi from "@/features/statistics/api/statisticsApi.ts";
import {rootReducer} from "@/app/store/rootReducer.ts";

export const store = configureStore({
  reducer: rootReducer,
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware().concat(
      authApi.middleware,
      offersApi.middleware,
      dealsApi.middleware,
      offerGroupsApi.middleware,
      usersApi.middleware,
      chatsApi.middleware,
      reviewsApi.middleware,
      statisticsApi.middleware,
    ),
});

export type AppDispatch = typeof store.dispatch;
