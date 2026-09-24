import { createAsyncThunk, createSlice } from "@reduxjs/toolkit";
import type { RootState } from "@/lib/store";

export type AuthRole = "buyer" | "seller";

export interface AuthUser {
  id: string;
  email: string;
  role: AuthRole;
  created_at: string;
}

interface AuthSession {
  access_token: string;
  csrf_token: string;
  token_type: "Bearer";
  expires_in: number;
  user: AuthUser;
}

interface AuthState {
  user: AuthUser | null;
  accessToken: string | null;
  csrfToken: string | null;
  status: "idle" | "loading" | "authenticated" | "failed";
  error: string | null;
  initialized: boolean;
}

const initialState: AuthState = {
  user: null,
  accessToken: null,
  csrfToken: null,
  status: "idle",
  error: null,
  initialized: false,
};

function csrfFromCookie() {
  if (typeof document === "undefined") return null;
  const match = document.cookie
    .split("; ")
    .find((entry) => entry.startsWith("csrf_token="));
  return match ? decodeURIComponent(match.slice("csrf_token=".length)) : null;
}

async function authRequest<T>(url: string, init: RequestInit = {}): Promise<T> {
  const response = await fetch(url, {
    ...init,
    credentials: "same-origin",
    headers: {
      ...(init.body ? { "Content-Type": "application/json" } : {}),
      ...init.headers,
    },
  });

  if (!response.ok) {
    const body = (await response.json().catch(() => null)) as {
      error?: string;
    } | null;
    throw new Error(body?.error ?? "Authentication request failed.");
  }

  return response.status === 204
    ? (undefined as T)
    : ((await response.json()) as T);
}

function sessionPayload(credentials: { email: string; password: string }) {
  return JSON.stringify({
    email: credentials.email.trim(),
    password: credentials.password,
  });
}

export const loginWithPassword = createAsyncThunk<
  AuthSession,
  { email: string; password: string },
  { rejectValue: string }
>("auth/loginWithPassword", async (credentials, { rejectWithValue }) => {
  try {
    return await authRequest<AuthSession>("/api/auth/login", {
      method: "POST",
      body: sessionPayload(credentials),
    });
  } catch (error) {
    return rejectWithValue(
      error instanceof Error ? error.message : "Couldn't log you in. Try again.",
    );
  }
});

export const registerWithPassword = createAsyncThunk<
  AuthSession,
  { email: string; password: string },
  { rejectValue: string }
>("auth/registerWithPassword", async (credentials, { rejectWithValue }) => {
  try {
    return await authRequest<AuthSession>("/api/auth/register", {
      method: "POST",
      body: sessionPayload(credentials),
    });
  } catch (error) {
    return rejectWithValue(
      error instanceof Error ? error.message : "Couldn't create your account. Try again.",
    );
  }
});

export const refreshSession = createAsyncThunk<
  AuthSession,
  void,
  { state: RootState; rejectValue: string }
>("auth/refreshSession", async (_, { getState, rejectWithValue }) => {
  const csrfToken = getState().auth.csrfToken ?? csrfFromCookie();
  if (!csrfToken) return rejectWithValue("No active session.");

  try {
    return await authRequest<AuthSession>("/api/auth/refresh", {
      method: "POST",
      headers: { "X-CSRF-Token": csrfToken },
    });
  } catch (error) {
    return rejectWithValue(
      error instanceof Error ? error.message : "Couldn't restore your session.",
    );
  }
});

export const becomeSeller = createAsyncThunk<
  AuthSession,
  void,
  { state: RootState; rejectValue: string }
>("auth/becomeSeller", async (_, { getState, rejectWithValue }) => {
  const { accessToken, csrfToken } = getState().auth;
  if (!accessToken || !csrfToken) return rejectWithValue("Please log in first.");

  try {
    return await authRequest<AuthSession>("/api/auth/become-seller", {
      method: "POST",
      headers: {
        Authorization: `Bearer ${accessToken}`,
        "X-CSRF-Token": csrfToken,
      },
    });
  } catch (error) {
    return rejectWithValue(
      error instanceof Error ? error.message : "Couldn't activate seller access.",
    );
  }
});

export const logout = createAsyncThunk<void, void, { rejectValue: string }>(
  "auth/logout",
  async (_, { rejectWithValue }) => {
    try {
      await authRequest<void>("/api/auth/logout", { method: "POST" });
    } catch (error) {
      return rejectWithValue(
        error instanceof Error ? error.message : "Couldn't log out cleanly.",
      );
    }
  },
);

function applySession(state: AuthState, session: AuthSession) {
  state.user = session.user;
  state.accessToken = session.access_token;
  state.csrfToken = session.csrf_token;
  state.status = "authenticated";
  state.error = null;
  state.initialized = true;
}

function clearSession(state: AuthState) {
  state.user = null;
  state.accessToken = null;
  state.csrfToken = null;
  state.status = "idle";
  state.error = null;
  state.initialized = true;
}

const authSlice = createSlice({
  name: "auth",
  initialState,
  reducers: {
    authErrorCleared(state) {
      state.error = null;
      if (state.status === "failed") state.status = "idle";
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(refreshSession.rejected, (state) => clearSession(state))
      .addCase(logout.fulfilled, clearSession)
      .addCase(logout.rejected, clearSession)
      .addMatcher(
        (action) =>
          loginWithPassword.pending.match(action) ||
          registerWithPassword.pending.match(action) ||
          refreshSession.pending.match(action) ||
          becomeSeller.pending.match(action),
        (state) => {
          state.status = "loading";
          state.error = null;
        },
      )
      .addMatcher(
        (action) =>
          loginWithPassword.fulfilled.match(action) ||
          registerWithPassword.fulfilled.match(action) ||
          refreshSession.fulfilled.match(action) ||
          becomeSeller.fulfilled.match(action),
        (state, action) => applySession(state, action.payload),
      )
      .addMatcher(
        (action) =>
          loginWithPassword.rejected.match(action) ||
          registerWithPassword.rejected.match(action) ||
          becomeSeller.rejected.match(action),
        (state, action) => {
          state.status = "failed";
          state.error = action.payload ?? action.error.message ?? "Authentication failed.";
          state.initialized = true;
        },
      );
  },
});

export const { authErrorCleared } = authSlice.actions;
export default authSlice.reducer;
