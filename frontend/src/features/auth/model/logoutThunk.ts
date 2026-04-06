import {createAsyncThunk} from "@reduxjs/toolkit";
import {logout} from "@/features/auth/model/authSlice.ts";
import authApi from "@/features/auth/api/authApi.ts";
import dealsApi from "@/features/deals/api/dealsApi.ts";
import offersApi from "@/features/offers/api/offersApi.ts";
import usersApi from "@/features/users/api/usersApi.ts";

export const performLogout = createAsyncThunk(
  "auth/performLogout",
  async (_, {dispatch}) => {
    try {
      await dispatch(authApi.endpoints.logout.initiate()).unwrap();
    } catch (error) {
      console.warn("Logout failed:", error, "continuing anyway.");
    }

    dispatch(logout());
    dispatch(authApi.util.resetApiState());
    dispatch(dealsApi.util.resetApiState());
    dispatch(offersApi.util.resetApiState());
    dispatch(usersApi.util.resetApiState());
  }
)