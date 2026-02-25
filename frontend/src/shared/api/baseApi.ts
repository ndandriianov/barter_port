import {fetchBaseQuery} from "@reduxjs/toolkit/query/react";
import type {BaseQueryFn} from "@reduxjs/toolkit/query";
import type {FetchArgs, FetchBaseQueryError} from "@reduxjs/toolkit/query";
import {type RootState} from "@/app/store/store";
import {setCredentials, logout} from "@/features/auth/model/authSlice";

const rawBaseQuery = fetchBaseQuery({
  baseUrl: "http://localhost:8080",
  credentials: "include", // для refresh cookie
  prepareHeaders: (headers, {getState}) => {
    const token = (getState() as RootState).auth.accessToken;
    if (token) {
      headers.set("Authorization", `Bearer ${token}`);
    }
    return headers;
  },
});

export const baseQueryWithReauth: BaseQueryFn<string | FetchArgs, unknown, FetchBaseQueryError>
  = async (args, api, extraOptions) => {

  let result = await rawBaseQuery(args, api, extraOptions);

  if (result.error && result.error.status === 401) {
    // пробуем refresh
    const refreshResult = await rawBaseQuery(
      {url: "/auth/refresh", method: "POST"},
      api,
      extraOptions
    );

    if (refreshResult.data) {
      const {accessToken} = refreshResult.data as { accessToken: string };
      api.dispatch(setCredentials(accessToken));

      result = await rawBaseQuery(args, api, extraOptions);
    } else {
      api.dispatch(logout());
    }
  }

  return result;
};