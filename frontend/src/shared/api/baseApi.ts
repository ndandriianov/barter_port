import {fetchBaseQuery} from "@reduxjs/toolkit/query/react";
import type {BaseQueryFn} from "@reduxjs/toolkit/query";
import type {FetchArgs, FetchBaseQueryError} from "@reduxjs/toolkit/query";
import {setCredentials, logout} from "@/features/auth/model/authSlice";
import type {RootState} from "@/app/store/rootReducer.ts";
import authApi from "@/features/auth/api/authApi.ts";

type RefreshResponse = {
  access_token: string;
};

const rawBaseQuery = fetchBaseQuery({
  baseUrl: "http://localhost:80",
  credentials: "include", // для refresh cookie
  prepareHeaders: (headers, {getState}) => {
    const token = (getState() as RootState).auth.accessToken;
    console.log("token", token);
    if (token) {
      headers.set("Authorization", `Bearer ${token}`);
    }
    return headers;
  },
});

export const baseQueryWithReauth: BaseQueryFn<string | FetchArgs, unknown, FetchBaseQueryError>
  = async (args, api, extraOptions) => {

  console.log("попытка отправить запрос")
  let result = await rawBaseQuery(args, api, extraOptions);

  if (result.error && result.error.status === 401) {
    // пробуем refresh
    console.log("попытка обновить токен")
    const refreshResult = await rawBaseQuery(
      {url: "/auth/refresh", method: "POST"},
      api,
      extraOptions
    );

    console.log("токен обновлен", refreshResult)
    if (refreshResult.data) {
      const {access_token} = refreshResult.data as RefreshResponse;
      api.dispatch(setCredentials(access_token));

      console.log("повторная попытка отправить запрос с новым токеном")
      result = await rawBaseQuery(args, api, extraOptions);
    } else {
      api.dispatch(logout());
      authApi.util.resetApiState();
    }
  }

  return result;
};
