import {fetchBaseQuery} from "@reduxjs/toolkit/query/react";
import type {BaseQueryApi, BaseQueryFn} from "@reduxjs/toolkit/query";
import type {FetchArgs, FetchBaseQueryError} from "@reduxjs/toolkit/query";
import {sessionExpired, setCredentials} from "@/features/auth/model/authSlice";
import type {RootState} from "@/app/store/rootReducer.ts";
import { resetAllApiCaches } from "@/app/store/resetCaches.ts";

type RefreshResponse = {
  access_token: string;
};

const SESSION_EXPIRED_MESSAGE = "Сессия истекла. Войдите снова.";
const REAUTH_EXCLUDED_URLS = new Set([
  "/auth/login",
  "/auth/logout",
  "/auth/refresh",
  "/auth/register",
  "/auth/request-password-reset",
  "/auth/reset-password",
  "/auth/verify-email",
]);

let refreshPromise: Promise<boolean> | null = null;

const rawBaseQuery = fetchBaseQuery({
  baseUrl: import.meta.env.VITE_API_BASE_URL ?? "",
  credentials: "include", // для refresh cookie
  paramsSerializer: (params) => {
    const searchParams = new URLSearchParams();

    for (const [key, value] of Object.entries(params)) {
      if (value === undefined || value === null) {
        continue;
      }

      if (Array.isArray(value)) {
        for (const item of value) {
          searchParams.append(key, String(item));
        }
        continue;
      }

      searchParams.set(key, String(value));
    }

    return searchParams.toString();
  },
  prepareHeaders: (headers, {getState}) => {
    const token = (getState() as RootState).auth.accessToken;
    if (token) {
      headers.set("Authorization", `Bearer ${token}`);
    }
    return headers;
  },
});

function getRequestUrl(args: string | FetchArgs): string {
  return typeof args === "string" ? args : args.url;
}

function shouldAttemptReauth(args: string | FetchArgs): boolean {
  return !REAUTH_EXCLUDED_URLS.has(getRequestUrl(args));
}

function invalidateSession(api: BaseQueryApi, message = SESSION_EXPIRED_MESSAGE) {
  const state = api.getState() as RootState;

  if (state.auth.requiresReauth) {
    return;
  }

  api.dispatch(sessionExpired(message));
  api.dispatch(resetAllApiCaches());
}

async function refreshAccessToken(
  api: BaseQueryApi,
  extraOptions: object,
): Promise<boolean> {
  if (!refreshPromise) {
    refreshPromise = (async () => {
      const refreshResult = await rawBaseQuery(
        {url: "/auth/refresh", method: "POST"},
        api,
        extraOptions,
      );

      if (!refreshResult.data) {
        return false;
      }

      const {access_token} = refreshResult.data as RefreshResponse;

      if (!access_token) {
        return false;
      }

      api.dispatch(setCredentials(access_token));
      return true;
    })().finally(() => {
      refreshPromise = null;
    });
  }

  return refreshPromise;
}

export const baseQueryWithReauth: BaseQueryFn<string | FetchArgs, unknown, FetchBaseQueryError>
  = async (args, api, extraOptions) => {
  let result = await rawBaseQuery(args, api, extraOptions);

  if (result.error?.status === 401 && shouldAttemptReauth(args)) {
    const refreshed = await refreshAccessToken(api, extraOptions);

    if (refreshed) {
      result = await rawBaseQuery(args, api, extraOptions);

      if (result.error?.status === 401) {
        invalidateSession(api);
      }
    } else {
      invalidateSession(api);
    }
  }

  return result;
};
