import {createSlice, type PayloadAction} from "@reduxjs/toolkit";

interface AuthState {
  accessToken: string | null;
  requiresReauth: boolean;
  reauthMessage: string | null;
}

const initialState: AuthState = {
  accessToken: null,
  requiresReauth: false,
  reauthMessage: null,
};

const authSlice = createSlice({
  name: "auth",
  initialState,
  reducers: {
    setCredentials(state, action: PayloadAction<string>) {
      state.accessToken = action.payload;
      state.requiresReauth = false;
      state.reauthMessage = null;
    },
    logout(state) {
      state.accessToken = null;
      state.requiresReauth = false;
      state.reauthMessage = null;
    },
    sessionExpired(state, action: PayloadAction<string | undefined>) {
      state.accessToken = null;
      state.requiresReauth = true;
      state.reauthMessage = action.payload ?? "Сессия истекла. Войдите снова.";
    },
  },
});

export const { setCredentials, logout, sessionExpired } = authSlice.actions;
export default authSlice.reducer;
