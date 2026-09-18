import { createAsyncThunk, createSlice } from "@reduxjs/toolkit";

// TODO: route this through a Next.js route handler to the Go auth service
// (POST /auth/login, POST /auth/register) so the HttpOnly refresh cookie
// and CSRF pair stay intact. Direct browser calls would break on CORS.
export const loginWithPassword = createAsyncThunk(
  "auth/loginWithPassword",
  async ({ email }: { email: string; password: string }) => {
    await new Promise((resolve) => setTimeout(resolve, 600));
    return { email };
  },
);

interface AuthState {
  userEmail: string | null;
  status: "idle" | "loading" | "succeeded" | "failed";
  error: string | null;
}

const initialState: AuthState = {
  userEmail: null,
  status: "idle",
  error: null,
};

const authSlice = createSlice({
  name: "auth",
  initialState,
  reducers: {
    loggedOut(state) {
      state.userEmail = null;
      state.status = "idle";
      state.error = null;
    },
    authErrorCleared(state) {
      state.error = null;
      if (state.status === "failed") state.status = "idle";
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(loginWithPassword.pending, (state) => {
        state.status = "loading";
        state.error = null;
      })
      .addCase(loginWithPassword.fulfilled, (state, action) => {
        state.status = "succeeded";
        state.userEmail = action.payload.email;
      })
      .addCase(loginWithPassword.rejected, (state, action) => {
        state.status = "failed";
        state.error = action.error.message ?? "Couldn't log you in. Try again.";
      });
  },
});

export const { loggedOut, authErrorCleared } = authSlice.actions;
export default authSlice.reducer;
