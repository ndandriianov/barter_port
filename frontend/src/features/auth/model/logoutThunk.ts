import {createAsyncThunk} from "@reduxjs/toolkit";
import {logout} from "@/features/auth/model/authSlice.ts";
import authApi from "@/features/auth/api/authApi.ts";
import { resetAllApiCaches } from "@/app/store/resetCaches.ts";

export const performLogout = createAsyncThunk(
  "auth/performLogout",
  async (_, {dispatch}) => {
    try {
      await dispatch(authApi.endpoints.logout.initiate()).unwrap();
    } catch (error) {
      console.warn("Logout failed:", error, "continuing anyway.");
    }

    dispatch(logout());
    dispatch(resetAllApiCaches());
  }
)